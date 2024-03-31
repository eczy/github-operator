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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BranchProtectionSpec defines the desired state of BranchProtection
type BranchProtectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:MinLength=1

	// The owner of the repository associated with this branch protection rule.
	RepositoryOwner string `json:"repository_owner"`

	//+kubebuilder:validation:MinLength=1

	// The repository associated with this branch protection rule.
	RepositoryName string `json:"repository_name"`

	//+kubebuilder:validation:MinLength=1

	// Identifies the protection rule pattern.
	Pattern string `json:"branch_pattern"`

	// Can this branch be deleted.
	// +optional
	AllowsDeletions *bool `json:"allows_deletions,omitempty"`

	// Are force pushes allowed on this branch.
	// +optional
	AllowsForcePushes *bool `json:"allows_force_pushes,omitempty"`

	// Is branch creation a protected operation.
	// +optional
	BlocksCreations *bool `json:"blocks_creations,omitempty"`

	// A list of users able to force push for this branch protection rule.
	// +optional
	BypassForcePushUsers []string `json:"bypass_force_push_users,omitempty"`

	// A list of apps able to force push for this branch protection rule.
	// +optional
	BypassForcePushApps []string `json:"bypass_force_push_apps,omitempty"`

	// A list of teams able to force push for this branch protection rule.
	// +optional
	BypassForcePushTeams []string `json:"bypass_force_push_teams,omitempty"`

	// A list of users able to bypass PRs for this branch protection rule.
	// +optional
	BypassPullRequestUsers []string `json:"bypass_pull_request_users,omitempty"`

	// A list of apps able to bypass PRs for this branch protection rule.
	// +optional
	BypassPullRequestApps []string `json:"bypass_pull_request_apps,omitempty"`

	// A list of teams able to bypass PRs for this branch protection rule.
	// +optional
	BypassPullRequestTeams []string `json:"bypass_pull_request_teams,omitempty"`

	// Will new commits pushed to matching branches dismiss pull request review approvals.
	// +optional
	DismissesStaleReviews *bool `json:"dismisses_stale_reviews,omitempty"`

	// Can admins override branch protection.
	// +optional
	IsAdminEnforced *bool `json:"is_admin_enforced,omitempty"`

	// Whether users can pull changes from upstream when the branch is locked. Set to true to allow fork syncing. Set to false to prevent fork syncing.
	// +optional
	LockAllowsFetchAndMerge *bool `json:"lock_allows_fetch_and_merge,omitempty"`

	// Whether to set the branch as read-only. If this is true, users will not be able to push to the branch.
	// +optional
	LockBranch *bool `json:"lock_branch,omitempty"`

	// A list of user push allowances for this branch protection rule.
	// +optional
	PushAllowanceUsers []string `json:"push_allowance_users,omitempty"`

	// A list of app push allowances for this branch protection rule.
	// +optional
	PushAllowanceApps []string `json:"push_allowance_apps,omitempty"`

	// A list of team push allowances for this branch protection rule.
	// +optional
	PushAllowanceTeams []string `json:"push_allowance_teams,omitempty"`

	// Whether the most recent push must be approved by someone other than the person who pushed it.
	// +optional
	RequireLastPushApproval *bool `json:"require_last_push_approval,omitempty"`

	// Number of approving reviews required to update matching branches.
	// +optional
	RequiredApprovingReviewCount *int `json:"required_approving_review_count,omitempty"`

	// List of required deployment environments that must be deployed successfully to update matching branches.
	// +optional
	RequiredDeploymentEnvironments []string `json:"required_deployment_environments,omitempty"`

	// List of required status check contexts that must pass for commits to be accepted to matching branches.
	// +optional
	RequiredStatusCheckContexts []string `json:"required_status_check_contexts,omitempty"`

	// List of required status checks that must pass for commits to be accepted to matching branches.
	// +optional
	RequiredStatusChecks []RequiredStatusCheck `json:"required_status_checks,omitempty"`

	// Are approving reviews required to update matching branches.
	// +optional
	RequiresApprovingReviews *bool `json:"requires_approving_reviews,omitempty"`

	// Are reviews from code owners required to update matching branches.
	// +optional
	RequiresCodeOwnerReviews *bool `json:"requires_code_owner_reviews,omitempty"`

	// Are commits required to be signed.
	// +optional
	RequiresCommitSignatures *bool `json:"requires_commit_signatures,omitempty"`

	// Are conversations required to be resolved before merging.
	// +optional
	RequiresConversationResolution *bool `json:"requires_conversation_resolution,omitempty"`

	// Does this branch require deployment to specific environments before merging.
	// +optional
	RequiresDeployments *bool `json:"requires_deployments,omitempty"`

	// Are merge commits prohibited from being pushed to this branch.
	// +optional
	RequiresLinearHistory *bool `json:"requires_linear_history,omitempty"`

	// Are status checks required to update matching branches.
	// +optional
	RequiresStatusChecks *bool `json:"requires_status_checks,omitempty"`

	// Are branches required to be up to date before merging.
	// +optional
	RequiresStrictStatusChecks *bool `json:"requires_strict_status_checks,omitempty"`

	// Is pushing to matching branches restricted.
	// +optional
	RestrictsPushes *bool `json:"restricts_pushes,omitempty"`

	// Is dismissal of pull request reviews restricted.
	// +optional
	RestrictsReviewDismissals *bool `json:"restricts_review_dismissals,omitempty"`

	// A list of user review dismissal allowances for this branch protection rule.
	// +optional
	ReviewDismissalUsers []string `json:"review_dismissal_users,omitempty"`

	// A list of app review dismissal allowances for this branch protection rule.
	// +optional
	ReviewDismissalApps []string `json:"review_dismissal_apps,omitempty"`

	// A list of team review dismissal allowances for this branch protection rule.
	// +optional
	ReviewDismissalTeams []string `json:"review_dismissal_teams,omitempty"`
}

