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

// OrganizationSpec defines the desired state of Organization
type OrganizationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The organization name. The name is not case sensitive.
	Login string `json:"login"`

	// The shorthand name of the company.
	Name string `json:"name"`

	// Billing email address. This address is not publicized.
	BillingEmail string `json:"billing_email,omitempty"`

	// The company name.
	Company string `json:"comapny,omitempty"`

	// The publicly visible email address.
	Email string `json:"email"`

	// The Twitter username of the company.
	TwitterUsername *string `json:"twitter_username,omitempty"`

	// The location.
	Location *string `json:"location,omitempty"`

	// The description of the company.
	Description *string `json:"description,omitempty"`

	// Whether an organization can use organization projects.
	HasOrganizationProjects *bool `json:"has_organization_projects,omitempty"`

	// Whether repositories that belong to the organization can use repository projects.
	HasRepositoryProjects *bool `json:"has_repository_projects,omitempty"`

	// Default permission level members have for organization repositories.
	// Can be one of: read, write, admin, none
	DefaultRepositoryPermission *string `json:"default_repository_permission,omitempty"` // TODO: enum

	// Whether of non-admin organization members can create repositories.
	MembersCanCreateRepositories *bool `json:"members_can_create_repositories,omitempty"`

	// Whether organization members can create internal repositories, which are visible to all enterprise members. You can only allow members to create internal repositories if your organization is associated with an enterprise account using GitHub Enterprise Cloud or GitHub Enterprise Server 2.20+.
	MembersCanCreateInternalRepositories *bool `json:"members_can_create_internal_repositories,omitempty"`

	// Whether organization members can create private repositories, which are visible to organization members with permission.
	MembersCanCreatePrivateRepositories *bool `json:"members_can_create_private_repositories,omitempty"`

	// Whether organization members can create public repositories, which are visible to anyone.
	MembersCanCreatePublicRepositories *bool `json:"members_can_create_public_repositories,omitempty"`

	// Whether organization members can create GitHub Pages sites.
	MembersCanCreatePages *bool `json:"members_can_create_pages,omitempty"`

	// Whether organization members can create public GitHub Pages sites.
	MembersCanCreatePublicPages *bool `json:"members_can_create_public_pages,omitempty"`

	// Whether organization members can create private GitHub Pages sites.
	MembersCanCreatePrivatePages *bool `json:"members_can_create_private_pages,omitempty"`

	// Whether organization members can create private GitHub Pages sites.
	MembersCanForkPrivateRepositories *bool `json:"members_can_fork_private_repositories,omitempty"`

	// Whether contributors to organization repositories are required to sign off on commits they make through GitHub's web interface.
	WebCommitSignoffRequired *bool `json:"web_commit_signoff_required,omitempty"`

	Blog *string `json:"blog,omitempty"`

	// Whether GitHub Advanced Security is automatically enabled for new repositories.
	AdvancedSecurityEnabledForNewRepositories *bool `json:"advanced_security_enabled_for_new_repositories,omitempty"`

	// Whether Dependabot alerts is automatically enabled for new repositories.
	DependabotAlertsEnabledForNewRepositories *bool `json:"dependabot_alerts_enabled_for_new_repositories,omitempty"`

	// Whether Dependabot security updates is automatically enabled for new repositories.
	DependabotSecurityUpdatesEnabledForNewRepositories *bool `json:"dependabot_security_updates_enabled_for_new_repositories,omitempty"`

	// Whether dependency graph is automatically enabled for new repositories.
	DependencyGraphEnabledForNewRepositories *bool `json:"dependency_graph_enabled_for_new_repositories,omitempty"`

	// Whether secret scanning is automatically enabled for new repositories.
	SecretScanningEnabledForNewRepositories *bool `json:"secret_scanning_enabled_for_new_repositories,omitempty"`

	// Whether secret scanning push protection is automatically enabled for new repositories.
	SecretScanningPushProtectionEnabledForNewRepositories *bool `json:"secret_scanning_push_protection_enabled_for_new_repositories,omitempty"`
}

// OrganizationStatus defines the observed state of Organization
type OrganizationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Login               *string      `json:"login,omitempty"`
	Id                  *int64       `json:"id,omitempty"`
	LastUpdateTimestamp *metav1.Time `json:"last_update_timestamp,omitempty"`

	Name                                                  string  `json:"name"`
	BillingEmail                                          string  `json:"billing_email,omitempty"`
	Company                                               string  `json:"comapny,omitempty"`
	Email                                                 string  `json:"email"`
	TwitterUsername                                       *string `json:"twitter_username,omitempty"`
	Location                                              *string `json:"location,omitempty"`
	Description                                           *string `json:"description,omitempty"`
	HasOrganizationProjects                               *bool   `json:"has_organization_projects,omitempty"`
	HasRepositoryProjects                                 *bool   `json:"has_repository_projects,omitempty"`
	DefaultRepositoryPermission                           *string `json:"default_repository_permission,omitempty"` // TODO: enum
	MembersCanCreateRepositories                          *bool   `json:"members_can_create_repositories,omitempty"`
	MembersCanCreateInternalRepositories                  *bool   `json:"members_can_create_internal_repositories,omitempty"`
	MembersCanCreatePrivateRepositories                   *bool   `json:"members_can_create_private_repositories,omitempty"`
	MembersCanCreatePublicRepositories                    *bool   `json:"members_can_create_public_repositories,omitempty"`
	MembersCanCreatePages                                 *bool   `json:"members_can_create_pages,omitempty"`
	MembersCanCreatePublicPages                           *bool   `json:"members_can_create_public_pages,omitempty"`
	MembersCanCreatePrivatePages                          *bool   `json:"members_can_create_private_pages,omitempty"`
	MembersCanForkPrivateRepositories                     *bool   `json:"members_can_fork_private_repositories,omitempty"`
	WebCommitSignoffRequired                              *bool   `json:"web_commit_signoff_required,omitempty"`
	Blog                                                  *string `json:"blog,omitempty"`
	AdvancedSecurityEnabledForNewRepositories             *bool   `json:"advanced_security_enabled_for_new_repositories,omitempty"`
	DependabotAlertsEnabledForNewRepositories             *bool   `json:"dependabot_alerts_enabled_for_new_repositories,omitempty"`
	DependabotSecurityUpdatesEnabledForNewRepositories    *bool   `json:"dependabot_security_updates_enabled_for_new_repositories,omitempty"`
	DependencyGraphEnabledForNewRepositories              *bool   `json:"dependency_graph_enabled_for_new_repositories,omitempty"`
	SecretScanningEnabledForNewRepositories               *bool   `json:"secret_scanning_enabled_for_new_repositories,omitempty"`
	SecretScanningPushProtectionEnabledForNewRepositories *bool   `json:"secret_scanning_push_protection_enabled_for_new_repositories,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Organization is the Schema for the organizations API
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationSpec   `json:"spec,omitempty"`
	Status OrganizationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OrganizationList contains a list of Organization
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Organization `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}
