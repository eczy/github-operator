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

package github

import "fmt"

type TeamNotFoundError struct {
	OrgSlug  *string
	OrgId    *int64
	TeamSlug *string
	TeamId   *int64
}

func (e *TeamNotFoundError) Error() string {
	if e.OrgSlug != nil && e.TeamSlug != nil {
		return fmt.Sprintf("team '%s' not found in org '%s'", *e.TeamSlug, *e.OrgSlug)
	} else if e.OrgId != nil && e.TeamId != nil {
		return fmt.Sprintf("team %d not found in org %d", *e.TeamId, *e.OrgId)
	} else {
		return "team not found"
	}
}

type OrganizationNotFoundError struct {
	Login      *string
	DatabaseId *int64
}

func (e *OrganizationNotFoundError) Error() string {
	if e.Login != nil {
		return fmt.Sprintf("organization '%s' not found", *e.Login)
	} else if e.DatabaseId != nil {
		return fmt.Sprintf("organization '%d' not found", *e.DatabaseId)
	} else {
		return "organization not found"
	}
}

type RepositoryNotFoundError struct {
	OwnerLogin *string
	OwnerId    *int64
	Slug       *string
	Id         *int64
}

func (e *RepositoryNotFoundError) Error() string {
	if e.OwnerLogin != nil && e.Slug != nil {
		return fmt.Sprintf("repository '%s' not found for owner '%s'", *e.Slug, *e.OwnerLogin)
	} else if e.OwnerId != nil && e.Id != nil {
		return fmt.Sprintf("repository %d not found for owner %d", *e.Id, *e.OwnerId)
	} else {
		return "repository not found"
	}
}
