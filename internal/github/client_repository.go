package github

import (
	"context"

	"github.com/google/go-github/v60/github"
)

// Repositories
func (c *Client) GetRepositoryBySlug(ctx context.Context, owner string, name string) (*github.Repository, error) {
	repo, resp, err := c.rest.Repositories.Get(ctx, owner, name)
	if resp.StatusCode == 404 {
		return nil, &RepositoryNotFoundError{
			OwnerLogin: github.String(owner),
			Slug:       github.String(name),
		}
	} else if err != nil {
		return nil, err
	}
	return repo, nil
}

func (c *Client) UpdateRepositoryBySlug(ctx context.Context, owner, name string, update *github.Repository) (*github.Repository, error) {
	repo, _, err := c.rest.Repositories.Edit(ctx, owner, name, update)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// Pass empty string as org to create a user-owned repo
func (c *Client) CreateRepository(ctx context.Context, org string, create *github.Repository) (*github.Repository, error) {
	repo, _, err := c.rest.Repositories.Create(ctx, org, create)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// Pass empty string as org to create a user-owned repo
func (c *Client) CreateRepositoryFromTemplate(ctx context.Context, templateOwner string, templateRepository string, req *github.TemplateRepoRequest) (*github.Repository, error) {
	repo, _, err := c.rest.Repositories.CreateFromTemplate(ctx, templateOwner, templateRepository, req)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (c *Client) UpdateRepositoryTopics(ctx context.Context, owner string, repo string, topics []string) ([]string, error) {
	newTopics, _, err := c.rest.Repositories.ReplaceAllTopics(ctx, owner, repo, topics)
	if err != nil {
		return nil, err
	}
	return newTopics, nil
}

func (c *Client) DeleteRepositoryBySlug(ctx context.Context, owner, name string) error {
	_, err := c.rest.Repositories.Delete(ctx, owner, name)
	return err
}
