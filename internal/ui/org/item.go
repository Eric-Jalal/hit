package org

import (
	"fmt"

	gh "github.com/elisa-content-delivery/hit/internal/github"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

type orgItem struct{ org gh.Org }

func (o orgItem) Title() string {
	return styles.IconOrg + " " + o.org.Login
}
func (o orgItem) Description() string { return o.org.Description }
func (o orgItem) FilterValue() string { return o.org.Login }

type repoItem struct{ repo gh.OrgRepo }

func (r repoItem) Title() string {
	name := r.repo.Name
	var badges string
	if r.repo.Archived {
		badges += " " + styles.BadgeNeutral.Render("[archived]")
	}
	if r.repo.Private {
		badges += " " + styles.BadgePending.Render("[private]")
	}
	return fmt.Sprintf("%s%s", name, badges)
}
func (r repoItem) Description() string { return r.repo.Description }
func (r repoItem) FilterValue() string { return r.repo.Name }
