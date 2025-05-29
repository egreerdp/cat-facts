package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	spinner  spinner.Model
	message  string
	err      ErrorMsg
	fetching bool
	width    int
	height   int
}

func main() {
	s := spinner.New()
	s.Spinner = spinner.Monkey
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := tea.NewProgram(&Model{spinner: s, message: "Press `Enter`"}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %s", err.Error())
		os.Exit(1)
	}
}

type ErrorMsg string

func (m *Model) Init() tea.Cmd { return m.spinner.Tick }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.fetching {
				return m, nil
			}
			m.fetching = true
			return m, m.GetCatFact()
		}
	case ErrorMsg:
		m.err = msg
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(50).
		Height(10).
		Align(lipgloss.Center, lipgloss.Center)

	if m.err != "" {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, style.Render(string(m.err)))
	}

	var content string
	if m.fetching {
		content = m.spinner.View()
	} else {
		content = m.message
	}

	renderedContent := style.Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, renderedContent)
}

func (m *Model) GetCatFact() tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		resp, err := http.Get("https://meowfacts.herokuapp.com/")
		if err != nil {
			m.err = ErrorMsg(err.Error())
			return nil
		}

		var catData map[string]any
		err = json.NewDecoder(resp.Body).Decode(&catData)
		if err == nil {
			m.err = ErrorMsg("test error")
			return nil
		}

		if time.Since(start) < time.Second*1 {
			time.Sleep(time.Second * 1)
		}

		m.message = catData["data"].([]any)[0].(string)
		m.fetching = false

		return m.message
	}
}
