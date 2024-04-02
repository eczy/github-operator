/*
Copyright 2024 Evan Czyzycki

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
	gh "github.com/eczy/github-operator/internal/github"
	"github.com/google/go-github/v60/github"
)

var (
	repositoryFinalizerName = "github.github-operator.eczy.io/repo-finalizer"
)

type RepositoryRequester interface {
	RepositoryGetter

	UpdateRepositoryByName(ctx context.Context, owner, name string, update *github.Repository) (*github.Repository, error)
	CreateRepository(ctx context.Context, org string, create *github.Repository) (*github.Repository, error)
	CreateRepositoryFromTemplate(ctx context.Context, templateOwner string, templateRepository string, req *github.TemplateRepoRequest) (*github.Repository, error)
	DeleteRepositoryByName(ctx context.Context, owner, name string) error
	UpdateRepositoryTopics(ctx context.Context, owner string, repo string, topics []string) ([]string, error)
}

type RepositoryGetter interface {
	GetRepositoryByName(ctx context.Context, owner string, name string) (*github.Repository, error)
	GetRepositoryByNodeId(ctx context.Context, nodeId string) (*github.Repository, error)
}

// RepositoryReconciler reconciles a Repository object
type RepositoryReconciler struct {
	client.Client
	Scheme                   *runtime.Scheme
	GitHubClient             RepositoryRequester
	DeleteOnResourceDeletion bool
	RequeueInterval          time.Duration
}

//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=repositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=repositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=repositories/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Repository object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if r.GitHubClient == nil {
		return ctrl.Result{}, fmt.Errorf("nil GitHub client")
	}

	// fetch resource
	repo := &githubv1alpha1.Repository{}
	if err := r.Get(ctx, req.NamespacedName, repo); err != nil {
		log.Error(err, "error fetching Repository resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var observed *github.Repository
	// try to fetch external resource
	if repo.Status.NodeId != nil {
		ghTeam, err := r.GitHubClient.GetRepositoryByNodeId(ctx, *repo.Status.NodeId)
		if _, ok := err.(*gh.RepositoryNotFoundError); ok {
			log.Info(err.Error())
		} else if err != nil {
			log.Error(err, "error fetching GitHub repository")
			return ctrl.Result{}, err
		}
		observed = ghTeam
	} else {
		ghRepo, err := r.GitHubClient.GetRepositoryByName(ctx, repo.Spec.Owner, repo.Spec.Name)
		if _, ok := err.(*gh.RepositoryNotFoundError); ok {
			log.Info(err.Error())
		} else if err != nil {
			log.Error(err, "error fetching GitHub repository")
			return ctrl.Result{}, err
		}
		observed = ghRepo
	}

	// if external resource does't exist and we aren't deleting the resource, create external resource
	if observed == nil && repo.ObjectMeta.DeletionTimestamp.IsZero() {
		ghRepo, err := r.createRepository(ctx, repo)
		if err != nil {
			log.Error(err, "error creating GitHub repository")
			return ctrl.Result{}, err
		}
		observed = ghRepo
	}

	// handle finalizer
	if r.DeleteOnResourceDeletion {
		if repo.ObjectMeta.DeletionTimestamp.IsZero() {
			// not being deleted
			if !controllerutil.ContainsFinalizer(repo, repositoryFinalizerName) {
				controllerutil.AddFinalizer(repo, repositoryFinalizerName)
				if err := r.Update(ctx, repo); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else {
			// being deleted
			if repo.Status.NodeId != nil {
				// if we have never resolved this resource before, don't
				// touch external state
				if err := r.deleteRepository(ctx, repo); err != nil {
					log.Error(err, "error deleting repository")
					return ctrl.Result{}, err
				}
			}

			controllerutil.RemoveFinalizer(repo, repositoryFinalizerName)
			if err := r.Update(ctx, repo); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	}

	// update external resource
	err := r.updateRepository(ctx, repo, observed)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: r.RequeueInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&githubv1alpha1.Repository{}).
		Complete(r)
}

func (r *RepositoryReconciler) createRepository(ctx context.Context, repo *githubv1alpha1.Repository) (*github.Repository, error) {
	if repo.Spec.TemplateRepository != nil && repo.Spec.TemplateOwner != nil {
		repository, err := r.GitHubClient.CreateRepositoryFromTemplate(ctx, *repo.Spec.TemplateOwner, *repo.Spec.TemplateRepository, &github.TemplateRepoRequest{
			Name:               github.String(repo.Spec.Name),
			Owner:              github.String(repo.Spec.Owner),
			Description:        repo.Spec.Description,
			IncludeAllBranches: github.Bool(false),
			Private:            github.Bool(repo.Spec.Visibility != nil && *repo.Spec.Visibility != "private"),
		})
		if err != nil {
			return nil, err
		}
		return repository, err
	}
	newRepo := repositoryToGitHubRepository(repo)
	repository, err := r.GitHubClient.CreateRepository(ctx, repo.Spec.Owner, newRepo)
	if err != nil {
		return nil, err
	}
	return repository, nil
}

// TODO: this needs refactoring, but ignore for now so that CI passes
//
//gocyclo:ignore
func (r *RepositoryReconciler) updateRepository(ctx context.Context, repo *githubv1alpha1.Repository, ghRepo *github.Repository) error {
	log := log.FromContext(ctx)

	updateRepo := &github.Repository{}
	needsUpdate := false
	needsTopicsUpdate := false

	// Name
	if repo.Spec.Name != ghRepo.GetName() {
		log.Info("repository name update", "from", ghRepo.GetName(), "to", repo.Spec.Name)
		updateRepo.Name = &repo.Spec.Name
		needsUpdate = true
	}
	// Owner
	owner := ghRepo.GetOwner()
	if owner != nil && repo.Spec.Owner != owner.GetLogin() {
		return fmt.Errorf("repository owner '%s' does not match Spec owner '%s'", repo.Spec.Name, owner.GetLogin())
	}
	// TODO: transfer repository to new owner

	// Description
	if ptrNonNilAndNotEqualTo(repo.Spec.Description, ghRepo.GetDescription()) {
		log.Info("repository Description update", "from", ghRepo.GetDescription(), "to", repo.Spec.Description)
		updateRepo.Description = repo.Spec.Description
		needsUpdate = true
	}
	// Homepage
	if ptrNonNilAndNotEqualTo(repo.Spec.Homepage, ghRepo.GetDescription()) {
		log.Info("repository Homepage update", "from", ghRepo.GetHomepage(), "to", repo.Spec.Homepage)
		updateRepo.Homepage = repo.Spec.Homepage
		needsUpdate = true
	}
	// DefaultBranch
	if ptrNonNilAndNotEqualTo(repo.Spec.DefaultBranch, ghRepo.GetDefaultBranch()) {
		log.Info("repository DefaultBranch update", "from", ghRepo.GetDefaultBranch(), "to", repo.Spec.Description)
		updateRepo.DefaultBranch = repo.Spec.DefaultBranch
		needsUpdate = true
	}
	// AllowRebaseMerge
	if ptrNonNilAndNotEqualTo(repo.Spec.AllowRebaseMerge, ghRepo.GetAllowRebaseMerge()) {
		log.Info("repository AllowRebaseMerge update", "from", ghRepo.GetDescription(), "to", repo.Spec.AllowRebaseMerge)
		updateRepo.AllowRebaseMerge = repo.Spec.AllowRebaseMerge
		needsUpdate = true
	}
	// AllowUpdateBranch
	if ptrNonNilAndNotEqualTo(repo.Spec.AllowUpdateBranch, ghRepo.GetAllowUpdateBranch()) {
		log.Info("repository AllowUpdateBranch update", "from", ghRepo.GetAllowUpdateBranch(), "to", repo.Spec.AllowUpdateBranch)
		updateRepo.AllowUpdateBranch = repo.Spec.AllowUpdateBranch
		needsUpdate = true
	}
	// AllowSquashMerge
	if ptrNonNilAndNotEqualTo(repo.Spec.AllowSquashMerge, ghRepo.GetAllowSquashMerge()) {
		log.Info("repository AllowSquashMerge update", "from", ghRepo.GetAllowSquashMerge(), "to", repo.Spec.AllowSquashMerge)
		updateRepo.AllowSquashMerge = repo.Spec.AllowSquashMerge
		needsUpdate = true
	}
	// AllowMergeCommit
	if ptrNonNilAndNotEqualTo(repo.Spec.AllowMergeCommit, ghRepo.GetAllowMergeCommit()) {
		log.Info("repository AllowMergeCommit update", "from", ghRepo.GetAllowMergeCommit(), "to", repo.Spec.AllowMergeCommit)
		updateRepo.AllowMergeCommit = repo.Spec.AllowMergeCommit
		needsUpdate = true
	}
	// AllowAutoMerge
	if ptrNonNilAndNotEqualTo(repo.Spec.AllowAutoMerge, ghRepo.GetAllowAutoMerge()) {
		log.Info("repository AllowAutoMerge update", "from", ghRepo.GetAllowAutoMerge(), "to", repo.Spec.AllowAutoMerge)
		updateRepo.AllowAutoMerge = repo.Spec.AllowAutoMerge
		needsUpdate = true
	}
	// AllowForking
	if ptrNonNilAndNotEqualTo(repo.Spec.AllowForking, ghRepo.GetAllowForking()) {
		log.Info("repository AllowForking update", "from", ghRepo.GetAllowForking(), "to", repo.Spec.AllowForking)
		updateRepo.AllowForking = repo.Spec.AllowForking
		needsUpdate = true
	}
	// WebCommitSignoffRequired
	if ptrNonNilAndNotEqualTo(repo.Spec.WebCommitSignoffRequired, ghRepo.GetWebCommitSignoffRequired()) {
		log.Info("repository WebCommitSignoffRequired update", "from", ghRepo.GetWebCommitSignoffRequired(), "to", repo.Spec.WebCommitSignoffRequired)
		updateRepo.WebCommitSignoffRequired = repo.Spec.WebCommitSignoffRequired
		needsUpdate = true
	}
	// DeleteBranchOnMerge
	if ptrNonNilAndNotEqualTo(repo.Spec.DeleteBranchOnMerge, ghRepo.GetDeleteBranchOnMerge()) {
		log.Info("repository DeleteBranchOnMerge update", "from", ghRepo.GetDeleteBranchOnMerge(), "to", repo.Spec.DeleteBranchOnMerge)
		updateRepo.DeleteBranchOnMerge = repo.Spec.DeleteBranchOnMerge
		needsUpdate = true
	}
	// SquashMergeCommitTitle
	if ptrNonNilAndNotEqualTo(repo.Spec.SquashMergeCommitTitle, (githubv1alpha1.SquashMergeCommitTitle)(ghRepo.GetSquashMergeCommitTitle())) {
		log.Info("repository SquashMergeCommitTitle update", "from", ghRepo.GetSquashMergeCommitTitle(), "to", repo.Spec.SquashMergeCommitTitle)
		updateRepo.SquashMergeCommitTitle = (*string)(repo.Spec.SquashMergeCommitTitle)
		needsUpdate = true
	}
	// SquashMergeCommitMessage
	if ptrNonNilAndNotEqualTo(repo.Spec.SquashMergeCommitMessage, (githubv1alpha1.SquashMergeCommitMessage)(ghRepo.GetSquashMergeCommitMessage())) {
		log.Info("repository SquashMergeCommitMessage update", "from", ghRepo.GetSquashMergeCommitMessage(), "to", repo.Spec.SquashMergeCommitMessage)
		updateRepo.SquashMergeCommitMessage = (*string)(repo.Spec.SquashMergeCommitMessage)
		needsUpdate = true
	}
	// MergeCommitTitle
	if ptrNonNilAndNotEqualTo(repo.Spec.MergeCommitTitle, (githubv1alpha1.MergeCommitTitle)(ghRepo.GetMergeCommitTitle())) {
		log.Info("repository MergeCommitTitle update", "from", ghRepo.GetMergeCommitTitle(), "to", repo.Spec.MergeCommitTitle)
		updateRepo.MergeCommitTitle = (*string)(repo.Spec.MergeCommitTitle)
		needsUpdate = true
	}
	// MergeCommitMessage
	if ptrNonNilAndNotEqualTo(repo.Spec.MergeCommitMessage, (githubv1alpha1.MergeCommitMessage)(ghRepo.GetMergeCommitMessage())) {
		log.Info("repository MergeCommitMessage update", "from", ghRepo.GetMergeCommitMessage(), "to", repo.Spec.MergeCommitMessage)
		updateRepo.MergeCommitMessage = (*string)(repo.Spec.MergeCommitMessage)
		needsUpdate = true
	}
	// Topics
	if !cmpSlices(repo.Spec.Topics, ghRepo.Topics) {
		log.Info("repository Topics update", "from", ghRepo.Topics, "to", repo.Spec.Topics)
		updateRepo.Topics = repo.Spec.Topics
		needsTopicsUpdate = true
	}
	// Archived
	if ptrNonNilAndNotEqualTo(repo.Spec.Archived, ghRepo.GetArchived()) {
		log.Info("repository Archived update", "from", ghRepo.GetArchived(), "to", repo.Spec.Archived)
		updateRepo.Archived = repo.Spec.Archived
		needsUpdate = true
	}
	// HasIssues
	if ptrNonNilAndNotEqualTo(repo.Spec.HasIssues, ghRepo.GetHasIssues()) {
		log.Info("repository HasIssues update", "from", ghRepo.GetHasIssues(), "to", repo.Spec.HasIssues)
		updateRepo.HasIssues = repo.Spec.HasIssues
		needsUpdate = true
	}
	// HasWiki
	if ptrNonNilAndNotEqualTo(repo.Spec.HasWiki, ghRepo.GetHasWiki()) {
		log.Info("repository HasWiki update", "from", ghRepo.GetHasWiki(), "to", repo.Spec.HasWiki)
		updateRepo.HasWiki = repo.Spec.HasWiki
		needsUpdate = true
	}
	// HasProjects
	if ptrNonNilAndNotEqualTo(repo.Spec.HasProjects, ghRepo.GetHasProjects()) {
		log.Info("repository HasProjects update", "from", ghRepo.GetHasProjects(), "to", repo.Spec.HasProjects)
		updateRepo.HasProjects = repo.Spec.HasProjects
		needsUpdate = true
	}
	// HasDownloads
	if ptrNonNilAndNotEqualTo(repo.Spec.HasDownloads, ghRepo.GetHasDownloads()) {
		log.Info("repository HasDownloads update", "from", ghRepo.GetHasDownloads(), "to", repo.Spec.HasDownloads)
		updateRepo.HasDownloads = repo.Spec.HasDownloads
		needsUpdate = true
	}
	// HasDiscussions
	if ptrNonNilAndNotEqualTo(repo.Spec.HasDiscussions, ghRepo.GetHasDiscussions()) {
		log.Info("repository HasDiscussions update", "from", ghRepo.GetHasDiscussions(), "to", repo.Spec.HasDiscussions)
		updateRepo.HasDownloads = repo.Spec.HasDownloads
		needsUpdate = true
	}
	// Visibility
	if ptrNonNilAndNotEqualTo(repo.Spec.Visibility, ghRepo.GetVisibility()) {
		log.Info("repository Visibility update", "from", ghRepo.GetVisibility(), "to", repo.Spec.Visibility)
		updateRepo.Visibility = repo.Spec.Visibility
		needsUpdate = true
	}

	// perform update if necessary
	// TODO: more granular updates (allow just topic update)
	if needsUpdate || needsTopicsUpdate || repo.Status.LastUpdateTimestamp == nil {
		log.Info("updating repository", "name", ghRepo.GetName())
		updated, err := r.GitHubClient.UpdateRepositoryByName(ctx, ghRepo.GetOwner().GetLogin(), ghRepo.GetName(), updateRepo)
		if err != nil {
			log.Error(err, "error updating repository", "name", repo.Spec.Name)
			return err
		}

		ghRepoTopics, err := r.GitHubClient.UpdateRepositoryTopics(ctx, repo.Spec.Owner, repo.Spec.Name, repo.Spec.Topics)
		if err != nil {
			log.Error(err, "error updating repository topics", "name", repo.Spec.Name)
		}
		ghRepo.Topics = ghRepoTopics

		ghRepo = updated

		now := v1.Now()

		owner := ghRepo.GetOwner()
		var ownerLogin string
		var ownerId int64
		if owner != nil {
			ownerLogin = owner.GetLogin()
			ownerId = owner.GetID()
		}

		parent := ghRepo.GetParent()
		var parentName string
		var parentId int64
		if owner != nil {
			parentName = parent.GetName()
			parentId = parent.GetID()
		}

		templateRepository := ghRepo.GetTemplateRepository()
		var templateRepositoryOwnerLogin string
		var templateRepositoryName string
		var templateRepositoryId int64
		if templateRepository != nil {
			templateRepositoryOwnerLogin = templateRepository.Owner.GetName()
			templateRepositoryName = templateRepository.GetName()
			templateRepositoryId = templateRepository.GetID()
		}

		organization := ghRepo.GetOrganization()
		var organizationLogin string
		var organizationId int64
		if organization != nil {
			organizationLogin = organization.GetLogin()
			organizationId = organization.GetID()
		}

		repo.Status = githubv1alpha1.RepositoryStatus{
			LastUpdateTimestamp:          &now,
			Id:                           ghRepo.ID,
			NodeId:                       ghRepo.NodeID,
			OwnerLogin:                   github.String(ownerLogin),
			OwnerNodeId:                  github.Int64(ownerId),
			Name:                         ghRepo.Name,
			FullName:                     ghRepo.FullName,
			Description:                  ghRepo.Description,
			Homepage:                     ghRepo.Homepage,
			DefaultBranch:                ghRepo.DefaultBranch,
			CreatedAt:                    (*v1.Time)(ghRepo.CreatedAt),
			PushedAt:                     (*v1.Time)(ghRepo.PushedAt),
			UpdatedAt:                    (*v1.Time)(ghRepo.UpdatedAt),
			ParentName:                   github.String(parentName),
			ParentId:                     github.Int64(parentId),
			TemplateRepositoryOwnerLogin: github.String(templateRepositoryOwnerLogin),
			TemplateRepositoryName:       github.String(templateRepositoryName),
			TemplateRepositoryId:         github.Int64(templateRepositoryId),
			OrganizationLogin:            github.String(organizationLogin),
			OrganizationId:               github.Int64(organizationId),
			AllowRebaseMerge:             ghRepo.AllowRebaseMerge,
			AllowUpdateBranch:            ghRepo.AllowUpdateBranch,
			AllowSquashMerge:             ghRepo.AllowSquashMerge,
			AllowMergeCommit:             ghRepo.AllowMergeCommit,
			AllowAutoMerge:               ghRepo.AllowAutoMerge,
			AllowForking:                 ghRepo.AllowForking,
			WebCommitSignoffRequired:     ghRepo.WebCommitSignoffRequired,
			DeleteBranchOnMerge:          ghRepo.DeleteBranchOnMerge,
			SquashMergeCommitTitle:       (*githubv1alpha1.SquashMergeCommitTitle)(ghRepo.SquashMergeCommitTitle),
			SquashMergeCommitMessage:     (*githubv1alpha1.SquashMergeCommitMessage)(ghRepo.SquashMergeCommitMessage),
			MergeCommitTitle:             (*githubv1alpha1.MergeCommitTitle)(ghRepo.MergeCommitTitle),
			MergeCommitMessage:           (*githubv1alpha1.MergeCommitMessage)(ghRepo.MergeCommitMessage),
			Topics:                       ghRepo.Topics,
			Archived:                     ghRepo.Archived,
			HasIssues:                    ghRepo.HasIssues,
			HasWiki:                      ghRepo.HasWiki,
			HasProjects:                  ghRepo.HasProjects,
			HasDownloads:                 ghRepo.HasDownloads,
			HasDiscussions:               ghRepo.HasDiscussions,
			Visibility:                   ghRepo.Visibility,
		}

		// update status
		if err := r.Status().Update(ctx, repo); err != nil {
			log.Error(err, "error updating Repository status", "name", repo.Spec.Name)
		}
	}
	return nil
}

func (r *RepositoryReconciler) deleteRepository(ctx context.Context, repo *githubv1alpha1.Repository) error {
	if repo.Status.OwnerLogin == nil {
		return fmt.Errorf("repo OwnerLogin is nil")
	} else if repo.Status.Name == nil {
		return fmt.Errorf("repo Name is nil")
	}
	return r.GitHubClient.DeleteRepositoryByName(ctx, *repo.Status.OwnerLogin, *repo.Status.Name)
}

func repositoryToGitHubRepository(repository *githubv1alpha1.Repository) *github.Repository {
	ghRepo := &github.Repository{
		Name:                     github.String(repository.Spec.Name),
		Description:              repository.Spec.Description,
		Homepage:                 repository.Spec.Homepage,
		DefaultBranch:            repository.Spec.DefaultBranch,
		AllowRebaseMerge:         repository.Spec.AllowRebaseMerge,
		AllowUpdateBranch:        repository.Spec.AllowUpdateBranch,
		AllowSquashMerge:         repository.Spec.AllowSquashMerge,
		AllowMergeCommit:         repository.Spec.AllowMergeCommit,
		AllowAutoMerge:           repository.Spec.AllowAutoMerge,
		AllowForking:             repository.Spec.AllowForking,
		WebCommitSignoffRequired: repository.Spec.WebCommitSignoffRequired,
		DeleteBranchOnMerge:      repository.Spec.DeleteBranchOnMerge,
		SquashMergeCommitTitle:   (*string)(repository.Spec.SquashMergeCommitTitle),
		SquashMergeCommitMessage: (*string)(repository.Spec.SquashMergeCommitMessage),
		MergeCommitTitle:         (*string)(repository.Spec.MergeCommitTitle),
		MergeCommitMessage:       (*string)(repository.Spec.MergeCommitMessage),
		Topics:                   repository.Spec.Topics,
		Archived:                 repository.Spec.Archived,
		HasIssues:                repository.Spec.HasIssues,
		HasWiki:                  repository.Spec.HasWiki,
		HasProjects:              repository.Spec.HasProjects,
		HasDownloads:             repository.Spec.HasDownloads,
		HasDiscussions:           repository.Spec.HasDiscussions,
		Visibility:               repository.Spec.Visibility,
	}
	return ghRepo
}
