/*
Copyright 2024.

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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
	gh "github.com/eczy/github-operator/internal/github"
	"github.com/google/go-github/v60/github"
)

// var (
// 	organizationFinalizerName = "github.github-operator.eczy.io/organization-finalizer"
// )

type OrganizationRequester interface {
	GetOrganization(ctx context.Context, org string) (*github.Organization, error)
	UpdateOrganization(ctx context.Context, org string, updateOrg *github.Organization) (*github.Organization, error)
	// DeleteOrganization(ctx context.Context, org string) error
}

// OrganizationReconciler reconciles a Organization object
type OrganizationReconciler struct {
	client.Client
	Scheme                   *runtime.Scheme
	GitHubClient             OrganizationRequester
	DeleteOnResourceDeletion bool
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
		err := fmt.Errorf("nil GitHub client")
		log.Error(err, "GitHub client is nil")
		return ctrl.Result{}, err
	}

	// fetch organization resource
	org := &githubv1alpha1.Organization{}
	if err := r.Get(ctx, req.NamespacedName, org); err != nil {
		log.Error(err, "error fetching Organization resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var observed *github.Organization
	// try to fetch external resource
	if org.Status.Login != nil {
		log.Info("getting organization", "login", *org.Status.Login)
		ghOrg, err := r.GitHubClient.GetOrganization(ctx, *org.Status.Login)
		if _, ok := err.(*gh.OrganizationNotFoundError); ok {
			log.Info(err.Error())
		} else if err != nil {
			log.Error(err, "unable to get organization")
			return ctrl.Result{}, err
		}
		observed = ghOrg
	} else {
		log.Info("getting organization", "login", org.Spec.Login)
		ghOrg, err := r.GitHubClient.GetOrganization(ctx, org.Spec.Login)
		if _, ok := err.(*gh.OrganizationNotFoundError); ok {
			log.Info(err.Error())
		} else if err != nil {
			log.Error(err, "unable to get organization")
			return ctrl.Result{}, err
		}
		observed = ghOrg
	}

	// if organization does't exist, check if scheduled for deletion
	if observed == nil {
		// if scheduled for deletion
		if !org.ObjectMeta.DeletionTimestamp.IsZero() {
			// do nothing and return since the external resource doesn't exist
			return ctrl.Result{}, nil
		}
		// can't create organizations, so return not found error
		return ctrl.Result{}, &gh.OrganizationNotFoundError{Login: &org.Spec.Login}
	}

	// NOTE: Keep in case organization deletion is supported in the future
	// handle finalizer
	// if r.DeleteOnResourceDeletion {
	// 	if org.ObjectMeta.DeletionTimestamp.IsZero() {
	// 		// not being deleted
	// 		if !controllerutil.ContainsFinalizer(org, organizationFinalizerName) {
	// 			controllerutil.AddFinalizer(org, organizationFinalizerName)
	// 			if err := r.Update(ctx, org); err != nil {
	// 				return ctrl.Result{}, err
	// 			}
	// 		}
	// 	} else {
	// 		// being deleted
	// 		log.Info("deleting organization", "login", org.Status.Login)
	// 		if org.Status.LastUpdateTimestamp != nil {
	// 			// if we have never resolved this resource before, don't
	// 			// touch external state
	// 			if err := r.deleteOrganization(ctx, org); err != nil {
	// 				log.Error(err, "unable to delete organization")
	// 				return ctrl.Result{}, err
	// 			}
	// 		}

	// 		controllerutil.RemoveFinalizer(org, teamFinalizerName)
	// 		if err := r.Update(ctx, org); err != nil {
	// 			return ctrl.Result{}, err
	// 		}

	// 		return ctrl.Result{}, nil
	// 	}
	// }

	// update team
	err := r.updateOrganization(ctx, org, observed)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
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
	if organization.Spec.Name != ghOrganization.GetName() {
		log.Info("organization name update", "from", ghOrganization.GetName(), "to", organization.Spec.Name)
		updateOrg.Name = github.String(organization.Spec.Name)
	}
	// billing email
	if organization.Spec.BillingEmail != ghOrganization.GetBillingEmail() {
		log.Info("organization billing email update", "from", ghOrganization.GetBillingEmail(), "to", organization.Spec.BillingEmail)
		updateOrg.BillingEmail = github.String(organization.Spec.BillingEmail)
	}
	// company
	if organization.Spec.Company != ghOrganization.GetCompany() {
		log.Info("organization company update", "from", ghOrganization.GetCompany(), "to", organization.Spec.Company)
		updateOrg.Company = github.String(organization.Spec.Company)
	}
	// email
	if organization.Spec.Email != ghOrganization.GetEmail() {
		log.Info("organization email update", "from", ghOrganization.GetEmail(), "to", organization.Spec.Email)
		updateOrg.Email = github.String(organization.Spec.Email)
	}
	// twitter username
	if ptrNonNilAndNotEqualTo(organization.Spec.TwitterUsername, ghOrganization.GetTwitterUsername()) {
		log.Info("organization twitter username update", "from", ghOrganization.GetTwitterUsername(), "to", *organization.Spec.TwitterUsername)
		updateOrg.TwitterUsername = organization.Spec.TwitterUsername
	}
	// location
	if ptrNonNilAndNotEqualTo(organization.Spec.Location, ghOrganization.GetLocation()) {
		log.Info("organization location update", "from", ghOrganization.GetLocation(), "to", *organization.Spec.Location)
		updateOrg.Location = organization.Spec.Location
	}
	// description
	if ptrNonNilAndNotEqualTo(organization.Spec.Description, ghOrganization.GetDescription()) {
		log.Info("organization description update", "from", ghOrganization.GetDescription(), "to", *organization.Spec.Description)
		updateOrg.Description = organization.Spec.Description
	}
	// has organization projects
	if ptrNonNilAndNotEqualTo(organization.Spec.HasOrganizationProjects, ghOrganization.GetHasOrganizationProjects()) {
		log.Info("organization hasOrganizationProjects update", "from", ghOrganization.GetHasOrganizationProjects(), "to", *organization.Spec.HasOrganizationProjects)
		updateOrg.HasOrganizationProjects = organization.Spec.HasOrganizationProjects
	}
	// has repository projects
	if ptrNonNilAndNotEqualTo(organization.Spec.HasRepositoryProjects, ghOrganization.GetHasRepositoryProjects()) {
		log.Info("organization hasRepositoryProjects update", "from", ghOrganization.GetHasRepositoryProjects(), "to", *organization.Spec.HasRepositoryProjects)
		updateOrg.HasRepositoryProjects = organization.Spec.HasRepositoryProjects
	}
	// default repository permission
	if ptrNonNilAndNotEqualTo(organization.Spec.DefaultRepositoryPermission, ghOrganization.GetDefaultRepoPermission()) {
		log.Info("organization defaultRepositoryPermission update", "from", ghOrganization.GetDefaultRepoPermission(), "to", *organization.Spec.DefaultRepositoryPermission)
		updateOrg.DefaultRepoPermission = organization.Spec.DefaultRepositoryPermission
	}
	// members can create repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreateRepositories, ghOrganization.GetMembersCanCreateRepos()) {
		log.Info("organization membersCanCreateRepositories update", "from", ghOrganization.GetMembersCanCreateRepos(), "to", *organization.Spec.MembersCanCreateRepositories)
		updateOrg.MembersCanCreateRepos = organization.Spec.MembersCanCreateRepositories
	}
	// members can create internal repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreateInternalRepositories, ghOrganization.GetMembersCanCreateInternalRepos()) {
		log.Info("organization membersCanCreateInternalRepositories update", "from", ghOrganization.GetMembersCanCreateInternalRepos(), "to", *organization.Spec.MembersCanCreateInternalRepositories)
		updateOrg.MembersCanCreateInternalRepos = organization.Spec.MembersCanCreateInternalRepositories
	}
	// members can create private repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePrivateRepositories, ghOrganization.GetMembersCanCreatePrivateRepos()) {
		log.Info("organization membersCanCreatePrivateRepositories update", "from", ghOrganization.GetMembersCanCreatePrivateRepos(), "to", *organization.Spec.MembersCanCreatePrivateRepositories)
		updateOrg.MembersCanCreatePrivateRepos = organization.Spec.MembersCanCreatePrivateRepositories
	}
	// members can create public repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePublicRepositories, ghOrganization.GetMembersCanCreatePublicRepos()) {
		log.Info("organization membersCanCreatePublicRepositories update", "from", ghOrganization.GetMembersCanCreatePublicRepos(), "to", *organization.Spec.MembersCanCreatePublicRepositories)
		updateOrg.MembersCanCreatePublicRepos = organization.Spec.MembersCanCreatePublicRepositories
	}
	// members can create pages
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePages, ghOrganization.GetMembersCanCreatePages()) {
		log.Info("organization membersCanCreatePages update", "from", ghOrganization.GetMembersCanCreatePages(), "to", *organization.Spec.MembersCanCreatePages)
		updateOrg.MembersCanCreatePages = organization.Spec.MembersCanCreatePages
	}
	// members can create public pages
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePublicPages, ghOrganization.GetMembersCanCreatePublicPages()) {
		log.Info("organization membersCanCreatePublicPages update", "from", ghOrganization.GetMembersCanCreatePublicPages(), "to", *organization.Spec.MembersCanCreatePublicPages)
		updateOrg.MembersCanCreatePublicPages = organization.Spec.MembersCanCreatePublicPages
	}
	// members can create private pages
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanCreatePrivatePages, ghOrganization.GetMembersCanCreatePrivatePages()) {
		log.Info("organization membersCanCreatePrivatePages update", "from", ghOrganization.GetMembersCanCreatePrivatePages(), "to", *organization.Spec.MembersCanCreatePrivatePages)
		updateOrg.MembersCanCreatePrivatePages = organization.Spec.MembersCanCreatePrivatePages
	}
	// members can fork private repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.MembersCanForkPrivateRepositories, ghOrganization.GetMembersCanForkPrivateRepos()) {
		log.Info("organization membersCanForkPrivateRepositories update", "from", ghOrganization.GetMembersCanForkPrivateRepos(), "to", *organization.Spec.MembersCanForkPrivateRepositories)
		updateOrg.MembersCanForkPrivateRepos = organization.Spec.MembersCanForkPrivateRepositories
	}
	// web commit signoff required
	if ptrNonNilAndNotEqualTo(organization.Spec.WebCommitSignoffRequired, ghOrganization.GetWebCommitSignoffRequired()) {
		log.Info("organization webCommitSignoffRequired update", "from", ghOrganization.GetWebCommitSignoffRequired(), "to", *organization.Spec.WebCommitSignoffRequired)
		updateOrg.WebCommitSignoffRequired = organization.Spec.WebCommitSignoffRequired
	}
	// blog
	if ptrNonNilAndNotEqualTo(organization.Spec.Blog, ghOrganization.GetBlog()) {
		log.Info("organization blog update", "from", ghOrganization.GetBlog(), "to", *organization.Spec.Blog)
		updateOrg.Blog = organization.Spec.Blog
	}
	// advanced security enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.AdvancedSecurityEnabledForNewRepositories, ghOrganization.GetAdvancedSecurityEnabledForNewRepos()) {
		log.Info("organization advancedSecurityEnabledForNewRepositories update", "from", ghOrganization.GetAdvancedSecurityEnabledForNewRepos(), "to", *organization.Spec.AdvancedSecurityEnabledForNewRepositories)
		updateOrg.AdvancedSecurityEnabledForNewRepos = organization.Spec.AdvancedSecurityEnabledForNewRepositories
	}
	// dependabot alerts enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.DependabotAlertsEnabledForNewRepositories, ghOrganization.GetAdvancedSecurityEnabledForNewRepos()) {
		log.Info("organization dependabotAlertsEnabledForNewRepositories update", "from", ghOrganization.GetAdvancedSecurityEnabledForNewRepos(), "to", *organization.Spec.DependabotAlertsEnabledForNewRepositories)
		updateOrg.DependabotAlertsEnabledForNewRepos = organization.Spec.DependabotAlertsEnabledForNewRepositories
	}
	// dependabot secruity updates enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.DependabotSecurityUpdatesEnabledForNewRepositories, ghOrganization.GetDependabotSecurityUpdatesEnabledForNewRepos()) {
		log.Info("organization dependabotSecurityUpdatesEnabledForNewRepositories update", "from", ghOrganization.GetDependabotSecurityUpdatesEnabledForNewRepos(), "to", *organization.Spec.DependabotAlertsEnabledForNewRepositories)
		updateOrg.DependabotSecurityUpdatesEnabledForNewRepos = organization.Spec.DependabotSecurityUpdatesEnabledForNewRepositories
	}
	// dependency graph enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.DependencyGraphEnabledForNewRepositories, ghOrganization.GetDependencyGraphEnabledForNewRepos()) {
		log.Info("organization dependencyGraphEnabledForNewRepositories update", "from", ghOrganization.GetDependencyGraphEnabledForNewRepos(), "to", *organization.Spec.DependencyGraphEnabledForNewRepositories)
		updateOrg.DependencyGraphEnabledForNewRepos = organization.Spec.DependencyGraphEnabledForNewRepositories
	}
	// secret scanning enabled for new repositories
	if ptrNonNilAndNotEqualTo(organization.Spec.SecretScanningEnabledForNewRepositories, ghOrganization.GetSecretScanningEnabledForNewRepos()) {
		log.Info("organization secretScanningEnabledForNewRepositories update", "from", ghOrganization.GetSecretScanningEnabledForNewRepos(), "to", *organization.Spec.SecretScanningEnabledForNewRepositories)
		updateOrg.SecretScanningEnabledForNewRepos = organization.Spec.SecretScanningEnabledForNewRepositories
	}

	// perform update if necessary
	if needsUpdate || organization.Status.LastUpdateTimestamp == nil {
		log.Info("updating organization", "name", organization.Spec.Name)
		updated, err := r.GitHubClient.UpdateOrganization(ctx, *ghOrganization.Login, &updateOrg)
		if err != nil {
			log.Error(err, "unable to update organization", "name", organization.Spec.Name)
			return err
		}
		ghOrganization = updated

		now := v1.Now()
		organization.Status = githubv1alpha1.OrganizationStatus{
			Login:                                updated.Login,
			Id:                                   updated.ID,
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
			DefaultRepositoryPermission:          updated.DefaultRepoPermission,
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
