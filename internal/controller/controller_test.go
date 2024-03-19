package controller

import (
	"context"
	"fmt"

	gh "github.com/eczy/github-operator/internal/github"
	"github.com/google/go-github/v60/github"
)

var _ GitHubRequester = &TestGitHubClient{}

type TestOrganization struct {
	TeamById      map[int64]*github.Team
	TeamBySlug    map[string]*github.Team
	TeamIdCounter int64

	GitHubOrganization *github.Organization
}

func NewTestOrganization(login string, id int64) *TestOrganization {
	return &TestOrganization{
		TeamById:      map[int64]*github.Team{},
		TeamBySlug:    map[string]*github.Team{},
		TeamIdCounter: 0,
		GitHubOrganization: &github.Organization{
			Login: github.String(login),
			ID:    github.Int64(id),
			Name:  github.String(login),
		},
	}
}

type TestGitHubClient struct {
	OrgsBySlug   map[string]*TestOrganization
	OrgsById     map[int64]*TestOrganization
	OrgIdCounter int64
}

type TestGitHubClientOption = func(*TestGitHubClient)

func WithTestOrganization(org TestOrganization) TestGitHubClientOption {
	return func(tghc *TestGitHubClient) {
		tghc.OrgsById[org.GitHubOrganization.GetID()] = &org
		tghc.OrgsBySlug[org.GitHubOrganization.GetLogin()] = &org
	}
}

