package github

import (
	"context"

	"github.com/google/go-github/v60/github"
)

func (c *Client) GetOrganization(ctx context.Context, login string) (*github.Organization, error) {
	organization, resp, err := c.rest.Organizations.Get(ctx, login)
	if resp.StatusCode == 404 {
		return nil, &OrganizationNotFoundError{Login: &login}
	} else if err != nil {
		return nil, err
	}
	return organization, nil
}

func (c *Client) UpdateOrganization(ctx context.Context, login string, updateOrg *github.Organization) (*github.Organization, error) {
	organization, _, err := c.rest.Organizations.Edit(ctx, login, updateOrg)
	if err != nil {
		return nil, err
	}
	return organization, nil
}

func (c *Client) DeleteOrganization(ctx context.Context, login string) error {
	_, err := c.rest.Organizations.Delete(ctx, login)
	return err
}
