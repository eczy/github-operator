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

// RepositorySpec defines the desired state of Repository
type RepositorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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
	DefaultBranch *string `json:"default_branch,omitempty"`

	// The account owner of the template repository. The name is not case sensitive.
	// +optional
	TemplateOwner *string `json:"template_owner,omitempty"`

	// The name of the template repository without the .git extension. The name is not case sensitive.
	// +optional
	TemplateRepository *string `json:"template_repository,omitempty"`

	// Either true to allow rebase-merging pull requests, or false to prevent rebase-merging.
	// Default: true
	// +optional
	AllowRebaseMerge *bool `json:"allow_rebase_merge,omitempty"`

	// Either true to always allow a pull request head branch that is behind its base branch to be updated even if it is not required to be up to date before merging, or false otherwise.
	// Default: false
	// +optional
	AllowUpdateBranch *bool `json:"allow_update_branch,omitempty"`

	//Either true to allow squash-merging pull requests, or false to prevent squash-merging. Default: true.
	// +optional
	AllowSquashMerge *bool `json:"allow_squash_merge,omitempty"`

	// Either true to allow merging pull requests with a merge commit, or false to prevent merging pull requests with merge commits. Default: true.
	// +optional
	AllowMergeCommit *bool `json:"allow_merge_commit,omitempty"`

	// Either true to allow auto-merge on pull requests, or false to disallow auto-merge. Default: false.
	// +optional
	AllowAutoMerge *bool `json:"allow_auto_merge,omitempty"`

	// Either true to allow private forks, or false to prevent private forks.
	// Default: false
	// +optional
	AllowForking *bool `json:"allow_forking,omitempty"`

	// Either true to require contributors to sign off on web-based commits, or false to not require contributors to sign off on web-based commits.
	// Default: false
	// +optional
	WebCommitSignoffRequired *bool `json:"web_commit_signoff_required,omitempty"`

	// Either true to allow automatically deleting head branches when pull requests are merged, or false to prevent automatic deletion. Default: false.
	// +optional
	DeleteBranchOnMerge *bool `json:"delete_branch_on_merge,omitempty"`

	// The default value for a squash merge commit title:
	//   - PR_TITLE - default to the pull request's title.
	//   - COMMIT_OR_PR_TITLE - default to the commit's title (if only one commit) or the pull request's title (when more than one commit).
	// Can be one of: PR_TITLE, COMMIT_OR_PR_TITLE
	// +optional
	SquashMergeCommitTitle *SquashMergeCommitTitle `json:"squash_merge_commit_title,omitempty"`

	// The default value for a squash merge commit message:
	//   - PR_BODY - default to the pull request's body.
	//   - COMMIT_MESSAGES - default to the branch's commit messages.
	//   - BLANK - default to a blank commit message.
	// Can be one of: PR_BODY, COMMIT_MESSAGES, BLANK
	// +optional
	SquashMergeCommitMessage *SquashMergeCommitMessage `json:"squash_merge_commit_message,omitempty"`

	// The default value for a merge commit title.
	//   - PR_TITLE - default to the pull request's title.
	//   - MERGE_MESSAGE - default to the classic title for a merge message (e.g., Merge pull request #123 from branch-name).
	// Can be one of: PR_TITLE, MERGE_MESSAGE
	// +optional
	MergeCommitTitle *MergeCommitTitle `json:"merge_commit_title,omitempty"`

	// The default value for a merge commit message.
	//   - PR_TITLE - default to the pull request's title.
	//   - PR_BODY - default to the pull request's body.
	//   - BLANK - default to a blank commit message.
	// Can be one of: PR_BODY, PR_TITLE, BLANK
	// +optional
	MergeCommitMessage *MergeCommitMessage `json:"merge_commit_message,omitempty"`

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
	HasIssues *bool `json:"has_issues,omitempty"`

	// Whether the wiki is enabled.
	// Default: true
	// +optional
	HasWiki *bool `json:"has_wiki,omitempty"`

	// Either true to enable projects for this repository or false to disable them. Note: If you're creating a repository in an organization that has disabled repository projects, the default is false, and if you pass true, the API returns an error.
	// Default: true
	// +optional
	HasProjects *bool `json:"has_projects,omitempty"`

	// Whether downloads are enabled.
	// Default: true
	// +optional
	HasDownloads *bool `json:"has_downloads,omitempty"`

	// Whether discussions are enabled.
	// Default: false
	// +optional
	HasDiscussions *bool `json:"has_discussions,omitempty"`

	// The visibility of the repository. Can be one of: public, private, internal.
	// +optional
	Visibility *string `json:"visibility,omitempty"`

	// Specify which security and analysis features to enable or disable for the repository.
	//
	// To use this parameter, you must have admin permissions for the repository or be an owner or security manager for the organization that owns the repository. For more information, see [Managing security managers in your organization].
	//
	// [Managing security managers in your organization]: https://docs.github.com/en/organizations/managing-peoples-access-to-your-organization-with-roles/managing-security-managers-in-your-organization
	// +optional
	SecurityAndAnalysis *SecurityAndAnalysis `json:"security_and_analysis,omitempty"`
}

