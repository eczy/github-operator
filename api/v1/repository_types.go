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

// RepositorySpec defines the desired state of Repository
type RepositorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	//+kubebuilder:validation:MinLength=1
	// The name of the repository.
	Name string `json:"name"`

	//+kubebuilder:validation:MinLength=1
	// The organization name. The name is not case sensitive.
	Owner string `json:"owner"`

	// Repository description.
	// +optional
	Description *string `json:"description,omitempty"`

	// A URL with more information about the repository.
	// +optional
	Homepage *string `json:"homepage,omitempty"`

	// The default branch for this repository.
	// +optional
	DefaultBranch *string `json:"defaultBranch,omitempty"`

	// The account owner of the template repository. The name is not case sensitive.
	// +optional
	TemplateOwner *string `json:"templateOwner,omitempty"`

	// The name of the template repository without the .git extension. The name is not case sensitive.
	// +optional
	TemplateRepository *string `json:"templateRepository,omitempty"`

	// Either true to allow rebase-merging pull requests, or false to prevent rebase-merging.
	// Default: true
	// +optional
	AllowRebaseMerge *bool `json:"allowRebaseMerge,omitempty"`

	// Either true to always allow a pull request head branch that is behind its base branch to be updated even if it is not required to be up to date before merging, or false otherwise.
	// Default: false
	// +optional
	AllowUpdateBranch *bool `json:"allowUpdateBranch,omitempty"`

	//Either true to allow squash-merging pull requests, or false to prevent squash-merging. Default: true.
	// +optional
	AllowSquashMerge *bool `json:"allowSquashMerge,omitempty"`

	// Either true to allow merging pull requests with a merge commit, or false to prevent merging pull requests with merge commits. Default: true.
	// +optional
	AllowMergeCommit *bool `json:"allowMergeCommit,omitempty"`

	// Either true to allow auto-merge on pull requests, or false to disallow auto-merge. Default: false.
	// +optional
	AllowAutoMerge *bool `json:"allowAutoMerge,omitempty"`

	// Either true to allow private forks, or false to prevent private forks.
	// Default: false
	// +optional
	AllowForking *bool `json:"allowForking,omitempty"`

	// Either true to require contributors to sign off on web-based commits, or false to not require contributors to sign off on web-based commits.
	// Default: false
	// +optional
	WebCommitSignoffRequired *bool `json:"webCommitSignoffRequired,omitempty"`

	// Either true to allow automatically deleting head branches when pull requests are merged, or false to prevent automatic deletion. Default: false.
	// +optional
	DeleteBranchOnMerge *bool `json:"deleteBranchOnMerge,omitempty"`

	// The default value for a squash merge commit title:
	//   - PR_TITLE - default to the pull request's title.
	//   - COMMIT_OR_PR_TITLE - default to the commit's title (if only one commit) or the pull request's title (when more than one commit).
	// Can be one of: PR_TITLE, COMMIT_OR_PR_TITLE
	// +optional
	SquashMergeCommitTitle *SquashMergeCommitTitle `json:"squashMergeCommitTitle,omitempty"`

	// The default value for a squash merge commit message:
	//   - PR_BODY - default to the pull request's body.
	//   - COMMIT_MESSAGES - default to the branch's commit messages.
	//   - BLANK - default to a blank commit message.
	// Can be one of: PR_BODY, COMMIT_MESSAGES, BLANK
	// +optional
	SquashMergeCommitMessage *SquashMergeCommitMessage `json:"squashMergeCommitMessage,omitempty"`

	// The default value for a merge commit title.
	//   - PR_TITLE - default to the pull request's title.
	//   - MERGE_MESSAGE - default to the classic title for a merge message (e.g., Merge pull request #123 from branch-name).
	// Can be one of: PR_TITLE, MERGE_MESSAGE
	// +optional
	MergeCommitTitle *MergeCommitTitle `json:"mergeCommitTitle,omitempty"`

	// The default value for a merge commit message.
	//   - PR_TITLE - default to the pull request's title.
	//   - PR_BODY - default to the pull request's body.
	//   - BLANK - default to a blank commit message.
	// Can be one of: PR_BODY, PR_TITLE, BLANK
	// +optional
	MergeCommitMessage *MergeCommitMessage `json:"mergeCommitMessage,omitempty"`

	// Set of topics with which the repository will be associated.
	// +optional
	Topics []string `json:"topics,omitempty"`

	// Whether to archive this repository. false will unarchive a previously archived repository.
	// Default: false
	// +optional
	Archived *bool `json:"archived,omitempty"`

	// Either true to enable issues for this repository or false to disable them.
	// Default: true
	// +optional
	HasIssues *bool `json:"hasIssues,omitempty"`

	// Whether the wiki is enabled.
	// Default: true
	// +optional
	HasWiki *bool `json:"hasWiki,omitempty"`

	// Either true to enable projects for this repository or false to disable them. Note: If you're creating a repository in an organization that has disabled repository projects, the default is false, and if you pass true, the API returns an error.
	// Default: true
	// +optional
	HasProjects *bool `json:"hasProjects,omitempty"`

	// Whether downloads are enabled.
	// Default: true
	// +optional
	HasDownloads *bool `json:"hasDownloads,omitempty"`

	// Whether discussions are enabled.
	// Default: false
	// +optional
	HasDiscussions *bool `json:"hasDiscussions,omitempty"`

	// The visibility of the repository. Can be one of: public, private, internal.
	// +optional
	Visibility *string `json:"visibility,omitempty"`

	// Specify which security and analysis features to enable or disable for the repository.
	//
	// To use this parameter, you must have admin permissions for the repository or be an owner or security manager for the organization that owns the repository. For more information, see [Managing security managers in your organization].
	//
	// [Managing security managers in your organization]: https://docs.github.com/en/organizations/managing-peoples-access-to-your-organization-with-roles/managing-security-managers-in-your-organization
	// +optional
	SecurityAndAnalysis *SecurityAndAnalysis `json:"securitAandAnalysis,omitempty"`
}

