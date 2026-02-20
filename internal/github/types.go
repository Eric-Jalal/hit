package github

import "time"

type WorkflowRun struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion"`
	HeadBranch   string    `json:"head_branch"`
	HeadSHA      string    `json:"head_sha"`
	Event        string    `json:"event"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	HTMLURL      string    `json:"html_url"`
	RunNumber    int       `json:"run_number"`
	WorkflowID   int64     `json:"workflow_id"`
	RunAttempt   int       `json:"run_attempt"`
}

type Job struct {
	ID         int64     `json:"id"`
	RunID      int64     `json:"run_id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	StartedAt  time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Steps      []Step    `json:"steps"`
	HTMLURL    string    `json:"html_url"`
}

type Step struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Conclusion  string    `json:"conclusion"`
	Number      int       `json:"number"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

type ErrorAnnotation struct {
	File    string
	Line    int
	Column  int
	Message string
}

type Org struct {
	Login       string `json:"login"`
	Description string `json:"description"`
}

type OrgRepo struct {
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	SSHURL      string    `json:"ssh_url"`
	Private     bool      `json:"private"`
	Archived    bool      `json:"archived"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PR struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
	Head    struct {
		Ref string `json:"ref"`
	} `json:"head"`
}
