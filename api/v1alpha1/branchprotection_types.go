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

	RepositoryOwner string `json:"repository_owner"`
	RepositoryName  string `json:"repository_name"`
	Pattern         string `json:"branch_pattern"`

	AllowsDeletions   *bool `json:"allows_deletions,omitempty"`
	AllowsForcePushes *bool `json:"allows_force_pushes,omitempty"`
	BlocksCreations   *bool `json:"blocks_creations,omitempty"`

	BypassForcePushUsers []string `json:"bypass_force_push_users,omitempty"`
	BypassForcePushApps  []string `json:"bypass_force_push_apps,omitempty"`
	BypassForcePushTeams []string `json:"bypass_force_push_teams,omitempty"`

	BypassPullRequestUsers []string `json:"bypass_pull_request_users,omitempty"`
	BypassPullRequestApps  []string `json:"bypass_pull_request_apps,omitempty"`
	BypassPullRequestTeams []string `json:"bypass_pull_request_teams,omitempty"`

	DismissesStaleReviews   *bool `json:"dismisses_stale_reviews,omitempty"`
	IsAdminEnforced         *bool `json:"is_admin_enforced,omitempty"`
	LockAllowsFetchAndMerge *bool `json:"lock_allows_fetch_and_merge,omitempty"`
	LockBranch              *bool `json:"lock_branch,omitempty"`

	PushAllowanceUsers []string `json:"push_allowance_users,omitempty"`
	PushAllowanceApps  []string `json:"push_allowance_apps,omitempty"`
	PushAllowanceTeams []string `json:"push_allowance_teams,omitempty"`

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

	ReviewDismissalUsers []string `json:"review_dismissal_users,omitempty"`
	ReviewDismissalApps  []string `json:"review_dismissal_apps,omitempty"`
	ReviewDismissalTeams []string `json:"review_dismissal_teams,omitempty"`
}

// BranchProtectionStatus defines the observed state of BranchProtection
type BranchProtectionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LastUpdateTimestamp *metav1.Time `json:"last_update_timestamp,omitempty"`

	Id              *string `json:"id,omitempty"`
	RepositoryId    *string `json:"repository_id,omitempty"` // node id not database id
	RepositoryOwner *string `json:"repository_owner,omitempty"`
	RepositoryName  *string `json:"repository_name,omitempty"`
	Pattern         *string `json:"branch_pattern,omitempty"`

	AllowsDeletions   *bool `json:"allows_deletions,omitempty"`
	AllowsForcePushes *bool `json:"allows_force_pushes,omitempty"`
	BlocksCreations   *bool `json:"blocks_creations,omitempty"`

	BypassForcePushUsers []string `json:"bypass_force_push_users,omitempty"`
	BypassForcePushApps  []string `json:"bypass_force_push_apps,omitempty"`
	BypassForcePushTeams []string `json:"bypass_force_push_teams,omitempty"`

	BypassPullRequestUsers []string `json:"bypass_pull_request_users,omitempty"`
	BypassPullRequestApps  []string `json:"bypass_pull_request_apps,omitempty"`
	BypassPullRequestTeams []string `json:"bypass_pull_request_teams,omitempty"`

	DismissesStaleReviews   *bool `json:"dismisses_stale_reviews,omitempty"`
	IsAdminEnforced         *bool `json:"is_admin_enforced,omitempty"`
	LockAllowsFetchAndMerge *bool `json:"lock_allows_fetch_and_merge,omitempty"`
	LockBranch              *bool `json:"lock_branch,omitempty"`

	PushAllowanceUsers []string `json:"push_allowance_users,omitempty"`
	PushAllowanceApps  []string `json:"push_allowance_apps,omitempty"`
	PushAllowanceTeams []string `json:"push_allowance_teams,omitempty"`

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

	ReviewDismissalUsers []string `json:"review_dismissal_users,omitempty"`
	ReviewDismissalApps  []string `json:"review_dismissal_apps,omitempty"`
	ReviewDismissalTeams []string `json:"review_dismissal_teams,omitempty"`
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
