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

// OrganizationSpec defines the desired state of Organization
type OrganizationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// +kubebuilder:validation:MinLength=1
	// The organization name. The name is not case sensitive.
	Login string `json:"login"`

	// +kubebuilder:validation:MinLength=1
	// The shorthand name of the company.
	// +optional
	Name *string `json:"name,omitempty"`

	// +kubebuilder:validation:MinLength=1
	// Billing email address. This address is not publicized.
	// +optional
	BillingEmail *string `json:"billingEmail,omitempty"`

	// +kubebuilder:validation:MinLength=1
	// The company name.
	// +optional
	Company *string `json:"company,omitempty"`

	// The publicly visible email address.
	// +optional
	Email *string `json:"email,omitempty"`

	// The Twitter username of the company.
	// +optional
	TwitterUsername *string `json:"twitterUsername,omitempty"`

	// The location.
	// +optional
	Location *string `json:"location,omitempty"`

	// The description of the company.
	// +optional
	Description *string `json:"description,omitempty"`

	// Whether an organization can use organization projects.
	// +optional
	HasOrganizationProjects *bool `json:"hasOrganizationProjects,omitempty"`

	// Whether repositories that belong to the organization can use repository projects.
	// +optional
	HasRepositoryProjects *bool `json:"hasRepositoryProjects,omitempty"`

	// Default permission level members have for organization repositories.
	// Can be one of: read, write, admin, none
	// +optional
	DefaultRepositoryPermission *DefaultRepositoryPermission `json:"defaultRepositoryPermission,omitempty"`

	// Whether of non-admin organization members can create repositories.
	// +optional
	MembersCanCreateRepositories *bool `json:"membersCanCreateRepositories,omitempty"`

	// Whether organization members can create internal repositories, which are visible to all enterprise members. You can only allow members to create internal repositories if your organization is associated with an enterprise account using GitHub Enterprise Cloud or GitHub Enterprise Server 2.20+.
	// +optional
	MembersCanCreateInternalRepositories *bool `json:"membersCanCreateInternalRepositories,omitempty"`

	// Whether organization members can create private repositories, which are visible to organization members with permission.
	// +optional
	MembersCanCreatePrivateRepositories *bool `json:"membersCanCreatePrivateRepositories,omitempty"`

	// Whether organization members can create public repositories, which are visible to anyone.
	// +optional
	MembersCanCreatePublicRepositories *bool `json:"membersCanCreatePublicRepositories,omitempty"`

	// Whether organization members can create GitHub Pages sites.
	// +optional
	MembersCanCreatePages *bool `json:"membersCanCreatePages,omitempty"`

	// Whether organization members can create public GitHub Pages sites.
	// +optional
	MembersCanCreatePublicPages *bool `json:"membersCanCreatePublicPages,omitempty"`

	// Whether organization members can create private GitHub Pages sites.
	// +optional
	MembersCanCreatePrivatePages *bool `json:"membersCanCreatePrivatePages,omitempty"`

	// Whether organization members can create private GitHub Pages sites.
	// +optional
	MembersCanForkPrivateRepositories *bool `json:"membersCanForkPrivateRepositories,omitempty"`

	// Whether contributors to organization repositories are required to sign off on commits they make through GitHub's web interface.
	// +optional
	WebCommitSignoffRequired *bool `json:"webCommitSignoffRequired,omitempty"`

	// +optional
	Blog *string `json:"blog,omitempty"`

	// Whether GitHub Advanced Security is automatically enabled for new repositories.
	// +optional
	AdvancedSecurityEnabledForNewRepositories *bool `json:"advancedSecurityEnabledForNewRepositories,omitempty"`

	// Whether Dependabot alerts is automatically enabled for new repositories.
	// +optional
	DependabotAlertsEnabledForNewRepositories *bool `json:"dependabotAlertsEnabledForNewRepositories,omitempty"`

	// Whether Dependabot security updates is automatically enabled for new repositories.
	// +optional
	DependabotSecurityUpdatesEnabledForNewRepositories *bool `json:"dependabotSecurityUpdatesEnabledForNewRepositories,omitempty"`

	// Whether dependency graph is automatically enabled for new repositories.
	// +optional
	DependencyGraphEnabledForNewRepositories *bool `json:"dependencyGraphEnabledForNewRepositories,omitempty"`

	// Whether secret scanning is automatically enabled for new repositories.
	// +optional
	SecretScanningEnabledForNewRepositories *bool `json:"secretScanningEnabledForNewRepositories,omitempty"`

	// Whether secret scanning push protection is automatically enabled for new repositories.
	// +optional
	SecretScanningPushProtectionEnabledForNewRepositories *bool `json:"secretScanningPushProtectionEnabledForNewRepositories,omitempty"`
}

