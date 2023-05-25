package main

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type status int

var ErrInvalidTask = errors.New("invalid task")

const divisor = 4

const (
	todo status = iota
	inProgress
	done
)

/* styling */

var (
	columnStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.HiddenBorder())

	focusedStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			Foreground(lipgloss.Color("205")).
			BorderForeground(lipgloss.Color("62"))
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)


/* Task model */

type Task struct {
	status      status
	title       string
	description string
}

func (t Task) FilterValue() string {
	return t.title
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Description() string {
	return t.description
}
func NewTask(status status, title,description string) Task {
	return Task{
		status: status,
		title: title,
		description: description,
	}
}
func (t *Task) Next() {
	if t.status < done {
		t.status++
	} else {
		t.status = todo
	}
}

/* Model Management */

var models []tea.Model

const (
	mainMenu status = iota
	form
)

/*Main model */

type Model struct {
	focused  status
	lists    []list.Model
	err      error
	loaded   bool
	quitting bool
}

func NewModel() *Model {
	return &Model{}
}

func (m *Model) initLists(width, height int) {
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width/divisor, height/2)
	defaultList.SetShowHelp(false)
	m.lists = []list.Model{defaultList, defaultList, defaultList}
	// init lists
	m.lists[todo].Title = "Todo"
	m.lists[inProgress].Title = "In Progress"
	m.lists[done].Title = "Done"
	m.lists[todo].SetItems([]list.Item{
		Task{
			status:      todo,
			title:       "Task 1",
			description: "Description 1",
		},
		Task{
			status:      todo,
			title:       "Task 2",
			description: "Description 2",
		},
	})
	// init in progress list
	m.lists[inProgress].SetItems([]list.Item{
		Task{
			status:      inProgress,
			title:       "Task 3",
			description: "Description 3",
		},
		Task{
			status:      inProgress,
			title:       "Task 4",
			description: "Description 4",
		},
	})
	// init done list
	m.lists[done].SetItems([]list.Item{
		Task{
			status:      done,
			title:       "Task 5",
			description: "Description 5",
		},
		Task{
			status:      done,
			title:       "Task 6",
			description: "Description 6",
		},
	})
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Next() {
	if m.focused < done {
		m.focused++
	} else {
		m.focused = todo
	}
}
func (m *Model) Prev() {
	if m.focused > todo {
		m.focused--
	} else {
		m.focused = done
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.loaded {
			columnStyle.Width(msg.Width / divisor)
			focusedStyle.Width(msg.Width / divisor)
			columnStyle.Height(msg.Height  - divisor)
			focusedStyle.Height(msg.Height - divisor)
			m.initLists(msg.Width, msg.Height)
			for i, list := range m.lists {
				list.SetSize(msg.Width/divisor, msg.Height/2)
				m.lists[i], _ = list.Update(msg)
			}
			m.loaded = true
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "left":
			m.Prev()
		case "right":
			m.Next()
		case "enter":
			m.MoveTask()
		case "n":
			models[mainMenu] = m
			models[form] = NewForm(m.focused)
			return models[form].Update(nil)
		}
		case Task:
			task := msg
			m.lists[task.status].InsertItem(len(m.lists[m.focused].Items())-1, task)
			return m, nil

	}
	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	return m, cmd
}


func (m *Model ) MoveTask() tea.Msg {

	selectedItem := m.lists[m.focused].SelectedItem()

	taskIdx := m.lists[m.focused].Index()
	
	if selectedItem == nil {
		m.err = ErrInvalidTask
		return nil
	}
	task, ok := selectedItem.(Task)
	if !ok {
		m.err = ErrInvalidTask
		return nil
	}


	task.Next()

	m.lists[m.focused].RemoveItem(taskIdx)
	m.lists[task.status].InsertItem(len(m.lists[task.status].Items())-1, task)
	return tea.KeyMsg{Type: tea.KeyEnter}
}



func (m Model) View() string {
	if !m.loaded {
		return "Loading..."
	}
	if m.quitting {
		return ""
	}
	todoView := m.lists[todo].View()
	inProgressView := m.lists[inProgress].View()
	doneView := m.lists[done].View()
	switch m.focused {
	case inProgress:
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			columnStyle.Render(todoView),
			focusedStyle.Render(inProgressView),
			columnStyle.Render(doneView),
		)
	case done:
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			columnStyle.Render(todoView),
			columnStyle.Render(inProgressView),
			focusedStyle.Render(doneView),
		)
	default:
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			focusedStyle.Render(todoView),
			columnStyle.Render(inProgressView),
			columnStyle.Render(doneView),
		)
	}
}

/* Form model */

type Form struct {
	focused status
	title textinput.Model
	description textarea.Model
}

func (f Form) Init() tea.Cmd {
	return nil
}

func (f Form) CreateTask() tea.Msg {
	return NewTask(f.focused, f.title.Value(), f.description.Value())
}

func (f Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "ctrl+c":
				return f, tea.Quit
			case "enter":
				if f.title.Focused() {
					f.title.Blur()
					f.description.Focus()
					return f, textarea.Blink
				} else {
					models[form] = f
					return models[mainMenu], f.CreateTask
				}
		}
	}
	if f.title.Focused() {
		f.title, cmd = f.title.Update(msg)
		return f, cmd
	} else {
		f.description, cmd = f.description.Update(msg)
		return f, cmd
	}
}

func (m Form) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#E7E7E7")).Render("Title:"),
		m.title.View(),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#E7E7E7")).Render("Description:"),
		m.description.View(),
	)
}

func NewForm(focused status) *Form {
	form:= &Form{
		title: textinput.New(),
		description: textarea.New(),
		focused: focused,
	}
	form.title.Focus()
	return form
}