// RepositoryStatus defines the observed state of Repository
type RepositoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	LastUpdateTimestamp      *metav1.Time              `json:"last_update_timestamp,omitempty"`
	Id                       *int64                    `json:"id,omitempty"`
	NodeId                   *string                   `json:"node_id,omitempty"`
	OwnerLogin               *string                   `json:"owner_login,omitempty"`
	OwnerNodeId              *int64                    `json:"owner_node_id,omitempty"`
	Name                     *string                   `json:"name,omitempty"`
	FullName                 *string                   `json:"full_name,omitempty"`
	Owner                    *string                   `json:"owner,omitempty"`
	Description              *string                   `json:"description,omitempty"`
	Homepage                 *string                   `json:"homepage,omitempty"`
	DefaultBranch            *string                   `json:"default_branch,omitempty"`
	TemplateOwner            *string                   `json:"template_owner,omitempty"`
	TemplateRepository       *string                   `json:"template_repository,omitempty"`
	AllowRebaseMerge         *bool                     `json:"allow_rebase_merge,omitempty"`
	AllowUpdateBranch        *bool                     `json:"allow_update_branch,omitempty"`
	AllowSquashMerge         *bool                     `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit         *bool                     `json:"allow_merge_commit,omitempty"`
	AllowAutoMerge           *bool                     `json:"allow_auto_merge,omitempty"`
	AllowForking             *bool                     `json:"allow_forking,omitempty"`
	WebCommitSignoffRequired *bool                     `json:"web_commit_signoff_required,omitempty"`
	DeleteBranchOnMerge      *bool                     `json:"delete_branch_on_merge,omitempty"`
	SquashMergeCommitTitle   *SquashMergeCommitTitle   `json:"squash_merge_commit_title,omitempty"`
	SquashMergeCommitMessage *SquashMergeCommitMessage `json:"squash_merge_commit_message,omitempty"`
	MergeCommitTitle         *MergeCommitTitle         `json:"merge_commit_title,omitempty"`
	MergeCommitMessage       *MergeCommitMessage       `json:"merge_commit_message,omitempty"`
	Topics                   []string                  `json:"topics,omitempty"`
	Archived                 *bool                     `json:"archived,omitempty"`
	HasIssues                *bool                     `json:"has_issues,omitempty"`
	HasWiki                  *bool                     `json:"has_wiki,omitempty"`
	HasProjects              *bool                     `json:"has_projects,omitempty"`
	HasDownloads             *bool                     `json:"has_downloads,omitempty"`
	HasDiscussions           *bool                     `json:"has_discussions,omitempty"`
	Visibility               *string                   `json:"visibility,omitempty"`
	SecurityAndAnalysis      *SecurityAndAnalysis      `json:"security_and_analysis,omitempty"`

	ParentName                    *string `json:"parent_name,omitempty"`
	ParentId                      *int64  `json:"parent_id,omitempty"`
	ParentNodeId                  *string `json:"parent_node_id,omitempty"`
	TemplateRepositoryOwnerLogin  *string `json:"template_repository_owner_login,omitempty"`
	TemplateRepositoryOwnerNodeId *string `json:"template_repository_owner_node_id,omitempty"`
	TemplateRepositoryName        *string `json:"template_repository_name,omitempty"`
	TemplateRepositoryId          *int64  `json:"template_repository_id,omitempty"`
	OrganizationLogin             *string `json:"organization_login,omitempty"`
	OrganizationId                *int64  `json:"organization_id,omitempty"`

	CreatedAt *metav1.Time `json:"created_at,omitempty"`
	PushedAt  *metav1.Time `json:"pushed_at,omitempty"`
	UpdatedAt *metav1.Time `json:"updated_at,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Repository is the Schema for the repositories API
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec,omitempty"`
	Status RepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

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
	AdvancedSecurity SecurityAndAnalysisFeature `json:"advanced_security"`
	// Use the status property to enable or disable secret scanning for this repository. For more information, see [About secret scanning].
	//
	// [About secret scanning]: https://docs.github.com/en/code-security/secret-scanning/about-secret-scanning
	SecretScanning SecurityAndAnalysisFeature `json:"secret_scanning"`
	// Use the status property to enable or disable secret scanning push protection for this repository. For more information, see [Protecting pushes with secret scanning].
	//
	// [Protecting pushes with secret scanning]: https://docs.github.com/en/code-security/secret-scanning/push-protection-for-repositories-and-organizations
	SecretScanningPushProtection SecurityAndAnalysisFeature `json:"secret_scanning_push_protection"`
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