// OrganizationStatus defines the observed state of Organization.
type OrganizationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Login                                                 *string                      `json:"login,omitempty"`
	NodeId                                                *string                      `json:"nodeId,omitempty"`
	LastUpdateTimestamp                                   *metav1.Time                 `json:"lastUpdateTimestamp,omitempty"`
	Name                                                  string                       `json:"name"`
	BillingEmail                                          string                       `json:"billingEmail,omitempty"`
	Company                                               string                       `json:"company,omitempty"`
	Email                                                 string                       `json:"email"`
	TwitterUsername                                       *string                      `json:"twitterUsername,omitempty"`
	Location                                              *string                      `json:"location,omitempty"`
	Description                                           *string                      `json:"description,omitempty"`
	HasOrganizationProjects                               *bool                        `json:"hasOrganizationProjects,omitempty"`
	HasRepositoryProjects                                 *bool                        `json:"hasRepositoryProjects,omitempty"`
	DefaultRepositoryPermission                           *DefaultRepositoryPermission `json:"defaultRepositoryPermission,omitempty"`
	MembersCanCreateRepositories                          *bool                        `json:"membersCanCreateRepositories,omitempty"`
	MembersCanCreateInternalRepositories                  *bool                        `json:"membersCanCreateInternalRepositories,omitempty"`
	MembersCanCreatePrivateRepositories                   *bool                        `json:"membersCanCreatePrivateRepositories,omitempty"`
	MembersCanCreatePublicRepositories                    *bool                        `json:"membersCanCreatePublicRepositories,omitempty"`
	MembersCanCreatePages                                 *bool                        `json:"membersCanCreatePages,omitempty"`
	MembersCanCreatePublicPages                           *bool                        `json:"membersCanCreatePublicPages,omitempty"`
	MembersCanCreatePrivatePages                          *bool                        `json:"membersCanCreatePrivatePages,omitempty"`
	MembersCanForkPrivateRepositories                     *bool                        `json:"membersCanForkPrivateRepositories,omitempty"`
	WebCommitSignoffRequired                              *bool                        `json:"webCommitSignoffRequired,omitempty"`
	Blog                                                  *string                      `json:"blog,omitempty"`
	AdvancedSecurityEnabledForNewRepositories             *bool                        `json:"advancedSecurityEnabledForNewRepositories,omitempty"`
	DependabotAlertsEnabledForNewRepositories             *bool                        `json:"dependabotAlertsEnabledForNewRepositories,omitempty"`
	DependabotSecurityUpdatesEnabledForNewRepositories    *bool                        `json:"dependabotSecurityUpdatesEnabledForNewRepositories,omitempty"`
	DependencyGraphEnabledForNewRepositories              *bool                        `json:"dependencyFraphEnabledForNewRepositories,omitempty"`
	SecretScanningEnabledForNewRepositories               *bool                        `json:"secretScanningEnabledForNewRepositories,omitempty"`
	SecretScanningPushProtectionEnabledForNewRepositories *bool                        `json:"secretScanningPushProtectionEnabledForNewRepositories,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Organization is the Schema for the organizations API
type Organization struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Organization
	// +required
	Spec OrganizationSpec `json:"spec"`

	// status defines the observed state of Organization
	// +optional
	Status OrganizationStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// OrganizationList contains a list of Organization
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Organization `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}

// +kubebuilder:validation:Enum=read;write;none;admin
type DefaultRepositoryPermission string

const (
	DefaultRepositoryPermissionRead  DefaultRepositoryPermission = "read"
	DefaultRepositoryPermissionWrite DefaultRepositoryPermission = "write"
	DefaultRepositoryPermissionNone  DefaultRepositoryPermission = "none"
	DefaultRepositoryPermissionAdmin DefaultRepositoryPermission = "admin"
)
