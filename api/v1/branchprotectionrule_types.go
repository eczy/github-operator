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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BranchProtectionRuleSpec defines the desired state of BranchProtectionRule
type BranchProtectionRuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	//+kubebuilder:validation:MinLength=1

	// The owner of the repository associated with this branch protection rule.
	RepositoryOwner string `json:"repositoryOwner"`

	//+kubebuilder:validation:MinLength=1

	// The repository associated with this branch protection rule.
	RepositoryName string `json:"repositoryName"`

	//+kubebuilder:validation:MinLength=1

	// Identifies the protection rule pattern.
	Pattern string `json:"pattern"`

	// Can this branch be deleted.
	// +optional
	AllowsDeletions *bool `json:"allowsDeletions,omitempty"`

	// Are force pushes allowed on this branch.
	// +optional
	AllowsForcePushes *bool `json:"allowsForcePushes,omitempty"`

	// Is branch creation a protected operation.
	// +optional
	BlocksCreations *bool `json:"blocksCreations,omitempty"`

	// A list of users able to force push for this branch protection rule.
	// +optional
	BypassForcePushUsers []string `json:"bypassForcePushUsers,omitempty"`

	// A list of apps able to force push for this branch protection rule.
	// +optional
	BypassForcePushApps []string `json:"bypassForcePushApps,omitempty"`

	// A list of teams able to force push for this branch protection rule.
	// +optional
	BypassForcePushTeams []string `json:"bypassForcePushTeams,omitempty"`

	// A list of users able to bypass PRs for this branch protection rule.
	// +optional
	BypassPullRequestUsers []string `json:"bypassPullRequestUsers,omitempty"`

	// A list of apps able to bypass PRs for this branch protection rule.
	// +optional
	BypassPullRequestApps []string `json:"bypassPullRequestApps,omitempty"`

	// A list of teams able to bypass PRs for this branch protection rule.
	// +optional
	BypassPullRequestTeams []string `json:"bypassPullRequestTeams,omitempty"`

	// Will new commits pushed to matching branches dismiss pull request review approvals.
	// +optional
	DismissesStaleReviews *bool `json:"dismissesStaleReviews,omitempty"`

	// Can admins override branch protection.
	// +optional
	IsAdminEnforced *bool `json:"isAdminEnforced,omitempty"`

	// Whether users can pull changes from upstream when the branch is locked. Set to true to allow fork syncing. Set to false to prevent fork syncing.
	// +optional
	LockAllowsFetchAndMerge *bool `json:"lockAllowsFetchAndMerge,omitempty"`

	// Whether to set the branch as read-only. If this is true, users will not be able to push to the branch.
	// +optional
	LockBranch *bool `json:"lockBranch,omitempty"`

	// A list of user push allowances for this branch protection rule.
	// +optional
	PushAllowanceUsers []string `json:"pushAllowanceUsers,omitempty"`

	// A list of app push allowances for this branch protection rule.
	// +optional
	PushAllowanceApps []string `json:"pushAllowanceApps,omitempty"`

	// A list of team push allowances for this branch protection rule.
	// +optional
	PushAllowanceTeams []string `json:"pushAllowanceTeams,omitempty"`

	// Whether the most recent push must be approved by someone other than the person who pushed it.
	// +optional
	RequireLastPushApproval *bool `json:"requireLastPushApproval,omitempty"`

	// Number of approving reviews required to update matching branches.
	// +optional
	RequiredApprovingReviewCount *int `json:"requiredApprovingReviewCount,omitempty"`

	// List of required deployment environments that must be deployed successfully to update matching branches.
	// +optional
	RequiredDeploymentEnvironments []string `json:"requiredDeploymentEnvironments,omitempty"`

	// List of required status check contexts that must pass for commits to be accepted to matching branches.
	// +optional
	RequiredStatusCheckContexts []string `json:"requiredStatusCheckContexts,omitempty"`

	// List of required status checks that must pass for commits to be accepted to matching branches.
	// +optional
	RequiredStatusChecks []RequiredStatusCheck `json:"requiredStatusChecks,omitempty"`

	// Are approving reviews required to update matching branches.
	// +optional
	RequiresApprovingReviews *bool `json:"requiresApprovingReviews,omitempty"`

	// Are reviews from code owners required to update matching branches.
	// +optional
	RequiresCodeOwnerReviews *bool `json:"requiresCodeOwnerReviews,omitempty"`

	// Are commits required to be signed.
	// +optional
	RequiresCommitSignatures *bool `json:"requiresCommitSignatures,omitempty"`

	// Are conversations required to be resolved before merging.
	// +optional
	RequiresConversationResolution *bool `json:"requiresConversationResolution,omitempty"`

	// Does this branch require deployment to specific environments before merging.
	// +optional
	RequiresDeployments *bool `json:"requiresDeployments,omitempty"`

	// Are merge commits prohibited from being pushed to this branch.
	// +optional
	RequiresLinearHistory *bool `json:"requiresLinearHistory,omitempty"`

	// Are status checks required to update matching branches.
	// +optional
	RequiresStatusChecks *bool `json:"requiresStatusChecks,omitempty"`

	// Are branches required to be up to date before merging.
	// +optional
	RequiresStrictStatusChecks *bool `json:"requiresStrictStatusChecks,omitempty"`

	// Is pushing to matching branches restricted.
	// +optional
	RestrictsPushes *bool `json:"restrictsPushes,omitempty"`

	// Is dismissal of pull request reviews restricted.
	// +optional
	RestrictsReviewDismissals *bool `json:"restrictsReviewDismissals,omitempty"`

	// A list of user review dismissal allowances for this branch protection rule.
	// +optional
	ReviewDismissalUsers []string `json:"reviewDismissalUsers,omitempty"`

	// A list of app review dismissal allowances for this branch protection rule.
	// +optional
	ReviewDismissalApps []string `json:"reviewDismissalApps,omitempty"`

	// A list of team review dismissal allowances for this branch protection rule.
	// +optional
	ReviewDismissalTeams []string `json:"reviewDismissalTeams,omitempty"`
}

