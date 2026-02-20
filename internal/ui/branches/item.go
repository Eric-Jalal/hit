package branches

import (
	"fmt"
	"strings"

	"github.com/elisa-content-delivery/hit/internal/git"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

type branchItem struct {
	branch git.Branch
}

func (b branchItem) Title() string {
	prefix := "  "
	if b.branch.IsCurrent {
		prefix = styles.HighlightStyle.Render(styles.IconBranch+" ")
	}
	return prefix + b.branch.Name
}

func (b branchItem) Description() string {
	parts := []string{
		fmt.Sprintf("%s %s", styles.SubtitleStyle.Render(b.branch.Hash), b.branch.Subject),
	}

	parts = append(parts, remoteStatus(b.branch))

	if ds := defaultStatus(b.branch); ds != "" {
		parts = append(parts, ds)
	}

	return strings.Join(parts, " · ")
}

func (b branchItem) FilterValue() string {
	return b.branch.Name
}

func remoteStatus(b git.Branch) string {
	if !b.HasRemote {
		return styles.SubtitleStyle.Render(styles.IconLocal + " local")
	}
	return styles.BadgeSuccess.Render(styles.IconCloud + " " + formatAheadBehind(b.RemoteAhead, b.RemoteBehind))
}

func defaultStatus(b git.Branch) string {
	if b.IsDefault || b.DefaultBranch == "" {
		return ""
	}
	ab := formatAheadBehind(b.DefaultAhead, b.DefaultBehind)
	return styles.SubtitleStyle.Render(ab + " " + b.DefaultBranch)
}

func formatAheadBehind(ahead, behind int) string {
	switch {
	case ahead == 0 && behind == 0:
		return styles.IconCheck
	case behind == 0:
		return fmt.Sprintf("%s%d", styles.IconArrowUp, ahead)
	case ahead == 0:
		return fmt.Sprintf("%s%d", styles.IconArrowDn, behind)
	default:
		return fmt.Sprintf("%s%d%s%d", styles.IconArrowUp, ahead, styles.IconArrowDn, behind)
	}
}
