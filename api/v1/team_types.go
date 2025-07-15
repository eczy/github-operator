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

// TeamSpec defines the desired state of Team
type TeamSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// +kubebuilder:validation:MinLength=1
	// Organization name. Not case sensitive.
	Organization string `json:"organization"`

	// +kubebuilder:validation:MinLength=1
	// Name of the team.
	Name string `json:"name"`

	// Description of the team.
	// +optional
	Description *string `json:"description,omitempty"`

	// Level of privacy the team should have.
	// +optional
	Privacy *Privacy `json:"privacy,omitempty"`

	// Notification setting for members of the team.
	// +optional
	NotificationSetting *NotificationSetting `json:"notificationSetting,omitempty"`

	// ID of the team to set as the parent of this team
	// +optional
	ParentTeamId *int64 `json:"parentTeamId,omitempty"`

	// Repository permissions to assign to this team
	// +optional
	Repositories map[string]RepositoryPermission `json:"repositories,omitempty"`
}

// TeamStatus defines the observed state of Team.
type TeamStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Id                  *int64                          `json:"id,omitempty"`
	NodeId              *string                         `json:"nodeId,omitempty"`
	Slug                *string                         `json:"slug,omitempty"`
	LastUpdateTimestamp *metav1.Time                    `json:"lastUpdateTimestamp,omitempty"`
	OrganizationLogin   *string                         `json:"organizationLogin,omitempty"`
	OrganizationId      *int64                          `json:"organizationId,omitempty"`
	Name                *string                         `json:"name,omitempty"`
	Description         *string                         `json:"description,omitempty"`
	Privacy             *Privacy                        `json:"privacy,omitempty"`
	NotificationSetting *NotificationSetting            `json:"notificationSetting,omitempty"`
	ParentTeamId        *int64                          `json:"parentTeamId,omitempty"`
	ParentTeamSlug      *string                         `json:"parentTeamSlug,omitempty"`
	Repositories        map[string]RepositoryPermission `json:"repositories,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Team is the Schema for the teams API
type Team struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Team
	// +required
	Spec TeamSpec `json:"spec"`

	// status defines the observed state of Team
	// +optional
	Status TeamStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// TeamList contains a list of Team
type TeamList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Team `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Team{}, &TeamList{})
}

// Privacy configures the visibility of the team.
// +kubebuilder:validation:Enum=secret;closed
type Privacy string

const (
	// only visible to organization owners and members of this team.
	// a parent team cannot be secret.
	Secret Privacy = "secret"
	// visible to all members of this organization.
	// for a parent or child team: visible to all members of this organization.
	Closed Privacy = "closed"
)

// +kubebuilder:validation:Enum=notifications_enabled;notifications_disabled
type NotificationSetting string

const (
	// team members receive notifications when the team is @mentioned.
	Enabled NotificationSetting = "notificationsEnabled"
	// no one receives notifications.
	Disabled NotificationSetting = "notificationsDisabled"
)

// +kubebuilder:validation:Enum=admin;push;maintain;triage;pull
type RepositoryPermission string

const (
	Admin    RepositoryPermission = "admin"
	Push     RepositoryPermission = "push"
	Maintain RepositoryPermission = "maintain"
	Triage   RepositoryPermission = "triage"
	Pull     RepositoryPermission = "pull"
)
