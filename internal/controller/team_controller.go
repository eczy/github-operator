/*
Copyright 2025.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	githuboperatoreczyiov1 "github.com/eczy/github-operator/api/v1"
	gh "github.com/eczy/github-operator/internal/github"
	"github.com/google/go-github/v60/github"
)

type TeamRequester interface {
	GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, error)
	GetTeamById(ctx context.Context, org, teamId int64) (*github.Team, error)
	GetTeamByNodeId(ctx context.Context, nodeId string) (*github.Team, error)

	CreateTeam(ctx context.Context, org string, newTeam github.NewTeam) (*github.Team, error)
	UpdateTeamBySlug(ctx context.Context, org, slug string, newTeam github.NewTeam) (*github.Team, error)
	UpdateTeamById(ctx context.Context, org, teamId int64, newTeam github.NewTeam) (*github.Team, error)
	DeleteTeamBySlug(ctx context.Context, org, slug string) error
	DeleteTeamById(ctx context.Context, org, teamId int64) error

	GetTeamRepositoryPermission(ctx context.Context, org, slug, repoName string) (*gh.TeamRepositoryPermission, error)
	GetTeamRepositoryPermissions(ctx context.Context, org, slug string) ([]*gh.TeamRepositoryPermission, error)
	UpdateTeamRepositoryPermissions(ctx context.Context, org, slug string, repoName, permission string) error
	RemoveTeamRepositoryPermissions(ctx context.Context, org, slug string, repoName string) error
}

// TeamReconciler reconciles a Team object
type TeamReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=github-operator.eczy.io,resources=teams,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=github-operator.eczy.io,resources=teams/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=github-operator.eczy.io,resources=teams/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Team object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *TeamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&githuboperatoreczyiov1.Team{}).
		Named("team").
		Complete(r)
}
