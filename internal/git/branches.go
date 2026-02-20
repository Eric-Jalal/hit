package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (r *Repo) DefaultBranch() string {
	// Try refs/remotes/origin/HEAD symbolic ref
	ref, err := r.repo.Reference(plumbing.NewRemoteReferenceName("origin", "HEAD"), false)
	if err == nil {
		target := ref.Target()
		if target.IsRemote() {
			return strings.TrimPrefix(target.Short(), "origin/")
		}
	}

	// Fall back to checking origin/main then origin/master
	for _, name := range []string{"main", "master"} {
		_, err := r.repo.Reference(plumbing.NewRemoteReferenceName("origin", name), true)
		if err == nil {
			return name
		}
	}
	return ""
}

func (r *Repo) AheadBehind(refA, refB string) (int, int) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", refA+"..."+refB)
	cmd.Dir = r.path
	out, err := cmd.Output()
	if err != nil {
		return 0, 0
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0
	}
	ahead, _ := strconv.Atoi(parts[0])
	behind, _ := strconv.Atoi(parts[1])
	return ahead, behind
}

func (r *Repo) ListBranches() ([]Branch, error) {
	head, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("cannot read HEAD: %w", err)
	}
	currentBranch := head.Name().Short()

	refs, err := r.repo.References()
	if err != nil {
		return nil, err
	}

	// Build set of remote branch names
	remoteSet := make(map[string]bool)
	var localRefs []*plumbing.Reference
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name()
		if name.IsRemote() {
			short := strings.TrimPrefix(name.Short(), "origin/")
			if short != "HEAD" {
				remoteSet[short] = true
			}
		}
		if name.IsBranch() {
			localRefs = append(localRefs, ref)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	defaultBranch := r.DefaultBranch()

	var branches []Branch
	for _, ref := range localRefs {
		name := ref.Name().Short()
		hash := ref.Hash()

		commit, err := r.repo.CommitObject(hash)
		if err != nil {
			continue
		}

		b := Branch{
			Name:          name,
			Hash:          hash.String()[:7],
			Subject:       firstLine(commit.Message),
			Author:        commit.Author.Name,
			When:          commit.Author.When,
			IsCurrent:     name == currentBranch,
			HasRemote:     remoteSet[name],
			DefaultBranch: defaultBranch,
			IsDefault:     name == defaultBranch,
		}

		if b.HasRemote {
			b.RemoteAhead, b.RemoteBehind = r.AheadBehind(name, "origin/"+name)
		}

		if defaultBranch != "" && !b.IsDefault {
			b.DefaultAhead, b.DefaultBehind = r.AheadBehind(name, "origin/"+defaultBranch)
		}

		branches = append(branches, b)
	}

	sort.Slice(branches, func(i, j int) bool {
		if branches[i].IsCurrent {
			return true
		}
		if branches[j].IsCurrent {
			return false
		}
		return branches[i].When.After(branches[j].When)
	})

	return branches, nil
}

func (r *Repo) CurrentBranch() string {
	head, err := r.repo.Head()
	if err != nil {
		return ""
	}
	return head.Name().Short()
}

func (r *Repo) Checkout(branchName string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	localRef := plumbing.NewBranchReferenceName(branchName)
	_, err = r.repo.Reference(localRef, false)
	if err == nil {
		return w.Checkout(&gogit.CheckoutOptions{
			Branch: localRef,
		})
	}

	remoteRef := plumbing.NewRemoteReferenceName("origin", branchName)
	ref, err := r.repo.Reference(remoteRef, true)
	if err != nil {
		return fmt.Errorf("branch %q not found locally or in origin", branchName)
	}

	err = r.repo.Storer.SetReference(plumbing.NewHashReference(localRef, ref.Hash()))
	if err != nil {
		return err
	}

	return w.Checkout(&gogit.CheckoutOptions{
		Branch: localRef,
	})
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