// RepositoryStatus defines the observed state of Repository.
type RepositoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	LastUpdateTimestamp      *metav1.Time              `json:"lastUpdateTimestamp,omitempty"`
	Id                       *int64                    `json:"id,omitempty"`
	NodeId                   *string                   `json:"nodeId,omitempty"`
	OwnerLogin               *string                   `json:"ownerLogin,omitempty"`
	OwnerNodeId              *int64                    `json:"ownerNodeId,omitempty"`
	Name                     *string                   `json:"name,omitempty"`
	FullName                 *string                   `json:"fullName,omitempty"`
	Owner                    *string                   `json:"owner,omitempty"`
	Description              *string                   `json:"description,omitempty"`
	Homepage                 *string                   `json:"homepage,omitempty"`
	DefaultBranch            *string                   `json:"defaultBranch,omitempty"`
	TemplateOwner            *string                   `json:"templateOwner,omitempty"`
	TemplateRepository       *string                   `json:"templateRepository,omitempty"`
	AllowRebaseMerge         *bool                     `json:"allowRebaseMerge,omitempty"`
	AllowUpdateBranch        *bool                     `json:"allowUpdateBranch,omitempty"`
	AllowSquashMerge         *bool                     `json:"allowSquashMerge,omitempty"`
	AllowMergeCommit         *bool                     `json:"allowMergeCommit,omitempty"`
	AllowAutoMerge           *bool                     `json:"allowAutoMerge,omitempty"`
	AllowForking             *bool                     `json:"allowForking,omitempty"`
	WebCommitSignoffRequired *bool                     `json:"webCommitSignoffRequired,omitempty"`
	DeleteBranchOnMerge      *bool                     `json:"deleteBranchOnMerge,omitempty"`
	SquashMergeCommitTitle   *SquashMergeCommitTitle   `json:"squashMergeCommitTitle,omitempty"`
	SquashMergeCommitMessage *SquashMergeCommitMessage `json:"squashMergeCommitMessage,omitempty"`
	MergeCommitTitle         *MergeCommitTitle         `json:"mergeCommitTitle,omitempty"`
	MergeCommitMessage       *MergeCommitMessage       `json:"mergeCommitMessage,omitempty"`
	Topics                   []string                  `json:"topics,omitempty"`
	Archived                 *bool                     `json:"archived,omitempty"`
	HasIssues                *bool                     `json:"hasIssues,omitempty"`
	HasWiki                  *bool                     `json:"hasWiki,omitempty"`
	HasProjects              *bool                     `json:"hasProjects,omitempty"`
	HasDownloads             *bool                     `json:"hasDownloads,omitempty"`
	HasDiscussions           *bool                     `json:"hasDiscussions,omitempty"`
	Visibility               *string                   `json:"visibility,omitempty"`
	SecurityAndAnalysis      *SecurityAndAnalysis      `json:"securityAndAnalysis,omitempty"`

	ParentName                    *string `json:"parentName,omitempty"`
	ParentId                      *int64  `json:"parentId,omitempty"`
	ParentNodeId                  *string `json:"parentNodeId,omitempty"`
	TemplateRepositoryOwnerLogin  *string `json:"templateRepositoryOwnerLogin,omitempty"`
	TemplateRepositoryOwnerNodeId *string `json:"templateRepositoryOwnerNodeId,omitempty"`
	TemplateRepositoryName        *string `json:"templateRepositoryName,omitempty"`
	TemplateRepositoryId          *int64  `json:"templateRepositoryId,omitempty"`
	OrganizationLogin             *string `json:"organizationLogin,omitempty"`
	OrganizationId                *int64  `json:"organizationId,omitempty"`

	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
	PushedAt  *metav1.Time `json:"pushedAt,omitempty"`
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Repository is the Schema for the repositories API
type Repository struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Repository
	// +required
	Spec RepositorySpec `json:"spec"`

	// status defines the observed state of Repository
	// +optional
	Status RepositoryStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// RepositoryList contains a list of Repository
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Repository{}, &RepositoryList{})
}

