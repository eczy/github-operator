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
	"sigs.k8s.io/controller-runtime/pkg/log"

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
	gh "github.com/eczy/github-operator/internal/github"
	"github.com/google/go-github/v60/github"
)

type OrganizationRequester interface {
	GetOrganization(ctx context.Context, org string) (*github.Organization, error)
	GetOrganizationByNodeId(ctx context.Context, nodeId string) (*github.Organization, error)
	UpdateOrganization(ctx context.Context, org string, updateOrg *github.Organization) (*github.Organization, error)
}

// OrganizationReconciler reconciles a Organization object
type OrganizationReconciler struct {
	client.Client
	Scheme                   *runtime.Scheme
	GitHubClient             OrganizationRequester
	DeleteOnResourceDeletion bool
	RequeueInterval          time.Duration
}

//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=organizations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=organizations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=organizations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Organization object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if r.GitHubClient == nil {
		return ctrl.Result{}, fmt.Errorf("nil GitHub client")
	}

	// fetch resource
	org := &githubv1alpha1.Organization{}
	if err := r.Get(ctx, req.NamespacedName, org); err != nil {
		log.Error(err, "error fetching Organization resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var observed *github.Organization
	// try to fetch external resource
	if org.Status.NodeId != nil {
		ghOrg, err := r.GitHubClient.GetOrganizationByNodeId(ctx, *org.Status.NodeId)
		if _, ok := err.(*gh.OrganizationNotFoundError); ok {
			log.Info(err.Error())
		} else if err != nil {
			log.Error(err, "error fetching GitHub organization")
			return ctrl.Result{}, err
		}
		observed = ghOrg
	} else {
		ghOrg, err := r.GitHubClient.GetOrganization(ctx, org.Spec.Login)
		if _, ok := err.(*gh.OrganizationNotFoundError); ok {
			log.Info(err.Error())
		} else if err != nil {
			log.Error(err, "error fetching GitHub organization")
			return ctrl.Result{}, err
		}
		observed = ghOrg
	}

	// if external resource does't exist and we aren't deleting the resource, return error (since we can't create organizations)
	if observed == nil && org.DeletionTimestamp.IsZero() {
		// can't create organizations, so return not found error
		return ctrl.Result{}, &gh.OrganizationNotFoundError{Login: &org.Spec.Login}
	}

	// update external resource
	err := r.updateOrganization(ctx, org, observed)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: r.RequeueInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&githubv1alpha1.Organization{}).
		Complete(r)
}

