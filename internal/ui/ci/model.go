package ci

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gh "github.com/elisa-content-delivery/hit/internal/github"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

type pane int

const (
	paneRuns pane = iota
	paneJobs
	paneSteps
	paneLogs
)

type runsLoadedMsg struct {
	runs []gh.WorkflowRun
	err  error
}

type jobsLoadedMsg struct {
	jobs []gh.Job
	err  error
}

type logLoadedMsg struct {
	log string
	err error
}

type runItem struct{ run gh.WorkflowRun }

func (r runItem) Title() string {
	badge := statusBadge(r.run.Conclusion, r.run.Status)
	return fmt.Sprintf("%s %s #%d", badge, r.run.Name, r.run.RunNumber)
}

func (r runItem) Description() string {
	return styles.SubtitleStyle.Render(r.run.HeadSHA[:7]) + " " + r.run.CreatedAt.Format("Jan 02 15:04")
}
func (r runItem) FilterValue() string { return r.run.Name }

type jobItem struct{ job gh.Job }

func (j jobItem) Title() string {
	return statusBadge(j.job.Conclusion, j.job.Status) + " " + j.job.Name
}

func (j jobItem) Description() string {
	if j.job.CompletedAt.IsZero() {
		return "running..."
	}
	d := j.job.CompletedAt.Sub(j.job.StartedAt)
	return fmt.Sprintf("took %s", d.Round(1e9))
}
func (j jobItem) FilterValue() string { return j.job.Name }

type stepItem struct{ step gh.Step }

func (s stepItem) Title() string {
	return statusBadge(s.step.Conclusion, s.step.Status) + " " + s.step.Name
}
func (s stepItem) Description() string { return "" }
func (s stepItem) FilterValue() string { return s.step.Name }

type Model struct {
	client      *gh.Client
	branch      string
	currentPane pane
	runsList    list.Model
	jobsList    list.Model
	stepsList   list.Model
	logView     LogView
	spinner     spinner.Model
	loading     bool
	selectedRun *gh.WorkflowRun
	selectedJob *gh.Job
	annotations []gh.ErrorAnnotation
	width       int
	height      int
	status      string
}

func New(client *gh.Client, branch string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	makeList := func(title string) list.Model {
		l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
		l.Title = title
		l.SetShowHelp(false)
		l.SetFilteringEnabled(false)
		l.Styles.Title = styles.TitleStyle
		return l
	}

	return Model{
		client:      client,
		branch:      branch,
		currentPane: paneRuns,
		runsList:    makeList("Workflow Runs"),
		jobsList:    makeList("Jobs"),
		stepsList:   makeList("Steps"),
		logView:     NewLogView(),
		spinner:     s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadRuns)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listHeight := msg.Height - 4
		m.runsList.SetSize(msg.Width, listHeight)
		m.jobsList.SetSize(msg.Width, listHeight)
		m.stepsList.SetSize(msg.Width, listHeight)
		var cmd tea.Cmd
		m.logView, cmd = m.logView.Update(msg)
		return m, cmd

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case runsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %s", msg.err)
			return m, nil
		}
		items := make([]list.Item, len(msg.runs))
		for i, r := range msg.runs {
			items[i] = runItem{run: r}
		}
		m.runsList.SetItems(items)
		return m, nil

	case jobsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %s", msg.err)
			m.currentPane = paneRuns
			return m, nil
		}
		items := make([]list.Item, len(msg.jobs))
		for i, j := range msg.jobs {
			items[i] = jobItem{job: j}
		}
		m.jobsList.SetItems(items)
		m.currentPane = paneJobs
		return m, nil

	case logLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %s", msg.err)
			m.currentPane = paneSteps
			return m, nil
		}
		m.annotations = ParseAnnotations(msg.log)
		m.logView.SetContent(msg.log)
		m.currentPane = paneLogs
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m.goBack()

		case "enter":
			return m.drillDown()

		case "r":
			if m.currentPane == paneRuns {
				m.loading = true
				m.status = ""
				return m, tea.Batch(m.spinner.Tick, m.loadRuns)
			}
		}
	}

	var cmd tea.Cmd
	switch m.currentPane {
	case paneRuns:
		m.runsList, cmd = m.runsList.Update(msg)
	case paneJobs:
		m.jobsList, cmd = m.jobsList.Update(msg)
	case paneSteps:
		m.stepsList, cmd = m.stepsList.Update(msg)
	case paneLogs:
		m.logView, cmd = m.logView.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.loading {
		return m.spinner.View() + " Loading..."
	}

	var content string
	switch m.currentPane {
	case paneRuns:
		content = m.runsList.View()
	case paneJobs:
		content = m.jobsList.View()
	case paneSteps:
		content = m.stepsList.View()
	case paneLogs:
		header := styles.TitleStyle.Render("Job Log")
		if len(m.annotations) > 0 {
			header += styles.ErrorLineStyle.Render(fmt.Sprintf("  %d error(s)", len(m.annotations)))
		}
		content = header + "\n" + m.logView.View()
	}

	if m.status != "" {
		content += "\n" + lipgloss.NewStyle().MarginLeft(2).Render(m.status)
	}

	nav := m.breadcrumb()
	return nav + "\n" + content
}

