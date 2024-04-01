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

func (c *Client) GetOrganizationByDatabaseId(ctx context.Context, dbId int64) (*github.Organization, error) {
	organization, resp, err := c.rest.Organizations.GetByID(ctx, dbId)
	if resp.StatusCode == 404 {
		return nil, &OrganizationNotFoundError{DatabaseId: &dbId}
	} else if err != nil {
		return nil, err
	}
	return organization, nil
}

func (c *Client) GetOrganizationByNodeId(ctx context.Context, nodeId string) (*github.Organization, error) {
	// TODO: this is inefficient since it takes two API calls.
	// This is done this way for the moment since it lets us update an existing resource in event of naming changes.
	// In the future, we should probably move to an explicit internal data structure instead
	// of relying on a library and define conversions.
	var q struct {
		Node struct {
			Organization struct {
				DatabaseId int64
			} `graphql:"... on Organization"`
		} `graphql:"node(id: $nodeId)"`
	}

	variables := map[string]interface{}{
		"nodeId": nodeId,
	}

	err := c.graphql.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}
	return c.GetOrganizationByDatabaseId(ctx, q.Node.Organization.DatabaseId)
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
