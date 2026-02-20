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

func (c *Client) GetUserOrgs() ([]Org, error) {
	var orgs []Org
	err := c.rest.Get("user/orgs?per_page=100", &orgs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user orgs: %w", err)
	}
	return orgs, nil
}

func (c *Client) GetOrgRepos(org string) ([]OrgRepo, error) {
	var repos []OrgRepo
	err := c.rest.Get(fmt.Sprintf("orgs/%s/repos?sort=updated&per_page=100", org), &repos)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch org repos: %w", err)
	}
	return repos, nil
}