type SecurityAndAnalysisFeature struct {
	// Can be enabled or disabled.
	Status string `json:"status"`
}

type SecurityAndAnalysis struct {
	// Use the status property to enable or disable GitHub Advanced Security for this repository. For more information, see [About GitHub Advanced Security].
	//
	// [About GitHub Advanced Security]: https://docs.github.com/en/get-started/learning-about-github/about-github-advanced-security
	AdvancedSecurity SecurityAndAnalysisFeature `json:"advancedSecurity"`
	// Use the status property to enable or disable secret scanning for this repository. For more information, see [About secret scanning].
	//
	// [About secret scanning]: https://docs.github.com/en/code-security/secret-scanning/about-secret-scanning
	SecretScanning SecurityAndAnalysisFeature `json:"secretScanning"`
	// Use the status property to enable or disable secret scanning push protection for this repository. For more information, see [Protecting pushes with secret scanning].
	//
	// [Protecting pushes with secret scanning]: https://docs.github.com/en/code-security/secret-scanning/push-protection-for-repositories-and-organizations
	SecretScanningPushProtection SecurityAndAnalysisFeature `json:"secretScanningPushProtection"`
}

// +kubebuilder:validation:Enum=PR_TITLE;COMMIT_OR_PR_TITLE
type SquashMergeCommitTitle string

const (
	SquashMergeCommitTitlePrTitle         SquashMergeCommitTitle = "PR_TITLE"
	SquashMergeCommitTitleCommitOrPrTitle SquashMergeCommitTitle = "COMMIT_OR_PR_TITLE"
)

// +kubebuilder:validation:Enum=PR_BODY;COMMIT_MESSAGES;BLANK
type SquashMergeCommitMessage string

const (
	SquashMergeCommitMessagePrBody         SquashMergeCommitMessage = "PR_BODY"
	SquashMergeCommitMessageCommitMessages SquashMergeCommitMessage = "COMMIT_MESSAGES"
	SquashMergeCommitMessageBlank          SquashMergeCommitMessage = "BLANK"
)

// +kubebuilder:validation:Enum=PR_TITLE;MERGE_MESSAGE
type MergeCommitTitle string

const (
	MergeCommitTitlePrTitle      MergeCommitTitle = "PR_TITLE"
	MergeCommitTitleMergeMessage MergeCommitTitle = "MERGE_MESSAGE"
)

// +kubebuilder:validation:Enum=PR_BODY;PR_TITLE;BLANK
type MergeCommitMessage string

const (
	MergeCommitMessagePrBody  MergeCommitMessage = "PR_BODY"
	MergeCommitMessagePrTitle MergeCommitMessage = "PR_TITLE"
	MergeCommitMessageBlank   MergeCommitMessage = "BLANK"
)
