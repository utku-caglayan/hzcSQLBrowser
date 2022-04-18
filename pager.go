package main

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type StringResultMsg string
type TableResultMsg *sql.Rows
type model struct {
	content       string
	ready         bool
	keyboardFocus bool
	viewport      viewport.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case StringResultMsg:
		newVP := viewport.New(m.viewport.Width, m.viewport.Height)
		newVP.YPosition = m.viewport.Height
		newVP.SetContent(string(msg))
		m.viewport, cmd = newVP.Update(msg)
		return m, cmd
		m.viewport.SetContent(string(msg))
		m.viewport.GotoTop()
		return m, nil
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "esc" || k == "ctrl+q" {
			return m, tea.Quit
		}
		if msg.Type == tea.KeyTab {
			m.keyboardFocus = !m.keyboardFocus
			return m, nil
		}
		if !m.keyboardFocus {
			return m, nil
		}
	case tea.WindowSizeMsg:
		//headerHeight := lipgloss.Height(m.headerView())
		//footerHeight := lipgloss.Height(m.footerView())
		//verticalMarginHeight := headerHeight + footerHeight
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height)
			//m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	//return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	return m.viewport.View()
}
