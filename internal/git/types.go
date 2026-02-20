package git

import "time"

type Branch struct {
	Name          string
	Hash          string
	Subject       string
	Author        string
	When          time.Time
	IsCurrent     bool
	IsRemote      bool
	HasRemote     bool
	RemoteAhead   int
	RemoteBehind  int
	DefaultAhead  int
	DefaultBehind int
	DefaultBranch string
	IsDefault     bool
}
