package textarea

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type SubmitMsg string

type model struct {
	viewport      viewport.Model
	textInput     textinput.Model
	err           error
	ready         bool
	keyboardFocus bool
}

func InitTextArea() *model {
	ti := textinput.New()
	ti.Placeholder = "sql query here"
	ti.Focus()
	return &model{
		textInput:     ti,
		err:           nil,
		keyboardFocus: true,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//headerHeight := lipgloss.Height(m.headerView())
	//footerHeight := lipgloss.Height(m.footerView())
	//verticalMarginHeight := headerHeight //+ footerHeight
	switch tmsg := msg.(type) {
	case tea.KeyMsg:
		if tmsg.Type == tea.KeyCtrlT {
			var tmp tea.Cmd
			if m.keyboardFocus {
				m.textInput.Blur()
			} else {
				tmp = m.textInput.Focus()
			}
			m.keyboardFocus = !m.keyboardFocus
			return m, tmp
		}
		if !m.keyboardFocus {
			return m, nil
		}
		switch tmsg.Type {
		case tea.KeyEnter:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("¬"), Alt: false}
		case tea.KeyCtrlX:
			//m.textInput.Blur()
			//m.onSubmit(strings.Replace(m.textInput.Value(), "¬", "\n", -1))
			//return m, nil
			return m, func() tea.Msg {
				return SubmitMsg(strings.Replace(m.textInput.Value(), "¬", "\n", -1))
			}
		case tea.KeyRunes:
			for i, r := range tmsg.Runes {
				if r == '\n' {
					tmsg.Runes[i] = '¬'
				}
			}
		}
	case tea.WindowSizeMsg:
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(tmsg.Width, tmsg.Height)
			m.viewport.SetContent(m.textInput.View())
			m.ready = true
		} else {
			m.viewport.Width = tmsg.Width
			m.viewport.Height = tmsg.Height
		}
	}
	var cmd1 tea.Cmd
	m.textInput, cmd1 = m.textInput.Update(msg)
	m.viewport.SetContent(strings.Replace(m.textInput.View(), "¬", "\n", -1))
	m.viewport.GotoBottom()
	return m, tea.Batch(cmd1)
}

func (m model) View() string {
	return m.viewport.View()
	//return strings.Replace(m.textInput.View(), "¬", "\n", -1)
}
