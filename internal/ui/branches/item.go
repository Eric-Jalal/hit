package branches

import (
	"fmt"

	"github.com/elisa-content-delivery/hit/internal/git"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

type branchItem struct {
	branch git.Branch
}

func (b branchItem) Title() string {
	prefix := "  "
	if b.branch.IsCurrent {
		prefix = styles.HighlightStyle.Render("* ")
	}
	return prefix + b.branch.Name
}

func (b branchItem) Description() string {
	return fmt.Sprintf("%s %s — %s",
		styles.SubtitleStyle.Render(b.branch.Hash),
		b.branch.Subject,
		styles.SubtitleStyle.Render(b.branch.Author),
	)
}

func (b branchItem) FilterValue() string {
	return b.branch.Name
}
