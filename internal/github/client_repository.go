package github

import (
	"context"

	"github.com/google/go-github/v60/github"
)

// Repositories
func (c *Client) GetRepositoryByName(ctx context.Context, owner string, name string) (*github.Repository, error) {
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

func (c *Client) GetRepositoryByDatabaseId(ctx context.Context, dbId int64) (*github.Repository, error) {
	repo, resp, err := c.rest.Repositories.GetByID(ctx, dbId)
	if resp.StatusCode == 404 {
		return nil, &RepositoryNotFoundError{
			Id: &dbId,
		}
	} else if err != nil {
		return nil, err
	}
	return repo, nil
}

func (c *Client) GetRepositoryByNodeId(ctx context.Context, nodeId string) (*github.Repository, error) {
	// TOOD: this is inefficient since it takes two API calls.
	// This is done this way for the moment since it lets us update an existing resource in event of naming changes.
	// In the future, we should probably move to an explicit internal data structure instead
	// of relying on a library and define conversions.
	var q struct {
		Node struct {
			Repository struct {
				DatabaseId int64
			} `graphql:"... on Repository"`
		} `graphql:"node(id: $nodeId)"`
	}

	variables := map[string]interface{}{
		"nodeId": nodeId,
	}

	err := c.graphql.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}
	return c.GetRepositoryByDatabaseId(ctx, q.Node.Repository.DatabaseId)
}

func (c *Client) UpdateRepositoryByName(ctx context.Context, owner, name string, update *github.Repository) (*github.Repository, error) {
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

func (c *Client) DeleteRepositoryByName(ctx context.Context, owner, name string) error {
	_, err := c.rest.Repositories.Delete(ctx, owner, name)
	return err
}
