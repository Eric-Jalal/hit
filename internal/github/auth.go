package github

import (
	"os"

	"github.com/cli/go-gh/v2/pkg/auth"
)

type AuthStatus struct {
	Authenticated bool
	Token         string
	Source        string
}

func DetectAuth() AuthStatus {
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return AuthStatus{Authenticated: true, Token: token, Source: "GH_TOKEN"}
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return AuthStatus{Authenticated: true, Token: token, Source: "GITHUB_TOKEN"}
	}

	token, source := auth.TokenForHost("github.com")
	if token != "" {
		return AuthStatus{Authenticated: true, Token: token, Source: source}
	}

	return AuthStatus{}
}
