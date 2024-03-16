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

	// TODO(user): your logic here

	if r.GitHubClient == nil {
		err := fmt.Errorf("nil GitHub client")
		log.Error(err, "reconciler GitHub client is nil")
		return ctrl.Result{}, err
	}

	// fetch team
	team := &githubv1alpha1.Team{}
	if err := r.Get(ctx, req.NamespacedName, team); err != nil {
		log.Error(err, "unable to fetch Team")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// add deletion finalizer if specified
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

	// create or fetch team
	var observed *github.Team
	// if Status.Id is nil, need to create a new team
	// TODO: or find the existing one from GitHub
	if team.Status.Id == nil {
		log.Info("creating new repository", "name", team.Spec.Name)
		created, err := r.createTeam(ctx, team)
		if err != nil {
			log.Error(err, "unable to create Team")
			return ctrl.Result{}, err
		}
		if created.ID == nil {
			log.Error(err, "team ID is nil", "team", created)
			return ctrl.Result{}, err
		} else {
			team.Status.Id = created.ID
		}
		if created.Slug == nil {
			log.Error(err, "team slug is nil", "team", created)
			return ctrl.Result{}, err
		} else {
			team.Status.Slug = created.Slug
		}
		observed = created
	} else {
		existing, err := r.GitHubClient.GetTeamBySlug(ctx, team.Spec.Organization, *team.Status.Slug) // TODO: do this by id
		if err != nil {
			return ctrl.Result{}, err
		}
		observed = existing
	}

	// update team
	updateTeam := github.NewTeam{}
	needsUpdate := false

	// resolve owner
	// TODO

	// resolve name
	if !cmpPtrToValue(observed.Name, team.Spec.Name) {
		updateTeam.Name = team.Spec.Name
		needsUpdate = true
	}
	if team.Spec.Name != team.GetObjectMeta().GetName() {
		log.Info("team spec Name does not match metadata Name", "spec name", team.Spec.Name, "meta name", team.GetObjectMeta().GetName())
	}

	// resolve description
	if !cmpPtrValues(team.Spec.Description, observed.Description) {
		updateTeam.Description = team.Spec.Description
		needsUpdate = true
	}

	// resolve privacy
	if !cmpPtrValues(team.Spec.Privacy, (*githubv1alpha1.Privacy)(observed.Privacy)) {
		updateTeam.Privacy = (*string)(team.Spec.Privacy)
		needsUpdate = true
	}

	// resolve parent
	if observed.Parent != nil {
		if !cmpPtrValues(team.Spec.ParentTeamId, observed.Parent.ID) {
			updateTeam.ParentTeamID = team.Spec.ParentTeamId
			needsUpdate = true
		}
	} else {
		if team.Spec.ParentTeamId != nil {
			updateTeam.ParentTeamID = team.Spec.ParentTeamId
			needsUpdate = true
		}
	}

	// TODO: team members and maintainers

	// perform update if necessary
	if needsUpdate {
		_, err := r.GitHubClient.UpdateTeamById(ctx, *observed.Organization.ID, *team.Status.Id, updateTeam)
		if err != nil {
			log.Error(err, "unable to update team", "team", *observed, "update", updateTeam)
			return ctrl.Result{}, err
		}
		now := v1.Now()
		team.Status.LastUpdateTimestamp = &now
	}

	// update status
	if err := r.Status().Update(ctx, team); err != nil {
		log.Error(err, "unable to update Team status")
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

func (r *TeamReconciler) createTeam(ctx context.Context, team *githubv1alpha1.Team) (*github.Team, error) {
	newTeam, err := teamToNewTeam(team)
	if err != nil {
		return nil, fmt.Errorf("creating github.NewTeam object: %w", err)
	}
	created, err := r.GitHubClient.CreateTeam(ctx, team.Spec.Organization, newTeam)
	if err != nil {
		return nil, fmt.Errorf("creating GitHub Team: %w", err)
	}
	return created, nil
}

func (r *TeamReconciler) deleteTeam(ctx context.Context, team *githubv1alpha1.Team) error {
	return r.GitHubClient.DeleteTeamBySlug(ctx, team.Spec.Organization, *team.Status.Slug)
}

func teamToNewTeam(team *githubv1alpha1.Team) (github.NewTeam, error) {
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

func cmpPtrValues[T comparable](a *T, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func cmpPtrToValue[T comparable](a *T, b T) bool {
	if a == nil {
		return false
	}
	return *a == b
}