// BranchProtectionStatus defines the observed state of BranchProtection
type BranchProtectionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LastUpdateTimestamp *metav1.Time `json:"last_update_timestamp,omitempty"`

	NodeId                         *string               `json:"node_id,omitempty"`
	RepositoryNodeId               *string               `json:"repository_node_id,omitempty"` // node id not database id
	RepositoryOwner                *string               `json:"repository_owner,omitempty"`
	RepositoryName                 *string               `json:"repository_name,omitempty"`
	Pattern                        *string               `json:"branch_pattern,omitempty"`
	AllowsDeletions                *bool                 `json:"allows_deletions,omitempty"`
	AllowsForcePushes              *bool                 `json:"allows_force_pushes,omitempty"`
	BlocksCreations                *bool                 `json:"blocks_creations,omitempty"`
	BypassForcePushUsers           []string              `json:"bypass_force_push_users,omitempty"`
	BypassForcePushApps            []string              `json:"bypass_force_push_apps,omitempty"`
	BypassForcePushTeams           []string              `json:"bypass_force_push_teams,omitempty"`
	BypassPullRequestUsers         []string              `json:"bypass_pull_request_users,omitempty"`
	BypassPullRequestApps          []string              `json:"bypass_pull_request_apps,omitempty"`
	BypassPullRequestTeams         []string              `json:"bypass_pull_request_teams,omitempty"`
	DismissesStaleReviews          *bool                 `json:"dismisses_stale_reviews,omitempty"`
	IsAdminEnforced                *bool                 `json:"is_admin_enforced,omitempty"`
	LockAllowsFetchAndMerge        *bool                 `json:"lock_allows_fetch_and_merge,omitempty"`
	LockBranch                     *bool                 `json:"lock_branch,omitempty"`
	PushAllowanceUsers             []string              `json:"push_allowance_users,omitempty"`
	PushAllowanceApps              []string              `json:"push_allowance_apps,omitempty"`
	PushAllowanceTeams             []string              `json:"push_allowance_teams,omitempty"`
	RequireLastPushApproval        *bool                 `json:"require_last_push_approval,omitempty"`
	RequiredApprovingReviewCount   *int                  `json:"required_approving_review_count,omitempty"`
	RequiredDeploymentEnvironments []string              `json:"required_deployment_environments,omitempty"`
	RequiredStatusCheckContexts    []string              `json:"required_status_check_contexts,omitempty"`
	RequiredStatusChecks           []RequiredStatusCheck `json:"required_status_checks,omitempty"`
	RequiresApprovingReviews       *bool                 `json:"requires_approving_reviews,omitempty"`
	RequiresCodeOwnerReviews       *bool                 `json:"requires_code_owner_reviews,omitempty"`
	RequiresCommitSignatures       *bool                 `json:"requires_commit_signatures,omitempty"`
	RequiresConversationResolution *bool                 `json:"requires_conversation_resolution,omitempty"`
	RequiresDeployments            *bool                 `json:"requires_deployments,omitempty"`
	RequiresLinearHistory          *bool                 `json:"requires_linear_history,omitempty"`
	RequiresStatusChecks           *bool                 `json:"requires_status_checks,omitempty"`
	RequiresStrictStatusChecks     *bool                 `json:"requires_strict_status_checks,omitempty"`
	RestrictsPushes                *bool                 `json:"restricts_pushes,omitempty"`
	RestrictsReviewDismissals      *bool                 `json:"restricts_review_dismissals,omitempty"`
	ReviewDismissalUsers           []string              `json:"review_dismissal_users,omitempty"`
	ReviewDismissalApps            []string              `json:"review_dismissal_apps,omitempty"`
	ReviewDismissalTeams           []string              `json:"review_dismissal_teams,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BranchProtection is the Schema for the branchprotections API
type BranchProtection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BranchProtectionSpec   `json:"spec,omitempty"`
	Status BranchProtectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BranchProtectionList contains a list of BranchProtection
type BranchProtectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BranchProtection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BranchProtection{}, &BranchProtectionList{})
}

type RequiredStatusCheck struct {
	AppId   *string `json:"app_id,omitempty"`
	Context string  `json:"context"`
}
