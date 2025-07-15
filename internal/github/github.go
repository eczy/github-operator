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
	"net/http"

	"github.com/google/go-github/v60/github"
	"github.com/shurcooL/githubv4"
)

type Client struct {
	rest    *github.Client
	graphql *githubv4.Client
}

type ClientOption = func(*Client) error

func WithRoundTripper(rt http.RoundTripper) ClientOption {
	return func(c *Client) error {
		c.rest = github.NewClient(&http.Client{
			Transport: rt,
		})
		c.graphql = githubv4.NewClient(&http.Client{
			Transport: rt,
		})
		return nil
	}
}

func WithHttpClient(client *http.Client) ClientOption {
	return func(c *Client) error {
		c.rest = github.NewClient(client)
		c.graphql = githubv4.NewClient(client)
		return nil
	}
}

func NewClient(opts ...ClientOption) (*Client, error) {
	client := &Client{
		rest:    github.NewClient(nil),
		graphql: githubv4.NewClient(nil),
	}

	for _, opt := range opts {
		err := opt(client)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}
