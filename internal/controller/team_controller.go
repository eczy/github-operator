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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
	"github.com/google/go-github/v60/github"
)

type TeamRequester interface {
	GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, error)
	GetTeamById(ctx context.Context, org, teamId int64) (*github.Team, error)
	CreateTeam(ctx context.Context, org string, newTeam github.NewTeam) (*github.Team, error)
	UpdateTeamBySlug(ctx context.Context, org, slug string, newTeam github.NewTeam) (*github.Team, error)
	UpdateTeamById(ctx context.Context, org, teamId int64, newTeam github.NewTeam) (*github.Team, error)
	DeleteTeamBySlug(ctx context.Context, org, slug string) error
	DeleteTeamById(ctx context.Context, org, teamId int64) error
}

// TeamReconciler reconciles a Team object
type TeamReconciler struct {
	client.Client
	Scheme                   *runtime.Scheme
	GitHubClient             TeamRequester
	DeleteOnResourceDeletion bool
}

//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=teams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=teams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=teams/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Team object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *TeamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if r.GitHubClient == nil {
		err := fmt.Errorf("nil GitHub client")
		log.Error(err, "reconciler GitHub client is nil")
		return ctrl.Result{}, err
	}

	// fetch team resource
	team := &githubv1alpha1.Team{}
	if err := r.Get(ctx, req.NamespacedName, team); err != nil {
		log.Error(err, "unable to fetch Team")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// try to fetch external resource
	observed, err := r.GitHubClient.GetTeamBySlug(ctx, team.Spec.Organization, team.Spec.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	// if team does't exist, check if scheduled for deletion
	if observed == nil {
		// if scheduled for deletion
		if team.ObjectMeta.DeletionTimestamp.IsZero() {
			// do nothing and return since the external resource doesn't exist
			return ctrl.Result{}, nil
		} else {
			// otherwise create the external resource
			ghTeam, err := r.createTeam(ctx, team)
			if err != nil {
				return ctrl.Result{}, err
			}
			observed = ghTeam
		}
	}

	// handle finalizer
	finalizerName := "github.github-operator.eczy.io/team-finalizer"
	if r.DeleteOnResourceDeletion {
		if team.ObjectMeta.DeletionTimestamp.IsZero() {
			// not being deleted
			if !controllerutil.ContainsFinalizer(team, finalizerName) {
				controllerutil.AddFinalizer(team, finalizerName)
				if err := r.Update(ctx, team); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else {
			// being deleted
			if err := r.deleteTeam(ctx, team); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(team, finalizerName)
			if err := r.Update(ctx, team); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// update team
	err = r.updateTeam(ctx, team, observed)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&githubv1alpha1.Team{}).
		Complete(r)
}

// modifies team in place
func (r *TeamReconciler) createTeam(ctx context.Context, team *githubv1alpha1.Team) (*github.Team, error) {
	newTeam, err := teamResourceToNewTeam(team)
	if err != nil {
		return nil, fmt.Errorf("creating github.NewTeam object: %w", err)
	}
	created, err := r.GitHubClient.CreateTeam(ctx, team.Spec.Organization, newTeam)
	if err != nil {
		return nil, fmt.Errorf("creating GitHub Team: %w", err)
	}
	return created, nil
}

// modifies team and ghTeam in place
func (r *TeamReconciler) updateTeam(ctx context.Context, team *githubv1alpha1.Team, ghTeam *github.Team) error {
	log := log.FromContext(ctx)

	updateTeam := github.NewTeam{}
	needsUpdate := false

	// resolve owner
	// TODO

	// resolve name
	if team.Spec.Name != ghTeam.GetName() {
		updateTeam.Name = team.Spec.Name
		needsUpdate = true
	}
	if team.Spec.Name != team.GetObjectMeta().GetName() {
		log.Info("team spec Name does not match metadata Name", "spec", team.Spec.Name, "metadata", team.GetObjectMeta().GetName())
	}

	// resolve description
	if ptrNonNilAndNotEqualTo(team.Spec.Description, ghTeam.GetDescription()) {
		updateTeam.Description = team.Spec.Description
		needsUpdate = true
	}

	// resolve privacy
	if ptrNonNilAndNotEqualTo(team.Spec.Privacy, githubv1alpha1.Privacy(ghTeam.GetPrivacy())) {
		updateTeam.Privacy = (*string)(team.Spec.Privacy)
		needsUpdate = true
	}

	// resolve parent
	parent := ghTeam.GetParent()
	if parent != nil {
		if ptrNonNilAndNotEqualTo(team.Spec.ParentTeamId, parent.GetID()) {
			updateTeam.ParentTeamID = team.Spec.ParentTeamId
			needsUpdate = true
		} else if team.Spec.ParentTeamId == nil {
			updateTeam.ParentTeamID = nil
			needsUpdate = true
		}
	} else if team.Spec.ParentTeamId != nil {
		updateTeam.ParentTeamID = team.Spec.ParentTeamId
		needsUpdate = true
	}

	// TODO: team members and maintainers

	// perform update if necessary
	if needsUpdate {
		updated, err := r.GitHubClient.UpdateTeamById(ctx, *ghTeam.Organization.ID, *ghTeam.ID, updateTeam)
		if err != nil {
			log.Error(err, "unable to update team", "name", team.Spec.Name)
			return err
		}
		ghTeam = updated

		now := v1.Now()
		parent := ghTeam.GetParent()
		var parentId *int64
		var parentSlug *string
		if parent != nil {
			parentId = parent.ID
			parentSlug = parent.Slug
		}
		team.Status = githubv1alpha1.TeamStatus{
			Id:                  ghTeam.ID,
			Slug:                ghTeam.Slug,
			LastUpdateTimestamp: &now,
			OrganizationLogin:   ghTeam.GetOrganization().GetLogin(),
			OrganizationSlug:    ghTeam.GetOrganization().GetID(),
			Name:                ghTeam.GetName(),
			Description:         ghTeam.Description,
			// TODO
			// Members:             []string{},
			// Maintainers:         []string{},
			Privacy: (*githubv1alpha1.Privacy)(ghTeam.Privacy),
			// TODO
			// NotificationSetting: &"",
			ParentTeamId:   parentId,
			ParentTeamSlug: parentSlug,
		}

		// update status
		if err := r.Status().Update(ctx, team); err != nil {
			log.Error(err, "unable to update Team status", "name", team.Spec.Name)
		}
	}

	return nil
}

func (r *TeamReconciler) deleteTeam(ctx context.Context, team *githubv1alpha1.Team) error {
	return r.GitHubClient.DeleteTeamBySlug(ctx, team.Spec.Organization, *team.Status.Slug)
}

// teamResourceToNewTeam creates a github.NewTeam instance from a Team resource
func teamResourceToNewTeam(team *githubv1alpha1.Team) (github.NewTeam, error) {
	var privacy *string
	if team.Spec.Privacy != nil {
		tmp := string(*team.Spec.Privacy)
		privacy = &tmp
	}
	newTeam := github.NewTeam{
		Name:         team.Spec.Name,
		Description:  team.Spec.Description,
		ParentTeamID: team.Spec.ParentTeamId,
		Privacy:      privacy,
	}
	return newTeam, nil
}
