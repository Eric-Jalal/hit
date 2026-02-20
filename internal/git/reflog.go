package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type ReflogEntry struct {
	Selector string // HEAD@{0}
	Action   string // checkout, commit, merge, etc.
	Detail   string // human-friendly: "master → feature", "Fix bug", etc.
	TimeAgo  string // "2 hours ago"
}

func (r *Repo) GetReflog(limit int) ([]ReflogEntry, error) {
	cmd := exec.Command("git", "reflog", fmt.Sprintf("--format=%%gd|%%gs|%%ar"), "-n", fmt.Sprintf("%d", limit))
	cmd.Dir = r.path
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git reflog: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var entries []ReflogEntry
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}

		selector := parts[0]
		subject := parts[1]
		timeAgo := parts[2]

		action, detail := parseReflogSubject(subject)
		entries = append(entries, ReflogEntry{
			Selector: selector,
			Action:   action,
			Detail:   detail,
			TimeAgo:  timeAgo,
		})
	}
	return entries, nil
}

func parseReflogSubject(subject string) (string, string) {
	// Subject format: "action: detail"
	colonIdx := strings.Index(subject, ": ")
	if colonIdx < 0 {
		return subject, ""
	}

	action := subject[:colonIdx]
	detail := subject[colonIdx+2:]

	switch {
	case action == "checkout" && strings.HasPrefix(detail, "moving from "):
		// "moving from X to Y" → "X → Y"
		detail = strings.TrimPrefix(detail, "moving from ")
		if idx := strings.Index(detail, " to "); idx >= 0 {
			from := detail[:idx]
			to := detail[idx+4:]
			detail = from + " → " + to
		}
	case strings.HasPrefix(action, "merge "):
		// "merge branch" → action=merge, detail=branch (detail)
		branch := strings.TrimPrefix(action, "merge ")
		action = "merge"
		if detail != "" {
			detail = branch + " (" + strings.ToLower(detail) + ")"
		} else {
			detail = branch
		}
	case action == "commit (amend)":
		action = "commit (amend)"
	}

	return action, detail
}
