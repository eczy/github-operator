package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
	"github.com/shurcooL/githubv4"
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

type TeamRepositoryPermission struct {
	OrganizationLogin string
	TeamSlug          string
	RepositoryName    string
	RepositoryId      string
	Permission        string
}

func (c *Client) GetTeamRepositoryPermissions(ctx context.Context, org, slug string) ([]TeamRepositoryPermission, error) {
	var q struct {
		Organization struct {
			Team struct {
				Repositories struct {
					Edges []struct {
						Permission string
					}
					Nodes []struct {
						Id   string
						Name string
					}
					// TODO: move PageInfo to a common spot
					PageInfo PageInfo
				} `graphql:"repositories:(first: 100, after: $cursor)"`
			} `graphql:"team(slug: $slug)"`
		} `graphql:"organization(login: $login)"`
	}

	variables := map[string]interface{}{
		"login":  org,
		"slug":   slug,
		"cursor": (*githubv4.String)(nil),
	}

	out := []TeamRepositoryPermission{}
	for {
		err := c.graphql.Query(ctx, &q, variables)
		if err != nil {
			return nil, err
		}

		for i, edge := range q.Organization.Team.Repositories.Edges {
			node := q.Organization.Team.Repositories.Nodes[i]
			out = append(out, TeamRepositoryPermission{
				OrganizationLogin: org,
				TeamSlug:          slug,
				RepositoryName:    node.Name,
				RepositoryId:      node.Id,
				Permission:        edge.Permission,
			})
		}

		if q.Organization.Team.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = q.Organization.Team.Repositories.PageInfo.EndCursor
	}

	return out, nil
}

func (c *Client) UpdateTeamRepositoryPermissions(ctx context.Context, org, slug string, repoName, permission string) error {
	_, err := c.rest.Teams.AddTeamRepoBySlug(ctx, org, slug, org, repoName, &github.TeamAddTeamRepoOptions{
		Permission: permission,
	})
	return err
}

func (c *Client) RemoveTeamRepositoryPermissions(ctx context.Context, org, slug string, repoName, permission string) error {
	_, err := c.rest.Teams.RemoveTeamRepoBySlug(ctx, org, slug, org, repoName)
	return err
}