func (m Model) breadcrumb() string {
	parts := []string{styles.SubtitleStyle.Render("CI")}
	if m.selectedRun != nil {
		parts = append(parts, styles.HighlightStyle.Render(m.selectedRun.Name))
	}
	if m.selectedJob != nil {
		parts = append(parts, styles.HighlightStyle.Render(m.selectedJob.Name))
	}
	if m.currentPane == paneLogs {
		parts = append(parts, styles.HighlightStyle.Render("log"))
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += styles.SubtitleStyle.Render(" > ") + p
	}
	return result
}

func (m Model) goBack() (Model, tea.Cmd) {
	switch m.currentPane {
	case paneJobs:
		m.currentPane = paneRuns
		m.selectedRun = nil
	case paneSteps:
		m.currentPane = paneJobs
		m.selectedJob = nil
	case paneLogs:
		m.currentPane = paneSteps
	}
	m.status = ""
	return m, nil
}

func (m Model) drillDown() (Model, tea.Cmd) {
	switch m.currentPane {
	case paneRuns:
		selected, ok := m.runsList.SelectedItem().(runItem)
		if !ok {
			return m, nil
		}
		m.selectedRun = &selected.run
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadJobs(selected.run.ID))

	case paneJobs:
		selected, ok := m.jobsList.SelectedItem().(jobItem)
		if !ok {
			return m, nil
		}
		m.selectedJob = &selected.job
		items := make([]list.Item, len(selected.job.Steps))
		for i, s := range selected.job.Steps {
			items[i] = stepItem{step: s}
		}
		m.stepsList.SetItems(items)
		m.currentPane = paneSteps
		return m, nil

	case paneSteps:
		if m.selectedJob == nil {
			return m, nil
		}
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadLog(m.selectedJob.ID))
	}
	return m, nil
}

func (m Model) loadRuns() tea.Msg {
	runs, err := m.client.GetRuns(m.branch, 20)
	return runsLoadedMsg{runs: runs, err: err}
}

func (m Model) loadJobs(runID int64) tea.Cmd {
	return func() tea.Msg {
		jobs, err := m.client.GetJobs(runID)
		return jobsLoadedMsg{jobs: jobs, err: err}
	}
}

func (m Model) loadLog(jobID int64) tea.Cmd {
	return func() tea.Msg {
		log, err := m.client.GetJobLog(jobID)
		return logLoadedMsg{log: log, err: err}
	}
}

func statusBadge(conclusion, status string) string {
	if status == "in_progress" || status == "queued" {
		return styles.BadgePending.Render(styles.IconPending)
	}
	switch conclusion {
	case "success":
		return styles.BadgeSuccess.Render(styles.IconCheck)
	case "failure":
		return styles.BadgeFailure.Render(styles.IconCross)
	case "cancelled":
		return styles.BadgeNeutral.Render(styles.IconStop)
	case "skipped":
		return styles.BadgeNeutral.Render(styles.IconSkip)
	default:
		return styles.BadgeNeutral.Render("?")
	}
}
