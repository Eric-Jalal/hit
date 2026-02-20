package git

import "time"

type Branch struct {
	Name      string
	Hash      string
	Subject   string
	Author    string
	When      time.Time
	IsCurrent bool
	IsRemote  bool
}
