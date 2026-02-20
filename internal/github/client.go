package github

import (
	"fmt"

	ghAPI "github.com/cli/go-gh/v2/pkg/api"
)

type Client struct {
	rest  *ghAPI.RESTClient
	owner string
	repo  string
}

func NewClient(owner, repo, token string) (*Client, error) {
	opts := ghAPI.ClientOptions{
		AuthToken: token,
	}
	rest, err := ghAPI.NewRESTClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}
	return &Client{rest: rest, owner: owner, repo: repo}, nil
}

func (c *Client) endpoint(path string) string {
	return fmt.Sprintf("repos/%s/%s/%s", c.owner, c.repo, path)
}