// updates both args in place
func (r *OrganizationReconciler) updateOrganization(ctx context.Context, organization *githubv1alpha1.Organization, ghOrganization *github.Organization) error {
	log := log.FromContext(ctx)

	updateOrg := github.Organization{}
	needsUpdate := false

	// login
	// if organization.Spec.Login != ghOrganization.GetLogin() {
	// }
	// name
	if ptrNonNilAndNotEqualTo(organization.Spec.Name, ghOrganization.GetName()) {
		log.Info("organization name update", "from", ghOrganization.GetName(), "to", organization.Spec.Name)
		updateOrg.Name = organization.Spec.Name
		needsUpdate = true
	}
	// billing email
	if ptrNonNilAndNotEqualTo(organization.Spec.BillingEmail, ghOrganization.GetBillingEmail()) {
		log.Info("organization billing email update", "from", ghOrganization.GetBillingEmail(), "to", organization.Spec.BillingEmail)
		updateOrg.BillingEmail = organization.Spec.BillingEmail
		needsUpdate = true
	}
	// company
	if ptrNonNilAndNotEqualTo(organization.Spec.Company, ghOrganization.GetCompany()) {
		log.Info("organization company update", "from", ghOrganization.GetCompany(), "to", organization.Spec.Company)
		updateOrg.Company = organization.Spec.Company
		needsUpdate = true
	}
	// email
	if ptrNonNilAndNotEqualTo(organization.Spec.Email, ghOrganization.GetEmail()) {
		log.Info("organization email update", "from", ghOrganization.GetEmail(), "to", organization.Spec.Email)
		updateOrg.Email = organization.Spec.Email
		needsUpdate = true
	}
	// twitter username
	if ptrNonNilAndNotEqualTo(organization.Spec.TwitterUsername, ghOrganization.GetTwitterUsername()) {
		log.Info("organization twitter username update", "from", ghOrganization.GetTwitterUsername(), "to", *organization.Spec.TwitterUsername)
		updateOrg.TwitterUsername = organization.Spec.TwitterUsername
		needsUpdate = true
	}
	// location
	if ptrNonNilAndNotEqualTo(organization.Spec.Location, ghOrganization.GetLocation()) {
		log.Info("organization location update", "from", ghOrganization.GetLocation(), "to", *organization.Spec.Location)
		updateOrg.Location = organization.Spec.Location
		needsUpdate = true
	}
	// description
	if ptrNonNilAndNotEqualTo(organization.Spec.Description, ghOrganization.GetDescription()) {
		log.Info("organization description update", "from", ghOrganization.GetDescription(), "to", *organization.Spec.Description)
		updateOrg.Description = organization.Spec.Description
		needsUpdate = true
	}
	// has organization projects
	if ptrNonNilAndNotEqualTo(organization.Spec.HasOrganizationProjects, ghOrganization.GetHasOrganizationProjects()) {
		log.Info("organization hasOrganizationProjects update", "from", ghOrganization.GetHasOrganizationProjects(), "to", *organization.Spec.HasOrganizationProjects)
		updateOrg.HasOrganizationProjects = organization.Spec.HasOrganizationProjects
		needsUpdate = true
	}
	// has repository projects
	if ptrNonNilAndNotEqualTo(organization.Spec.HasRepositoryProjects, ghOrganization.GetHasRepositoryProjects()) {
		log.Info("organization hasRepositoryProjects update", "from", ghOrganization.GetHasRepositoryProjects(), "to", *organization.Spec.HasRepositoryProjects)
		updateOrg.HasRepositoryProjects = organization.Spec.HasRepositoryProjects
		needsUpdate = true
	}
	// default repository permission
	if ptrNonNilAndNotEqualTo(organization.Spec.DefaultRepositoryPermission, (githubv1alpha1.DefaultRepositoryPermission)(ghOrganization.GetDefaultRepoPermission())) {
		log.Info("organization defaultRepositoryPermission update", "from", ghOrganization.GetDefaultRepoPermission(), "to", *organization.Spec.DefaultRepositoryPermission)
		updateOrg.DefaultRepoPermission = (*string)(organization.Spec.DefaultRepositoryPermission)
		needsUpdate = true
	}
	// members can create repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreateRepositories, ghOrganization.GetMembersCanCreateRepos()) {
		log.Info("organization membersCanCreateRepositories update", "from", ghOrganization.GetMembersCanCreateRepos(), "to", *organization.Spec.MembersCanCreateRepositories)
		updateOrg.MembersCanCreateRepos = organization.Spec.MembersCanCreateRepositories
		needsUpdate = true
	}
	// members can create internal repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreateInternalRepositories, ghOrganization.GetMembersCanCreateInternalRepos()) {
		log.Info("organization membersCanCreateInternalRepositories update", "from", ghOrganization.GetMembersCanCreateInternalRepos(), "to", *organization.Spec.MembersCanCreateInternalRepositories)
		updateOrg.MembersCanCreateInternalRepos = organization.Spec.MembersCanCreateInternalRepositories
		needsUpdate = true
	}
	// members can create private repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePrivateRepositories, ghOrganization.GetMembersCanCreatePrivateRepos()) {
		log.Info("organization membersCanCreatePrivateRepositories update", "from", ghOrganization.GetMembersCanCreatePrivateRepos(), "to", *organization.Spec.MembersCanCreatePrivateRepositories)
		updateOrg.MembersCanCreatePrivateRepos = organization.Spec.MembersCanCreatePrivateRepositories
		needsUpdate = true
	}
	// members can create public repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePublicRepositories, ghOrganization.GetMembersCanCreatePublicRepos()) {
		log.Info("organization membersCanCreatePublicRepositories update", "from", ghOrganization.GetMembersCanCreatePublicRepos(), "to", *organization.Spec.MembersCanCreatePublicRepositories)
		updateOrg.MembersCanCreatePublicRepos = organization.Spec.MembersCanCreatePublicRepositories
		needsUpdate = true
	}
	// members can create pages
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePages, ghOrganization.GetMembersCanCreatePages()) {
		log.Info("organization membersCanCreatePages update", "from", ghOrganization.GetMembersCanCreatePages(), "to", *organization.Spec.MembersCanCreatePages)
		updateOrg.MembersCanCreatePages = organization.Spec.MembersCanCreatePages
		needsUpdate = true
	}
	// members can create public pages
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePublicPages, ghOrganization.GetMembersCanCreatePublicPages()) {
		log.Info("organization membersCanCreatePublicPages update", "from", ghOrganization.GetMembersCanCreatePublicPages(), "to", *organization.Spec.MembersCanCreatePublicPages)
		updateOrg.MembersCanCreatePublicPages = organization.Spec.MembersCanCreatePublicPages
		needsUpdate = true
	}
	// members can create private pages
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePrivatePages, ghOrganization.GetMembersCanCreatePrivatePages()) {
		log.Info("organization membersCanCreatePrivatePages update", "from", ghOrganization.GetMembersCanCreatePrivatePages(), "to", *organization.Spec.MembersCanCreatePrivatePages)
		updateOrg.MembersCanCreatePrivatePages = organization.Spec.MembersCanCreatePrivatePages
		needsUpdate = true
	}
	// members can fork private repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanForkPrivateRepositories, ghOrganization.GetMembersCanForkPrivateRepos()) {
		log.Info("organization membersCanForkPrivateRepositories update", "from", ghOrganization.GetMembersCanForkPrivateRepos(), "to", *organization.Spec.MembersCanForkPrivateRepositories)
		updateOrg.MembersCanForkPrivateRepos = organization.Spec.MembersCanForkPrivateRepositories
		needsUpdate = true
	}
	// web commit signoff required
	if ptrNonNilAndNotEqualTo(organization.Spec.WebCommitSignoffRequired, ghOrganization.GetWebCommitSignoffRequired()) {
		log.Info("organization webCommitSignoffRequired update", "from", ghOrganization.GetWebCommitSignoffRequired(), "to", *organization.Spec.WebCommitSignoffRequired)
		updateOrg.WebCommitSignoffRequired = organization.Spec.WebCommitSignoffRequired
		needsUpdate = true
	}
	// blog
	if ptrNonNilAndNotEqualTo(organization.Spec.Blog, ghOrganization.GetBlog()) {
		log.Info("organization blog update", "from", ghOrganization.GetBlog(), "to", *organization.Spec.Blog)
		updateOrg.Blog = organization.Spec.Blog
		needsUpdate = true
	}
	// advanced security enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.AdvancedSecurityEnabledForNewRepositories, ghOrganization.GetAdvancedSecurityEnabledForNewRepos()) {
		log.Info("organization advancedSecurityEnabledForNewRepositories update", "from", ghOrganization.GetAdvancedSecurityEnabledForNewRepos(), "to", *organization.Spec.AdvancedSecurityEnabledForNewRepositories)
		updateOrg.AdvancedSecurityEnabledForNewRepos = organization.Spec.AdvancedSecurityEnabledForNewRepositories
		needsUpdate = true
	}
	// dependabot alerts enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.DependabotAlertsEnabledForNewRepositories, ghOrganization.GetAdvancedSecurityEnabledForNewRepos()) {
		log.Info("organization dependabotAlertsEnabledForNewRepositories update", "from", ghOrganization.GetAdvancedSecurityEnabledForNewRepos(), "to", *organization.Spec.DependabotAlertsEnabledForNewRepositories)
		updateOrg.DependabotAlertsEnabledForNewRepos = organization.Spec.DependabotAlertsEnabledForNewRepositories
		needsUpdate = true
	}
	// dependabot security updates enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.DependabotSecurityUpdatesEnabledForNewRepositories, ghOrganization.GetDependabotSecurityUpdatesEnabledForNewRepos()) {
		log.Info("organization dependabotSecurityUpdatesEnabledForNewRepositories update", "from", ghOrganization.GetDependabotSecurityUpdatesEnabledForNewRepos(), "to", *organization.Spec.DependabotAlertsEnabledForNewRepositories)
		updateOrg.DependabotSecurityUpdatesEnabledForNewRepos = organization.Spec.DependabotSecurityUpdatesEnabledForNewRepositories
		needsUpdate = true
	}
	// dependency graph enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.DependencyGraphEnabledForNewRepositories, ghOrganization.GetDependencyGraphEnabledForNewRepos()) {
		log.Info("organization dependencyGraphEnabledForNewRepositories update", "from", ghOrganization.GetDependencyGraphEnabledForNewRepos(), "to", *organization.Spec.DependencyGraphEnabledForNewRepositories)
		updateOrg.DependencyGraphEnabledForNewRepos = organization.Spec.DependencyGraphEnabledForNewRepositories
		needsUpdate = true
	}
	// secret scanning enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.SecretScanningEnabledForNewRepositories, ghOrganization.GetSecretScanningEnabledForNewRepos()) {
		log.Info("organization secretScanningEnabledForNewRepositories update", "from", ghOrganization.GetSecretScanningEnabledForNewRepos(), "to", *organization.Spec.SecretScanningEnabledForNewRepositories)
		updateOrg.SecretScanningEnabledForNewRepos = organization.Spec.SecretScanningEnabledForNewRepositories
		needsUpdate = true
	}

	// perform update if necessary
	if needsUpdate || organization.Status.LastUpdateTimestamp == nil {
		log.Info("updating organization", "login", organization.Spec.Login)
		updated, err := r.GitHubClient.UpdateOrganization(ctx, *ghOrganization.Login, &updateOrg)
		if err != nil {
			log.Error(err, "unable to update organization", "login", organization.Spec.Login)
			return err
		}
		ghOrganization = updated

		now := v1.Now()
		organization.Status = githubv1alpha1.OrganizationStatus{
			Login:                                updated.Login,
			NodeId:                               updated.NodeID,
			LastUpdateTimestamp:                  &now,
			Name:                                 updated.GetName(),
			BillingEmail:                         updated.GetBillingEmail(),
			Company:                              updated.GetCompany(),
			Email:                                updated.GetEmail(),
			TwitterUsername:                      updated.TwitterUsername,
			Location:                             updated.Location,
			Description:                          updated.Description,
			HasOrganizationProjects:              updated.HasOrganizationProjects,
			HasRepositoryProjects:                updated.HasRepositoryProjects,
			DefaultRepositoryPermission:          (*githubv1alpha1.DefaultRepositoryPermission)(updated.DefaultRepoPermission),
			MembersCanCreateRepositories:         updated.MembersCanCreateRepos,
			MembersCanCreateInternalRepositories: updated.MembersCanCreateInternalRepos,
			MembersCanCreatePrivateRepositories:  updated.MembersCanCreatePrivateRepos,
			MembersCanCreatePublicRepositories:   updated.MembersCanCreatePublicRepos,
			MembersCanCreatePages:                updated.MembersCanCreatePages,
			MembersCanCreatePublicPages:          updated.MembersCanCreatePublicPages,
			MembersCanCreatePrivatePages:         updated.MembersCanCreatePrivatePages,
			MembersCanForkPrivateRepositories:    updated.MembersCanForkPrivateRepos,
			WebCommitSignoffRequired:             updated.WebCommitSignoffRequired,
			Blog:                                 updated.Blog,
			AdvancedSecurityEnabledForNewRepositories:             updated.AdvancedSecurityEnabledForNewRepos,
			DependabotAlertsEnabledForNewRepositories:             updated.DependabotAlertsEnabledForNewRepos,
			DependabotSecurityUpdatesEnabledForNewRepositories:    updated.DependabotSecurityUpdatesEnabledForNewRepos,
			DependencyGraphEnabledForNewRepositories:              updated.DependencyGraphEnabledForNewRepos,
			SecretScanningEnabledForNewRepositories:               updated.SecretScanningEnabledForNewRepos,
			SecretScanningPushProtectionEnabledForNewRepositories: updated.SecretScanningPushProtectionEnabledForNewRepos,
		}

		// update status
		if err := r.Status().Update(ctx, organization); err != nil {
			log.Error(err, "unable to update Organization status", "name", organization.Spec.Name)
		}
	}

	return nil
}

// func (r *OrganizationReconciler) deleteOrganization(ctx context.Context, organization *githubv1alpha1.Organization) error {
// 	if organization.Status.Login == nil {
// 		return fmt.Errorf("organization login is nil")
// 	}
// 	return r.GitHubClient.DeleteOrganization(ctx, *organization.Status.Login)
// }
