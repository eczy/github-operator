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
