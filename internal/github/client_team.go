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

func (c *Client) GetTeamById(ctx context.Context, orgId, teamId int64) (*github.Team, error) {
	team, resp, err := c.rest.Teams.GetTeamByID(ctx, orgId, teamId)
	if resp.StatusCode == 404 {
		return nil, &TeamNotFoundError{
			OrgId:  github.Int64(orgId),
			TeamId: github.Int64(teamId),
		}
	} else if err != nil {
		return nil, fmt.Errorf("getting GitHub team: %w", err)
	}
	return team, nil
}

func (c *Client) GetTeamByNodeId(ctx context.Context, nodeId string) (*github.Team, error) {
	// TOOD: this is inefficient since it takes two API calls.
	// This is done this way for the moment since it lets us update an existing resource in event of naming changes.
	// In the future, we should probably move to an explicit internal data structure instead
	// of relying on a library and define conversions.
	var q struct {
		Node struct {
			Team struct {
				DatabaseId   int64
				Organization struct {
					DatabaseId int64
				}
			} `graphql:"... on Team"`
		} `graphql:"node(id: $nodeId)"`
	}

	variables := map[string]interface{}{
		"nodeId": nodeId,
	}

	err := c.graphql.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}
	return c.GetTeamById(ctx, q.Node.Team.Organization.DatabaseId, q.Node.Team.DatabaseId)
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

// repository permissions in greatest to least order - for local use
var repositoryPermissions = []string{"admin", "maintain", "push", "triage", "pull"}

// Team repository permissions
func maxPermissionFromMap(permissionMap map[string]bool) (string, error) {
	for _, perm := range repositoryPermissions {
		if hasPerm, ok := permissionMap[perm]; ok {
			if hasPerm {
				return perm, nil
			}
		}
	}
	return "", fmt.Errorf("no valid permission found in permission map")
}

type TeamRepositoryPermission struct {
	OrganizationLogin string
	TeamSlug          string
	RepositoryName    string
	RepositoryId      string
	Permission        string
}

// assume repo is in the same org as team
func (c *Client) GetTeamRepositoryPermission(ctx context.Context, org, slug, repoName string) (*TeamRepositoryPermission, error) {
	repo, _, err := c.rest.Teams.IsTeamRepoBySlug(ctx, org, slug, org, repoName)
	if err != nil {
		return nil, err
	}

	permission, err := maxPermissionFromMap(repo.GetPermissions())
	if err != nil {
		return nil, err
	}

	return &TeamRepositoryPermission{
		OrganizationLogin: org,
		TeamSlug:          slug,
		RepositoryName:    repo.GetName(),
		RepositoryId:      repo.GetNodeID(),
		Permission:        permission,
	}, nil
}

// assume repo is in the same org as team
func (c *Client) GetTeamRepositoryPermissions(ctx context.Context, org, slug string) ([]*TeamRepositoryPermission, error) {
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
				} `graphql:"repositories(first: 100, after: $cursor)"`
			} `graphql:"team(slug: $slug)"`
		} `graphql:"organization(login: $login)"`
	}

	variables := map[string]interface{}{
		"login":  githubv4.String(org),
		"slug":   githubv4.String(slug),
		"cursor": (*githubv4.String)(nil),
	}

	out := []*TeamRepositoryPermission{}
	for {
		err := c.graphql.Query(ctx, &q, variables)
		if err != nil {
			return nil, err
		}

		for i, edge := range q.Organization.Team.Repositories.Edges {
			node := q.Organization.Team.Repositories.Nodes[i]
			out = append(out, &TeamRepositoryPermission{
				OrganizationLogin: org,
				TeamSlug:          slug,
				RepositoryName:    node.Name,
				RepositoryId:      node.Id,
				Permission:        edge.Permission,
			})
		}

		if !q.Organization.Team.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = &q.Organization.Team.Repositories.PageInfo.EndCursor
	}

	return out, nil
}

func (c *Client) UpdateTeamRepositoryPermissions(ctx context.Context, org, slug string, repoName, permission string) error {
	_, err := c.rest.Teams.AddTeamRepoBySlug(ctx, org, slug, org, repoName, &github.TeamAddTeamRepoOptions{
		Permission: permission,
	})
	return err
}

func (c *Client) RemoveTeamRepositoryPermissions(ctx context.Context, org, slug string, repoName string) error {
	_, err := c.rest.Teams.RemoveTeamRepoBySlug(ctx, org, slug, org, repoName)
	return err
}
