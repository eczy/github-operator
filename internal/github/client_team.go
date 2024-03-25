package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
)

// Teams
func (c *Client) GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, error) {
	team, resp, err := c.rest.Teams.GetTeamBySlug(ctx, org, slug)
	if resp.StatusCode == 404 {
		return nil, &TeamNotFoundError{
			OrgSlug:  github.String(org),
			TeamSlug: github.String(slug),
		}
	} else if err != nil {
		return nil, fmt.Errorf("getting GitHub team: %w", err)
	}
	return team, nil
}

func (c *Client) GetTeamById(ctx context.Context, org, teamId int64) (*github.Team, error) {
	team, resp, err := c.rest.Teams.GetTeamByID(ctx, org, teamId)
	if resp.StatusCode == 404 {
		return nil, &TeamNotFoundError{
			OrgId:  github.Int64(org),
			TeamId: github.Int64(teamId),
		}
	} else if err != nil {
		return nil, fmt.Errorf("getting GitHub team: %w", err)
	}
	return team, nil
}

func (c *Client) CreateTeam(ctx context.Context, org string, newTeam github.NewTeam) (*github.Team, error) {
	team, _, err := c.rest.Teams.CreateTeam(ctx, org, newTeam)
	if err != nil {
		return nil, fmt.Errorf("creating GitHub team: %w", err)
	}
	return team, nil
}

func (c *Client) UpdateTeamBySlug(ctx context.Context, org, slug string, newTeam github.NewTeam) (*github.Team, error) {
	team, _, err := c.rest.Teams.EditTeamBySlug(ctx, org, slug, newTeam, false)
	if err != nil {
		return nil, fmt.Errorf("editing GitHub team: %w", err)
	}
	return team, nil
}

func (c *Client) UpdateTeamById(ctx context.Context, org, teamId int64, newTeam github.NewTeam) (*github.Team, error) {
	team, _, err := c.rest.Teams.EditTeamByID(ctx, org, teamId, newTeam, false)
	if err != nil {
		return nil, fmt.Errorf("editing GitHub team: %w", err)
	}
	return team, nil
}

func (c *Client) DeleteTeamBySlug(ctx context.Context, org, slug string) error {
	_, err := c.rest.Teams.DeleteTeamBySlug(ctx, org, slug)
	if err != nil {
		return fmt.Errorf("deleting GitHub team: %w", err)
	}
	return nil
}

func (c *Client) DeleteTeamById(ctx context.Context, org, slug int64) error {
	_, err := c.rest.Teams.DeleteTeamByID(ctx, org, slug)
	if err != nil {
		return fmt.Errorf("deleting GitHub team: %w", err)
	}
	return nil
}

// Organizations

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
