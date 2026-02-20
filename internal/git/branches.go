package git

import (
	"fmt"
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

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

	seen := make(map[string]bool)
	var branches []Branch

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name()
		if !name.IsBranch() && !name.IsRemote() {
			return nil
		}

		short := name.Short()
		isRemote := name.IsRemote()

		if isRemote {
			short = strings.TrimPrefix(short, "origin/")
			if short == "HEAD" {
				return nil
			}
			if seen[short] {
				return nil
			}
		}

		seen[short] = true

		hash := ref.Hash()
		commit, err := r.repo.CommitObject(hash)
		if err != nil {
			return nil
		}

		branches = append(branches, Branch{
			Name:      short,
			Hash:      hash.String()[:7],
			Subject:   firstLine(commit.Message),
			Author:    commit.Author.Name,
			When:      commit.Author.When,
			IsCurrent: short == currentBranch,
			IsRemote:  isRemote && !seen[short],
		})
		return nil
	})
	if err != nil {
		return nil, err
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
