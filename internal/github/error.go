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
	Login *string
	Id    *int64
}

func (e *OrganizationNotFoundError) Error() string {
	if e.Login != nil {
		return fmt.Sprintf("organization '%s' not found", *e.Login)

	} else if e.Id != nil {
		return fmt.Sprintf("organization '%d' not found", *e.Id)
	} else {
		return "organization not found"
	}
}
