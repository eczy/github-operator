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
	gh "github.com/eczy/github-operator/internal/github"
	"github.com/shurcooL/githubv4"
)

var (
	branchProtectionFinalizerName = "github.github-operator.eczy.io/branch-protection-finalizer"
)

type BranchProtectionRequester interface {
	RepositoryGetter // needed to create new branch protection rules

	GetBranchProtection(ctx context.Context, nodeId string) (*gh.BranchProtection, error)
	CreateBranchProtection(ctx context.Context, input *githubv4.CreateBranchProtectionRuleInput) (*gh.BranchProtection, error)
	GetBranchProtectionByOwnerRepoPattern(ctx context.Context, repositoryOwner, repositoryName, pattern string) (*gh.BranchProtection, error)
	UpdateBranchProtection(ctx context.Context, input *githubv4.UpdateBranchProtectionRuleInput) (*gh.BranchProtection, error)
	DeleteBranchProtection(ctx context.Context, input *githubv4.DeleteBranchProtectionRuleInput) error
}

// BranchProtectionReconciler reconciles a BranchProtection object
type BranchProtectionReconciler struct {
	client.Client
	Scheme                   *runtime.Scheme
	GitHubClient             BranchProtectionRequester
	DeleteOnResourceDeletion bool
}

//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=branchprotections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=branchprotections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=github.github-operator.eczy.io,resources=branchprotections/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BranchProtection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *BranchProtectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO: move this check elsewhere
	if r.GitHubClient == nil {
		err := fmt.Errorf("nil GitHub client")
		log.Error(err, "reconciler GitHub client is nil")
		return ctrl.Result{}, err
	}

	// fetch team resource
	bp := &githubv1alpha1.BranchProtection{}
	if err := r.Get(ctx, req.NamespacedName, bp); err != nil {
		log.Error(err, "error fetching BranchProtection resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var observed *gh.BranchProtection
	// try to fetch external resource
	if bp.Status.Id != nil {
		log.Info("getting branch protection", "id", bp.Status.Id)
		ghBp, err := r.GitHubClient.GetBranchProtection(ctx, *bp.Status.Id)
		if err != nil {
			// TODO: determine if the error is a "not found" error vs another type of error
			log.Info(err.Error())
		}
		observed = ghBp
	} else {
		log.Info("getting branch protection", "repository", bp.Spec.RepositoryName, "owner", bp.Spec.RepositoryOwner, "pattern", bp.Spec.Pattern)
		ghBp, err := r.GitHubClient.GetBranchProtectionByOwnerRepoPattern(ctx, bp.Spec.RepositoryOwner, bp.Spec.RepositoryName, bp.Spec.Pattern)
		if err != nil {
			// TODO: determine if the error is a "not found" error vs another type of error
			log.Info(err.Error())
		}
		observed = ghBp
	}

	// if branch protection does't exist, check if scheduled for deletion
	if observed == nil {
		// if scheduled for deletion
		if !bp.ObjectMeta.DeletionTimestamp.IsZero() {
			// do nothing and return since the external resource doesn't exist
			return ctrl.Result{}, nil
		} else {
			// otherwise create the external resource
			log.Info("creating branch protection", "repository", bp.Spec.RepositoryName, "owner", bp.Spec.RepositoryOwner, "pattern", bp.Spec.Pattern)
			ghBp, err := r.createBranchProtection(ctx, bp)
			if err != nil {
				log.Error(err, "unable to create branch protection", "pattern", bp.Spec.Pattern)
				return ctrl.Result{}, err
			}
			observed = ghBp
		}
	}

	// handle finalizer
	if r.DeleteOnResourceDeletion {
		if bp.ObjectMeta.DeletionTimestamp.IsZero() {
			// not being deleted
			if !controllerutil.ContainsFinalizer(bp, branchProtectionFinalizerName) {
				controllerutil.AddFinalizer(bp, branchProtectionFinalizerName)
				if err := r.Update(ctx, bp); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else {
			// being deleted
			log.Info("deleting branch protection", "node_id", bp.Status.Id)
			if bp.Status.LastUpdateTimestamp != nil {
				// if we have never resolved this resource before, don't
				// touch external state
				if err := r.deleteBranchProtection(ctx, bp); err != nil {
					log.Error(err, "unable to delete branch protection")
					return ctrl.Result{}, err
				}
			}

			controllerutil.RemoveFinalizer(bp, branchProtectionFinalizerName)
			if err := r.Update(ctx, bp); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	}

	// update team
	err := r.updateBranchProtection(ctx, bp, observed)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BranchProtectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&githubv1alpha1.BranchProtection{}).
		Complete(r)
}

func (r *BranchProtectionReconciler) createBranchProtection(ctx context.Context, bp *githubv1alpha1.BranchProtection) (*gh.BranchProtection, error) {
	input, err := r.branchProtectionToCreateInput(ctx, bp)
	if err != nil {
		return nil, err
	}
	branchProtection, err := r.GitHubClient.CreateBranchProtection(ctx, input)
	if err != nil {
		return nil, err
	}
	return branchProtection, nil
}

func (r *BranchProtectionReconciler) updateBranchProtection(ctx context.Context, bp *githubv1alpha1.BranchProtection, ghBp *gh.BranchProtection) error {
	log := log.FromContext(ctx)

	update := githubv4.UpdateBranchProtectionRuleInput{
		BranchProtectionRuleID: ghBp.Id,
	}
	needsUpdate := false

	// Pattern
	if bp.Spec.Pattern != ghBp.Pattern {
		update.Pattern = (*githubv4.String)(&bp.Spec.Pattern)
		needsUpdate = true
	}

	// AllowsDeletions
	if ptrNonNilAndNotEqualTo(bp.Spec.AllowsDeletions, ghBp.AllowsDeletions) {
		update.AllowsDeletions = (*githubv4.Boolean)(bp.Spec.AllowsDeletions)
		needsUpdate = true
	}
	// AllowsForcePushes
	if ptrNonNilAndNotEqualTo(bp.Spec.AllowsForcePushes, ghBp.AllowsForcePushes) {
		update.AllowsForcePushes = (*githubv4.Boolean)(bp.Spec.AllowsForcePushes)
		needsUpdate = true
	}
	// BlocksCreations
	if ptrNonNilAndNotEqualTo(bp.Spec.BlocksCreations, ghBp.BlocksCreations) {
		update.BlocksCreations = (*githubv4.Boolean)(bp.Spec.BlocksCreations)
		needsUpdate = true
	}

	specBypassForcePushIds := []githubv4.ID{}
	needsBypassForcePushUpdate := false
	// BypassForcePushUsers
	userLogins := []string{}
	userIds := []githubv4.ID{}
	for _, user := range ghBp.GetPushAllowances().Users {
		userIds = append(userIds, user.Id)
		userLogins = append(userLogins, user.Login)
	}
	specBypassForcePushIds = append(specBypassForcePushIds, userIds...)
	if !cmpSlices(bp.Spec.BypassForcePushUsers, userLogins) {
		needsBypassForcePushUpdate = true
	}
	// BypassForcePushApps
	appSlugs := []string{}
	appIds := []githubv4.ID{}
	for _, app := range ghBp.GetPushAllowances().Apps {
		appIds = append(appIds, app.Id)
		appSlugs = append(appSlugs, app.Slug)
	}
	specBypassForcePushIds = append(specBypassForcePushIds, appIds...)
	if !cmpSlices(bp.Spec.PushAllowanceApps, appSlugs) {
		needsBypassForcePushUpdate = true
	}
	// BypassForcePushTeams
	teamSlugs := []string{}
	teamIds := []githubv4.ID{}
	for _, team := range ghBp.GetPushAllowances().Teams {
		teamIds = append(appIds, team.Id)
		teamSlugs = append(appSlugs, team.Slug)
	}
	specBypassForcePushIds = append(specBypassForcePushIds, teamIds...)
	if !cmpSlices(bp.Spec.PushAllowanceTeams, teamSlugs) {
		needsBypassForcePushUpdate = true
	}
	if needsBypassForcePushUpdate {
		update.PushActorIDs = &specBypassForcePushIds
		needsUpdate = true
	}

	specBypassPullRequestIds := []githubv4.ID{}
	needsBypassPullRequestUpdate := false
	// BypassPullRequestUsers
	userLogins = []string{}
	userIds = []githubv4.ID{}
	for _, user := range ghBp.GetPushAllowances().Users {
		userIds = append(userIds, user.Id)
		userLogins = append(userLogins, user.Login)
	}
	specBypassPullRequestIds = append(specBypassPullRequestIds, userIds...)
	if !cmpSlices(bp.Spec.BypassPullRequestUsers, userLogins) {
		needsBypassPullRequestUpdate = true
	}
	// BypassPullRequestApps
	appSlugs = []string{}
	appIds = []githubv4.ID{}
	for _, app := range ghBp.GetPushAllowances().Apps {
		appIds = append(appIds, app.Id)
		appSlugs = append(appSlugs, app.Slug)
	}
	specBypassPullRequestIds = append(specBypassPullRequestIds, appIds...)
	if !cmpSlices(bp.Spec.PushAllowanceApps, appSlugs) {
		needsBypassPullRequestUpdate = true
	}
	// BypassPullRequestTeams
	teamSlugs = []string{}
	teamIds = []githubv4.ID{}
	for _, team := range ghBp.GetPushAllowances().Teams {
		teamIds = append(appIds, team.Id)
		teamSlugs = append(appSlugs, team.Slug)
	}
	specBypassPullRequestIds = append(specBypassPullRequestIds, teamIds...)
	if !cmpSlices(bp.Spec.PushAllowanceTeams, teamSlugs) {
		needsBypassPullRequestUpdate = true
	}
	if needsBypassPullRequestUpdate {
		update.PushActorIDs = &specBypassPullRequestIds
		needsUpdate = true
	}

	// DismissesStaleReviews
	if ptrNonNilAndNotEqualTo(bp.Spec.DismissesStaleReviews, ghBp.DismissesStaleReviews) {
		update.DismissesStaleReviews = (*githubv4.Boolean)(bp.Spec.DismissesStaleReviews)
		needsUpdate = true
	}
	// IsAdminEnforced
	if ptrNonNilAndNotEqualTo(bp.Spec.AllowsDeletions, ghBp.AllowsDeletions) {
		update.IsAdminEnforced = (*githubv4.Boolean)(bp.Spec.IsAdminEnforced)
		needsUpdate = true
	}
	// LockAllowsFetchAndMerge
	if ptrNonNilAndNotEqualTo(bp.Spec.LockAllowsFetchAndMerge, ghBp.LockAllowsFetchAndMerge) {
		update.LockAllowsFetchAndMerge = (*githubv4.Boolean)(bp.Spec.LockAllowsFetchAndMerge)
		needsUpdate = true
	}
	// LockBranch
	if ptrNonNilAndNotEqualTo(bp.Spec.LockBranch, ghBp.LockBranch) {
		update.LockBranch = (*githubv4.Boolean)(bp.Spec.LockBranch)
		needsUpdate = true
	}

	specPushActorIds := []githubv4.ID{}
	needsPushAllowanceUpdate := false
	// PushAllowanceUsers
	userLogins = []string{}
	userIds = []githubv4.ID{}
	for _, user := range ghBp.GetPushAllowances().Users {
		userIds = append(userIds, user.Id)
		userLogins = append(userLogins, user.Login)
	}
	specPushActorIds = append(specPushActorIds, userIds...)
	if !cmpSlices(bp.Spec.PushAllowanceUsers, userLogins) {
		needsPushAllowanceUpdate = true
	}
	// PushAllowanceApps
	appSlugs = []string{}
	appIds = []githubv4.ID{}
	for _, app := range ghBp.GetPushAllowances().Apps {
		appIds = append(appIds, app.Id)
		appSlugs = append(appSlugs, app.Slug)
	}
	specPushActorIds = append(specPushActorIds, appIds...)
	if !cmpSlices(bp.Spec.PushAllowanceApps, appSlugs) {
		needsPushAllowanceUpdate = true
	}
	// PushAllowanceTeams
	teamSlugs = []string{}
	teamIds = []githubv4.ID{}
	for _, team := range ghBp.GetPushAllowances().Teams {
		teamIds = append(appIds, team.Id)
		teamSlugs = append(appSlugs, team.Slug)
	}
	specPushActorIds = append(specPushActorIds, teamIds...)
	if !cmpSlices(bp.Spec.PushAllowanceTeams, teamSlugs) {
		needsPushAllowanceUpdate = true
	}
	if needsPushAllowanceUpdate {
		update.PushActorIDs = &specPushActorIds
		needsUpdate = true
	}

	// RequireLastPushApproval
	if ptrNonNilAndNotEqualTo(bp.Spec.RequireLastPushApproval, ghBp.RequireLastPushApproval) {
		update.RequireLastPushApproval = (*githubv4.Boolean)(bp.Spec.RequireLastPushApproval)
		needsUpdate = true
	}
	// RequiredApprovingReviewCount
	ghCount := int(ghBp.RequiredApprovingReviewCount)
	if ptrNonNilAndNotEqualTo(bp.Spec.RequiredApprovingReviewCount, ghCount) {
		val := *bp.Spec.RequiredApprovingReviewCount
		update.RequiredApprovingReviewCount = githubv4.NewInt(githubv4.Int(val))
		needsUpdate = true
	}
	// RequiredDeploymentEnvironments
	if !cmpSlices(bp.Spec.RequiredDeploymentEnvironments, ghBp.RequiredDeploymentEnvironments) {
		conv := []githubv4.String{}
		for _, x := range bp.Spec.RequiredDeploymentEnvironments {
			conv = append(conv, githubv4.String(x))
		}
		update.RequiredDeploymentEnvironments = &conv
		needsUpdate = true
	}
	// RequiredStatusCheckContexts
	if !cmpSlices(bp.Spec.RequiredStatusCheckContexts, ghBp.RequiredStatusCheckContexts) {
		conv := []githubv4.String{}
		for _, x := range bp.Spec.RequiredStatusCheckContexts {
			conv = append(conv, githubv4.String(x))
		}
		update.RequiredStatusCheckContexts = &conv
		needsUpdate = true
	}
	// RequiredStatusChecks
	ghChecks := map[string]gh.RequiredStatusCheckDescription{}
	updateChecks := []githubv4.RequiredStatusCheckInput{}
	requiredStatusChecksNeedUpdate := len(ghBp.RequiredStatusChecks) == len(bp.Spec.RequiredStatusChecks)
	for _, check := range ghBp.RequiredStatusChecks {
		ghChecks[check.Context] = check
	}
	for _, check := range bp.Spec.RequiredStatusChecks {
		var appId githubv4.ID
		if check.AppId != nil {
			appId = check.AppId
		}
		updateChecks = append(updateChecks, githubv4.RequiredStatusCheckInput{
			Context: githubv4.String(check.Context),
			AppID:   &appId,
		})
		if ghCheck, ok := ghChecks[check.Context]; ok {
			if ghCheck.Context != check.Context {
				requiredStatusChecksNeedUpdate = true
			} else if !ptrNonNilAndNotEqualTo(check.AppId, ghCheck.App.Id) {
				requiredStatusChecksNeedUpdate = true
			}
		} else {
			requiredStatusChecksNeedUpdate = true
		}
	}
	if requiredStatusChecksNeedUpdate {
		update.RequiredStatusChecks = &updateChecks
		needsUpdate = true
	}

	// RequiresApprovingReviews
	if ptrNonNilAndNotEqualTo(bp.Spec.RequiresApprovingReviews, ghBp.RequiresApprovingReviews) {
		update.RequiresApprovingReviews = (*githubv4.Boolean)(bp.Spec.RequiresApprovingReviews)
		needsUpdate = true
	}
	// RequiresCodeOwnerReviews
	if ptrNonNilAndNotEqualTo(bp.Spec.RequiresCodeOwnerReviews, ghBp.RequiresCodeOwnerReviews) {
		update.RequiresCodeOwnerReviews = (*githubv4.Boolean)(bp.Spec.RequiresCodeOwnerReviews)
		needsUpdate = true
	}
	// RequiresCommitSignatures
	if ptrNonNilAndNotEqualTo(bp.Spec.RequiresCommitSignatures, ghBp.RequiresCommitSignatures) {
		update.RequiresCommitSignatures = (*githubv4.Boolean)(bp.Spec.RequiresCommitSignatures)
		needsUpdate = true
	}
	// RequiresConversationResolution
	if ptrNonNilAndNotEqualTo(bp.Spec.RequiresConversationResolution, ghBp.RequiresConversationResolution) {
		update.RequiresConversationResolution = (*githubv4.Boolean)(bp.Spec.RequiresConversationResolution)
		needsUpdate = true
	}
	// RequiresDeployments
	if ptrNonNilAndNotEqualTo(bp.Spec.RequiresDeployments, ghBp.RequiresDeployments) {
		update.RequiresDeployments = (*githubv4.Boolean)(bp.Spec.RequiresDeployments)
		needsUpdate = true
	}
	// RequiresLinearHistory
	if ptrNonNilAndNotEqualTo(bp.Spec.RequiresLinearHistory, ghBp.RequiresLinearHistory) {
		update.RequiresLinearHistory = (*githubv4.Boolean)(bp.Spec.RequiresLinearHistory)
		needsUpdate = true
	}
	// RequiresStatusChecks
	if ptrNonNilAndNotEqualTo(bp.Spec.RequiresStatusChecks, ghBp.RequiresStatusChecks) {
		update.RequiresStatusChecks = (*githubv4.Boolean)(bp.Spec.RequiresStatusChecks)
		needsUpdate = true
	}
	// RequiresStrictStatusChecks
	if ptrNonNilAndNotEqualTo(bp.Spec.RequiresStrictStatusChecks, ghBp.RequiresStrictStatusChecks) {
		update.RequiresStrictStatusChecks = (*githubv4.Boolean)(bp.Spec.RequiresStrictStatusChecks)
		needsUpdate = true
	}
	// RestrictsPushes
	if ptrNonNilAndNotEqualTo(bp.Spec.RestrictsPushes, ghBp.RestrictsPushes) {
		update.RestrictsPushes = (*githubv4.Boolean)(bp.Spec.RestrictsPushes)
		needsUpdate = true
	}
	// RestrictsReviewDismissals
	if ptrNonNilAndNotEqualTo(bp.Spec.RestrictsReviewDismissals, ghBp.RestrictsReviewDismissals) {
		update.RestrictsReviewDismissals = (*githubv4.Boolean)(bp.Spec.RestrictsReviewDismissals)
		needsUpdate = true
	}

	specReviewDismissalIds := []githubv4.ID{}
	needsReviewDismissalUpdate := false
	// ReviewDismissalUsers
	userLogins = []string{}
	userIds = []githubv4.ID{}
	for _, user := range ghBp.GetReviewDismissalAllowances().Users {
		userIds = append(userIds, user.Id)
		userLogins = append(userLogins, user.Login)
	}
	specReviewDismissalIds = append(specReviewDismissalIds, userIds...)
	if !cmpSlices(bp.Spec.PushAllowanceUsers, userLogins) {
		needsReviewDismissalUpdate = true
	}
	// ReviewDismissalApps
	appSlugs = []string{}
	appIds = []githubv4.ID{}
	for _, app := range ghBp.GetReviewDismissalAllowances().Apps {
		appIds = append(appIds, app.Id)
		appSlugs = append(appSlugs, app.Slug)
	}
	specReviewDismissalIds = append(specReviewDismissalIds, appIds...)
	if !cmpSlices(bp.Spec.PushAllowanceApps, appSlugs) {
		needsReviewDismissalUpdate = true
	}
	// ReviewDismissalTeams
	teamSlugs = []string{}
	teamIds = []githubv4.ID{}
	for _, team := range ghBp.GetReviewDismissalAllowances().Teams {
		teamIds = append(appIds, team.Id)
		teamSlugs = append(appSlugs, team.Slug)
	}
	specReviewDismissalIds = append(specReviewDismissalIds, teamIds...)
	if !cmpSlices(bp.Spec.PushAllowanceTeams, teamSlugs) {
		needsReviewDismissalUpdate = true
	}
	if needsReviewDismissalUpdate {
		update.ReviewDismissalActorIDs = &specReviewDismissalIds
		needsUpdate = true
	}

	// perform update if necessary
	if needsUpdate || bp.Status.LastUpdateTimestamp == nil {
		log.Info("updating branch protection", "pattern", bp.Spec.Pattern)

		updated, err := r.GitHubClient.UpdateBranchProtection(ctx, &update)
		if err != nil {
			return err
		}

		var ownerLogin string
		if updated.Repository.Owner.Id != "" {
			ownerLogin = updated.Repository.Owner.Login
		} else {
			return fmt.Errorf("repository '%s' has no owner", updated.Repository.Name)
		}

		now := v1.Now()
		reviewCount := int(updated.RequiredApprovingReviewCount)
		bp.Status = githubv1alpha1.BranchProtectionStatus{
			LastUpdateTimestamp:            &now,
			Id:                             &updated.Id,
			RepositoryId:                   &updated.Repository.Id,
			RepositoryOwner:                &ownerLogin,
			RepositoryName:                 &updated.Repository.Name,
			Pattern:                        &updated.Pattern,
			AllowsDeletions:                &updated.AllowsDeletions,
			AllowsForcePushes:              &updated.AllowsForcePushes,
			BlocksCreations:                &updated.BlocksCreations,
			BypassForcePushUsers:           bp.Spec.BypassForcePushUsers,
			BypassForcePushApps:            bp.Spec.BypassForcePushApps,
			BypassForcePushTeams:           bp.Spec.BypassForcePushTeams,
			BypassPullRequestUsers:         bp.Spec.BypassForcePushUsers,
			BypassPullRequestApps:          bp.Spec.BypassForcePushApps,
			BypassPullRequestTeams:         bp.Spec.BypassForcePushTeams,
			DismissesStaleReviews:          &updated.DismissesStaleReviews,
			IsAdminEnforced:                &updated.IsAdminEnforced,
			LockAllowsFetchAndMerge:        &updated.LockAllowsFetchAndMerge,
			LockBranch:                     &updated.LockBranch,
			PushAllowanceUsers:             bp.Spec.PushAllowanceUsers,
			PushAllowanceApps:              bp.Spec.PushAllowanceApps,
			PushAllowanceTeams:             bp.Spec.PushAllowanceTeams,
			RequireLastPushApproval:        &updated.RequireLastPushApproval,
			RequiredApprovingReviewCount:   &reviewCount,
			RequiredDeploymentEnvironments: bp.Spec.RequiredDeploymentEnvironments,
			RequiredStatusCheckContexts:    bp.Spec.RequiredStatusCheckContexts,
			RequiredStatusChecks:           bp.Spec.RequiredStatusChecks,
			RequiresApprovingReviews:       &updated.RequiresApprovingReviews,
			RequiresCodeOwnerReviews:       &updated.RequiresCodeOwnerReviews,
			RequiresCommitSignatures:       &updated.RequiresCommitSignatures,
			RequiresConversationResolution: &updated.RequiresConversationResolution,
			RequiresDeployments:            &updated.RequiresDeployments,
			RequiresLinearHistory:          &updated.RequiresLinearHistory,
			RequiresStatusChecks:           &updated.RequiresStatusChecks,
			RequiresStrictStatusChecks:     &updated.RequiresStrictStatusChecks,
			RestrictsPushes:                &updated.RestrictsPushes,
			RestrictsReviewDismissals:      &updated.RestrictsReviewDismissals,
			ReviewDismissalUsers:           bp.Spec.ReviewDismissalUsers,
			ReviewDismissalApps:            bp.Spec.ReviewDismissalApps,
			ReviewDismissalTeams:           bp.Spec.ReviewDismissalTeams,
		}

		// update status
		if err := r.Status().Update(ctx, bp); err != nil {
			log.Error(err, "error updating BranchProtection status", "pattern", bp.Spec.Pattern)
		}
	}
	return nil
}

func (r *BranchProtectionReconciler) deleteBranchProtection(ctx context.Context, bp *githubv1alpha1.BranchProtection) error {
	if bp.Status.Id == nil {
		return fmt.Errorf("branch protection NodeID is nil")
	}

	return r.GitHubClient.DeleteBranchProtection(ctx, &githubv4.DeleteBranchProtectionRuleInput{
		BranchProtectionRuleID: bp.Status.Id,
	})
}

func (r *BranchProtectionReconciler) branchProtectionToCreateInput(ctx context.Context, bp *githubv1alpha1.BranchProtection) (*githubv4.CreateBranchProtectionRuleInput, error) {
	var id githubv4.ID
	if bp.Status.RepositoryId != nil {
		id = bp.Status.RepositoryId
	} else {
		// return nil, fmt.Errorf("branch protection Status.RepositoryId is nil")
		repo, err := r.GitHubClient.GetRepositoryBySlug(ctx, bp.Spec.RepositoryOwner, bp.Spec.RepositoryName)
		if err != nil {
			// TODO: custom error type
			return nil, fmt.Errorf("no repository '%s' found for owner '%s'", bp.Spec.RepositoryName, bp.Spec.RepositoryOwner)
		}
		id = repo.GetNodeID()
	}

	var requiredApprovingReviewCount *githubv4.Int
	if bp.Spec.RequiredApprovingReviewCount != nil {
		val := (githubv4.Int)(*bp.Spec.RequiredApprovingReviewCount)
		requiredApprovingReviewCount = &val
	}

	return &githubv4.CreateBranchProtectionRuleInput{
		RepositoryID:                   id,
		Pattern:                        githubv4.String(bp.Spec.Pattern),
		RequiresApprovingReviews:       (*githubv4.Boolean)(bp.Spec.RequiresApprovingReviews),
		RequiredApprovingReviewCount:   requiredApprovingReviewCount,
		RequiresCommitSignatures:       (*githubv4.Boolean)(bp.Spec.RequiresCommitSignatures),
		RequiresLinearHistory:          (*githubv4.Boolean)(bp.Spec.RequiresLinearHistory),
		BlocksCreations:                (*githubv4.Boolean)(bp.Spec.BlocksCreations),
		AllowsForcePushes:              (*githubv4.Boolean)(bp.Spec.AllowsForcePushes),
		AllowsDeletions:                (*githubv4.Boolean)(bp.Spec.AllowsDeletions),
		IsAdminEnforced:                (*githubv4.Boolean)(bp.Spec.IsAdminEnforced),
		RequiresStatusChecks:           (*githubv4.Boolean)(bp.Spec.RequiresStatusChecks),
		RequiresStrictStatusChecks:     (*githubv4.Boolean)(bp.Spec.RequiresStatusChecks),
		RequiresCodeOwnerReviews:       (*githubv4.Boolean)(bp.Spec.RequiresCodeOwnerReviews),
		DismissesStaleReviews:          (*githubv4.Boolean)(bp.Spec.DismissesStaleReviews),
		RestrictsReviewDismissals:      (*githubv4.Boolean)(bp.Spec.RestrictsReviewDismissals),
		ReviewDismissalActorIDs:        &[]githubv4.ID{}, // TODO
		BypassPullRequestActorIDs:      &[]githubv4.ID{}, // TODO
		BypassForcePushActorIDs:        &[]githubv4.ID{}, // TODO
		RestrictsPushes:                (*githubv4.Boolean)(bp.Spec.RestrictsPushes),
		PushActorIDs:                   &[]githubv4.ID{},                       // TODO
		RequiredStatusCheckContexts:    &[]githubv4.String{},                   // TODO
		RequiredStatusChecks:           &[]githubv4.RequiredStatusCheckInput{}, // TODO
		RequiresDeployments:            (*githubv4.Boolean)(bp.Spec.RequiresDeployments),
		RequiredDeploymentEnvironments: &[]githubv4.String{}, // TODO
		RequiresConversationResolution: (*githubv4.Boolean)(bp.Spec.RequiresConversationResolution),
		RequireLastPushApproval:        (*githubv4.Boolean)(bp.Spec.RequireLastPushApproval),
		LockBranch:                     (*githubv4.Boolean)(bp.Spec.LockBranch),
		LockAllowsFetchAndMerge:        (*githubv4.Boolean)(bp.Spec.LockAllowsFetchAndMerge),
	}, nil
}
