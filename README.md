# hit

A terminal UI for GitHub-integrated git workflows. Browse branches, monitor CI/CD runs, and manage your workflow without leaving the terminal.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Install

```
go install github.com/elisa-content-delivery/hit@latest
```

Or build from source:

```
git clone https://github.com/elisa-content-delivery/hit.git
cd hit
go build -o hit
```

## Usage

Run `hit` from any git repository with a GitHub remote:

```
cd your-repo
hit
```

### Authentication

hit looks for a GitHub token in this order:

1. `GH_TOKEN` environment variable
2. `GITHUB_TOKEN` environment variable
3. `gh` CLI auth cache (from `gh auth login`)
4. Manual token input (prompted on launch)

### Views

Navigate between views with `tab` / `shift+tab`.

**Branches** -- List local branches with remote and default branch sync status.

| Indicator | Meaning |
|-----------|---------|
| `* branch` | Current branch |
| `☁ =` | Synced with remote |
| `☁ ↑N` | N commits ahead of remote (need push) |
| `☁ ↓N` | N commits behind remote (need pull) |
| `☁ ↑N↓M` | Diverged from remote |
| `local` | No remote tracking branch |
| `= main` | Synced with default branch |
| `↑N main` | N commits ahead of default branch |

Keys: `enter` checkout, `r` refresh, `/` filter.

**CI** -- Monitor GitHub Actions workflow runs for the current branch. Drill down from runs to jobs to steps to logs.

Keys: `enter` drill in, `esc` back, `r` refresh.

**PRs** -- Coming soon.

**Reviews** -- Coming soon.

### Global Keys

| Key | Action |
|-----|--------|
| `tab` | Next view |
| `shift+tab` | Previous view |
| `q` | Quit |

## Requirements

- Go 1.25+
- A [Nerd Font](https://www.nerdfonts.com/) in your terminal for icons
- A git repository with a GitHub remote
- A GitHub token (via `gh` CLI, environment variable, or manual entry)

## License

MIT