func NewTestGitHubClient(opts ...TestGitHubClientOption) *TestGitHubClient {
	client := &TestGitHubClient{
		OrgsBySlug: map[string]*TestOrganization{},
		OrgsById:   map[int64]*TestOrganization{},
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (tghc *TestGitHubClient) CreateOrganization(ctx context.Context, login string) (*github.Organization, error) {
	if _, ok := tghc.OrgsBySlug[login]; !ok {
		org := NewTestOrganization(login, tghc.OrgIdCounter)
		tghc.OrgsBySlug[login] = org
		tghc.OrgsById[tghc.OrgIdCounter] = org
		tghc.OrgIdCounter += 1
		return org.GitHubOrganization, nil
	}
	return nil, fmt.Errorf("failed to create org '%s'", login)
}

// TeamRequester
func (tghc *TestGitHubClient) GetTeamBySlug(ctx context.Context, org string, slug string) (*github.Team, error) {
	errMsg := fmt.Errorf("no team with slug '%s' in org '%s'", slug, org)
	if organization, ok := tghc.OrgsBySlug[org]; ok {
		if organization.TeamBySlug == nil {
			return nil, errMsg
		}
		if team, ok := organization.TeamBySlug[slug]; ok {
			return team, nil
		} else {
			return nil, &gh.TeamNotFoundError{
				OrgSlug:  github.String(org),
				TeamSlug: github.String(slug),
			}
		}
	}
	return nil, errMsg
}

func (tghc *TestGitHubClient) GetTeamById(ctx context.Context, orgId int64, teamId int64) (*github.Team, error) {
	errMsg := fmt.Errorf("no team with id '%d' in org '%d'", teamId, orgId)
	if organization, ok := tghc.OrgsById[orgId]; ok {
		if organization.TeamById == nil {
			return nil, nil
		}
		if team, ok := organization.TeamById[teamId]; ok {
			return team, nil
		} else {
			return nil, &gh.TeamNotFoundError{
				OrgId:  github.Int64(orgId),
				TeamId: github.Int64(teamId),
			}
		}
	}
	return nil, errMsg
}

// Note: newTeam.Name will be used as the slug for the new team without special handling. Do not include illegal characters for slugs.
func (tghc *TestGitHubClient) CreateTeam(ctx context.Context, org string, newTeam github.NewTeam) (*github.Team, error) {
	if organization, ok := tghc.OrgsBySlug[org]; ok {
		if _, ok := organization.TeamBySlug[newTeam.Name]; ok {
			return nil, fmt.Errorf("team '%s' already exists in org '%s'", newTeam.Name, org)
		}
		id := organization.TeamIdCounter
		organization.TeamIdCounter += 1
		team := &github.Team{
			ID:          github.Int64(id),
			Name:        &newTeam.Name,
			Description: newTeam.Description,
			Organization: &github.Organization{
				Login: github.String(organization.GitHubOrganization.GetLogin()),
				ID:    github.Int64(organization.GitHubOrganization.GetID()),
			},
			Slug: &newTeam.Name,
		}
		organization.TeamById[id] = team
		organization.TeamBySlug[newTeam.Name] = team
		return team, nil
	}
	return nil, fmt.Errorf("failed to create team")
}

func (tghc *TestGitHubClient) UpdateTeamBySlug(ctx context.Context, org string, slug string, newTeam github.NewTeam) (*github.Team, error) {
	if organization, ok := tghc.OrgsBySlug[org]; ok {
		if team, ok := organization.TeamBySlug[slug]; ok {
			team.Name = &newTeam.Name
			if newTeam.Description != nil {
				team.Description = newTeam.Description
			}
			return team, nil
		}
		return nil, fmt.Errorf("team '%s' not found in org '%s'", slug, org)
	}
	return nil, fmt.Errorf("org '%s' not found", org)
}

func (tghc *TestGitHubClient) UpdateTeamById(ctx context.Context, org int64, teamId int64, newTeam github.NewTeam) (*github.Team, error) {
	if organization, ok := tghc.OrgsById[org]; ok {
		if team, ok := organization.TeamById[teamId]; ok {
			team.Name = &newTeam.Name
			if newTeam.Description != nil {
				team.Description = newTeam.Description
			}
			return team, nil
		}
		return nil, fmt.Errorf("team '%d' not found in org '%d'", teamId, org)
	}
	return nil, fmt.Errorf("org '%d' not found", org)
}

func (tghc *TestGitHubClient) DeleteTeamBySlug(ctx context.Context, org string, slug string) error {
	if organization, ok := tghc.OrgsBySlug[org]; ok {
		if team, ok := organization.TeamBySlug[slug]; ok {
			delete(organization.TeamBySlug, slug)
			delete(organization.TeamById, *team.ID)
			return nil
		}
		return fmt.Errorf("team '%s' not found in org '%s'", slug, org)
	}
	return fmt.Errorf("org '%s' not found", org)
}

func (tghc *TestGitHubClient) DeleteTeamById(ctx context.Context, org int64, teamId int64) error {
	if organization, ok := tghc.OrgsById[org]; ok {
		if team, ok := organization.TeamById[teamId]; ok {
			delete(organization.TeamBySlug, *team.Slug)
			delete(organization.TeamById, teamId)
			return nil
		}
		return fmt.Errorf("team '%d' not found in org '%d'", teamId, org)
	}
	return fmt.Errorf("org '%d' not found", org)
}

// DeleteOrganization implements GitHubRequester.
func (tghc *TestGitHubClient) DeleteOrganization(ctx context.Context, login string) error {
	if org, ok := tghc.OrgsBySlug[login]; ok {
		id := org.GitHubOrganization.GetID()
		delete(tghc.OrgsById, id)
		delete(tghc.OrgsBySlug, login)
		return nil
	}
	return fmt.Errorf("org '%s' not found", login)

}

// GetOrganization implements GitHubRequester.
func (tghc *TestGitHubClient) GetOrganization(ctx context.Context, login string) (*github.Organization, error) {
	if org, ok := tghc.OrgsBySlug[login]; ok {
		return org.GitHubOrganization, nil
	}
	return nil, fmt.Errorf("org '%s' not found", login)
}

// UpdateOrganization implements GitHubRequester.
// Temporarily only considers name and description.
func (tghc *TestGitHubClient) UpdateOrganization(ctx context.Context, login string, updateOrg *github.Organization) (*github.Organization, error) {
	if _, ok := tghc.OrgsBySlug[login]; ok {
		org := tghc.OrgsBySlug[login].GitHubOrganization
		org.Name = updateOrg.Name
		org.Description = updateOrg.Description
		tghc.OrgsBySlug[login].GitHubOrganization = org
		return org, nil
	}
	return nil, fmt.Errorf("org '%s' not found", login)
}
