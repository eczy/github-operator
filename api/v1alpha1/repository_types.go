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

// RepositorySpec defines the desired state of Repository
type RepositorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name  string `json:"name"`
	Owner string `json:"owner"`

	Description              *string  `json:"description,omitempty"`
	Homepage                 *string  `json:"homepage,omitempty"`
	DefaultBranch            *string  `json:"default_branch,omitempty"`
	TemplateRepositoryOwner  *string  `json:"template_repository_owner,omitempty"`
	TemplateRepository       *string  `json:"template_repository,omitempty"`
	AllowRebaseMerge         *bool    `json:"allow_rebase_merge,omitempty"`
	AllowUpdateBranch        *bool    `json:"allow_update_branch,omitempty"`
	AllowSquashMerge         *bool    `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit         *bool    `json:"allow_merge_commit,omitempty"`
	AllowAutoMerge           *bool    `json:"allow_auto_merge,omitempty"`
	AllowForking             *bool    `json:"allow_forking,omitempty"`
	WebCommitSignoffRequired *bool    `json:"web_commit_signoff_required,omitempty"`
	DeleteBranchOnMerge      *bool    `json:"delete_branch_on_merge,omitempty"`
	SquashMergeCommitTitle   *string  `json:"squash_merge_commit_title,omitempty"`   // Can be one of: "PR_TITLE", "COMMIT_OR_PR_TITLE"
	SquashMergeCommitMessage *string  `json:"squash_merge_commit_message,omitempty"` // Can be one of: "PR_BODY", "COMMIT_MESSAGES", "BLANK"
	MergeCommitTitle         *string  `json:"merge_commit_title,omitempty"`          // Can be one of: "PR_TITLE", "MERGE_MESSAGE"
	MergeCommitMessage       *string  `json:"merge_commit_message,omitempty"`        // Can be one of: "PR_BODY", "PR_TITLE", "BLANK"
	Topics                   []string `json:"topics,omitempty"`
	Archived                 *bool    `json:"archived,omitempty"`
	Disabled                 *bool    `json:"disabled,omitempty"`
	HasIssues                *bool    `json:"has_issues,omitempty"`
	HasWiki                  *bool    `json:"has_wiki,omitempty"`
	HasPages                 *bool    `json:"has_pages,omitempty"`
	HasProjects              *bool    `json:"has_projects,omitempty"`
	HasDownloads             *bool    `json:"has_downloads,omitempty"`
	HasDiscussions           *bool    `json:"has_discussions,omitempty"`
	IsTemplate               *bool    `json:"is_template,omitempty"`
	Visibility               *string  `json:"visibility,omitempty"`

	// TODO: security and analysis
}

// RepositoryStatus defines the observed state of Repository
type RepositoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LastUpdateTimestamp *metav1.Time `json:"last_update_timestamp,omitempty"`

	Id                          *int64       `json:"id,omitempty"`
	NodeID                      *string      `json:"node_id,omitempty"`
	OwnerLogin                  *string      `json:"owner_login,omitempty"`
	OwnerId                     *int64       `json:"owner_id,omitempty"`
	Name                        *string      `json:"name,omitempty"`
	FullName                    *string      `json:"full_name,omitempty"`
	Description                 *string      `json:"description,omitempty"`
	Homepage                    *string      `json:"homepage,omitempty"`
	DefaultBranch               *string      `json:"default_branch,omitempty"`
	CreatedAt                   *metav1.Time `json:"created_at,omitempty"`
	PushedAt                    *metav1.Time `json:"pushed_at,omitempty"`
	UpdatedAt                   *metav1.Time `json:"updated_at,omitempty"`
	Language                    *string      `json:"language,omitempty"`
	Fork                        *bool        `json:"fork,omitempty"`
	Size                        *int         `json:"size,omitempty"`
	ParentName                  *string      `json:"parent_slug,omitempty"`
	ParentId                    *int64       `json:"parent_id,omitempty"`
	TemplateRepositoryOwnerName *string      `json:"template_repository_owner_name,omitempty"`
	TemplateRepositoryName      *string      `json:"template_repository_slug,omitempty"`
	TemplateRepositoryId        *int64       `json:"template_repository_id,omitempty"`
	OrganizationLogin           *string      `json:"organization_login,omitempty"`
	OrganizationId              *int64       `json:"organization_id,omitempty"`
	AllowRebaseMerge            *bool        `json:"allow_rebase_merge,omitempty"`
	AllowUpdateBranch           *bool        `json:"allow_update_branch,omitempty"`
	AllowSquashMerge            *bool        `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit            *bool        `json:"allow_merge_commit,omitempty"`
	AllowAutoMerge              *bool        `json:"allow_auto_merge,omitempty"`
	AllowForking                *bool        `json:"allow_forking,omitempty"`
	WebCommitSignoffRequired    *bool        `json:"web_commit_signoff_required,omitempty"`
	DeleteBranchOnMerge         *bool        `json:"delete_branch_on_merge,omitempty"`
	SquashMergeCommitTitle      *string      `json:"squash_merge_commit_title,omitempty"`   // Can be one of: "PR_TITLE", "COMMIT_OR_PR_TITLE"
	SquashMergeCommitMessage    *string      `json:"squash_merge_commit_message,omitempty"` // Can be one of: "PR_BODY", "COMMIT_MESSAGES", "BLANK"
	MergeCommitTitle            *string      `json:"merge_commit_title,omitempty"`          // Can be one of: "PR_TITLE", "MERGE_MESSAGE"
	MergeCommitMessage          *string      `json:"merge_commit_message,omitempty"`        // Can be one of: "PR_BODY", "PR_TITLE", "BLANK"
	Topics                      []string     `json:"topics,omitempty"`
	Archived                    *bool        `json:"archived,omitempty"`
	Disabled                    *bool        `json:"disabled,omitempty"`
	HasIssues                   *bool        `json:"has_issues,omitempty"`
	HasWiki                     *bool        `json:"has_wiki,omitempty"`
	HasPages                    *bool        `json:"has_pages,omitempty"`
	HasProjects                 *bool        `json:"has_projects,omitempty"`
	HasDownloads                *bool        `json:"has_downloads,omitempty"`
	HasDiscussions              *bool        `json:"has_discussions,omitempty"`
	IsTemplate                  *bool        `json:"is_template,omitempty"`
	LicenseTemplate             *string      `json:"license_template,omitempty"`
	Visibility                  *string      `json:"visibility,omitempty"`
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
