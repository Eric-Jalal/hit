package github

import (
	"fmt"
	"net/url"
)

type runsResponse struct {
	TotalCount int           `json:"total_count"`
	Runs       []WorkflowRun `json:"workflow_runs"`
}

type jobsResponse struct {
	TotalCount int   `json:"total_count"`
	Jobs       []Job `json:"jobs"`
}

func (c *Client) GetRuns(branch string, perPage int) ([]WorkflowRun, error) {
	params := url.Values{}
	params.Set("branch", branch)
	params.Set("per_page", fmt.Sprintf("%d", perPage))

	var resp runsResponse
	err := c.rest.Get(c.endpoint("actions/runs")+"?"+params.Encode(), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workflow runs: %w", err)
	}
	return resp.Runs, nil
}

func (c *Client) GetJobs(runID int64) ([]Job, error) {
	var resp jobsResponse
	err := c.rest.Get(c.endpoint(fmt.Sprintf("actions/runs/%d/jobs", runID)), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %w", err)
	}
	return resp.Jobs, nil
}

func (c *Client) GetJobLog(jobID int64) (string, error) {
	var log string
	err := c.rest.Get(c.endpoint(fmt.Sprintf("actions/jobs/%d/logs", jobID)), &log)
	if err != nil {
		return "", fmt.Errorf("failed to fetch job log: %w", err)
	}
	return log, nil
}
