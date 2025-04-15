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
	catFact  string
	err      ErrorMsg
	fetching bool
}

func main() {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := tea.NewProgram(&Model{spinner: s, catFact: "Press `Enter`"})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

type ErrorMsg string

func (m *Model) Init() tea.Cmd { return m.spinner.Tick }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	if m.err != "" {
		return string(m.err)
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(50).
		Height(10).
		Align(lipgloss.Center, lipgloss.Center)

	if m.fetching {
		return style.Render(m.spinner.View())
	}

	return style.Render(m.catFact)
}

func (m *Model) GetCatFact() tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		resp, err := http.Get("https://meowfacts.herokuapp.com/")
		if err != nil {
			m.err = ErrorMsg(err.Error())
			return nil
		}

		var catData struct {
			Data []string `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&catData)
		if err != nil {
			m.err = ErrorMsg(err.Error())
			return nil
		}

		if time.Since(start) < time.Second*1 {
			time.Sleep(time.Second * 1)
		}

		m.catFact = catData.Data[0]
		m.fetching = false

		return m.catFact
	}
}
