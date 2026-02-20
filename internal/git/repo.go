package git

import (
	"fmt"
	"os"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

type Repo struct {
	repo *gogit.Repository
	path string
}

func Open(path string) (*Repo, error) {
	r, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}
	return &Repo{repo: r, path: path}, nil
}

func OpenCwd() (*Repo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return Open(cwd)
}

func (r *Repo) RemoteURL() string {
	remote, err := r.repo.Remote("origin")
	if err != nil {
		return ""
	}
	urls := remote.Config().URLs
	if len(urls) == 0 {
		return ""
	}
	return urls[0]
}

func (r *Repo) OwnerRepo() (string, string, error) {
	url := r.RemoteURL()
	if url == "" {
		return "", "", fmt.Errorf("no origin remote found")
	}

	url = strings.TrimSuffix(url, ".git")

	if strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("cannot parse remote URL: %s", url)
		}
		segments := strings.Split(parts[1], "/")
		if len(segments) < 2 {
			return "", "", fmt.Errorf("cannot parse remote URL: %s", url)
		}
		return segments[len(segments)-2], segments[len(segments)-1], nil
	}

	if strings.Contains(url, "github.com") {
		parts := strings.Split(url, "/")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("cannot parse remote URL: %s", url)
		}
		return parts[len(parts)-2], parts[len(parts)-1], nil
	}

	return "", "", fmt.Errorf("not a GitHub remote: %s", url)
}