// BranchProtectionRuleStatus defines the observed state of BranchProtectionRule.
type BranchProtectionRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LastUpdateTimestamp *metav1.Time `json:"lastUpdateTimestamp,omitempty"`

	NodeId                         *string               `json:"nodeId,omitempty"`
	RepositoryNodeId               *string               `json:"repositoryNodeId,omitempty"`
	RepositoryOwner                *string               `json:"repositoryOwner,omitempty"`
	RepositoryName                 *string               `json:"repositoryName,omitempty"`
	Pattern                        *string               `json:"branchPattern,omitempty"`
	AllowsDeletions                *bool                 `json:"allowsDeletions,omitempty"`
	AllowsForcePushes              *bool                 `json:"allowsForcePushes,omitempty"`
	BlocksCreations                *bool                 `json:"blocksCreations,omitempty"`
	BypassForcePushUsers           []string              `json:"bypassForcePushUsers,omitempty"`
	BypassForcePushApps            []string              `json:"bypassForcePushApps,omitempty"`
	BypassForcePushTeams           []string              `json:"bypassForcePushteams,omitempty"`
	BypassPullRequestUsers         []string              `json:"bypassPullRequestUsers,omitempty"`
	BypassPullRequestApps          []string              `json:"bypassPullRequestApps,omitempty"`
	BypassPullRequestTeams         []string              `json:"bypassPullRequestTeams,omitempty"`
	DismissesStaleReviews          *bool                 `json:"dismissesStaleReviews,omitempty"`
	IsAdminEnforced                *bool                 `json:"isAdminEnforced,omitempty"`
	LockAllowsFetchAndMerge        *bool                 `json:"lockAllowsFetchAndMerge,omitempty"`
	LockBranch                     *bool                 `json:"lockBranch,omitempty"`
	PushAllowanceUsers             []string              `json:"pushAllowanceUsers,omitempty"`
	PushAllowanceApps              []string              `json:"pushAllowanceApps,omitempty"`
	PushAllowanceTeams             []string              `json:"pushAllowanceTeams,omitempty"`
	RequireLastPushApproval        *bool                 `json:"requireLastPushApproval,omitempty"`
	RequiredApprovingReviewCount   *int                  `json:"requiredApprovingReviewCount,omitempty"`
	RequiredDeploymentEnvironments []string              `json:"requiredDeploymentEnvironments,omitempty"`
	RequiredStatusCheckContexts    []string              `json:"requiredStatusCheckContexts,omitempty"`
	RequiredStatusChecks           []RequiredStatusCheck `json:"requiredStatusChecks,omitempty"`
	RequiresApprovingReviews       *bool                 `json:"requiresApprovingReviews,omitempty"`
	RequiresCodeOwnerReviews       *bool                 `json:"requiresCodeOwnerReviews,omitempty"`
	RequiresCommitSignatures       *bool                 `json:"requiresCommitSignatures,omitempty"`
	RequiresConversationResolution *bool                 `json:"requiresConversationResolution,omitempty"`
	RequiresDeployments            *bool                 `json:"requiresDeployments,omitempty"`
	RequiresLinearHistory          *bool                 `json:"requiresLinearHistory,omitempty"`
	RequiresStatusChecks           *bool                 `json:"requiresStatusChecks,omitempty"`
	RequiresStrictStatusChecks     *bool                 `json:"requiresStrictStatusChecks,omitempty"`
	RestrictsPushes                *bool                 `json:"restrictsPushes,omitempty"`
	RestrictsReviewDismissals      *bool                 `json:"restrictsReviewDismissals,omitempty"`
	ReviewDismissalUsers           []string              `json:"reviewDismissalUsers,omitempty"`
	ReviewDismissalApps            []string              `json:"reviewDismissalApps,omitempty"`
	ReviewDismissalTeams           []string              `json:"reviewDismissalTeams,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// BranchProtectionRule is the Schema for the branchprotectionrules API
type BranchProtectionRule struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of BranchProtectionRule
	// +required
	Spec BranchProtectionRuleSpec `json:"spec"`

	// status defines the observed state of BranchProtectionRule
	// +optional
	Status BranchProtectionRuleStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// BranchProtectionRuleList contains a list of BranchProtectionRule
type BranchProtectionRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BranchProtectionRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BranchProtectionRule{}, &BranchProtectionRuleList{})
}

type RequiredStatusCheck struct {
	AppId   *string `json:"appId,omitempty"`
	Context string  `json:"context"`
}
