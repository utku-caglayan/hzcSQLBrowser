package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"

	"browser/layout/vertical"
	"browser/textarea"
)

type controller struct {
	tea.Model
	client *hazelcast.Client
}

var HeadingStyle = func() lipgloss.Style {
	b := lipgloss.NormalBorder()
	return lipgloss.NewStyle().BorderStyle(b)
}()

type Heading struct {
	width int
	title string
	align lipgloss.Position
}

func (h Heading) Init() tea.Cmd {
	return nil
}

func (h Heading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		h.width = msg.Width - 2
	}
	return h, nil
}

func (h Heading) View() string {
	return HeadingStyle.Width(h.width).Align(h.align).Render(h.title)
}

func (c controller) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case textarea.SubmitMsg:
		return c, func() tea.Msg {
			lt := strings.ToLower(string(m))
			lt = strings.TrimSpace(lt)
			var w bytes.Buffer
			if err := execSQL(c.client, lt, &w); err != nil {
				w.WriteString(err.Error())
			}
			return UpdatePagerMsg(w.String())
		}
	}
	var cmd tea.Cmd
	c.Model, cmd = c.Model.Update(msg)
	return c, cmd
}

func main() {
	cnfg := hazelcast.NewConfig()
	cnfg.Logger.Level = logger.OffLevel
	client, err := hazelcast.StartNewClientWithConfig(context.Background(), cnfg)
	if err != nil {
		panic(fmt.Sprint("cannot start hzc client", err))
	}
	pager := &model{}
	textArea := textarea.InitTextArea()
	c := &controller{vertical.InitialModel([]tea.Model{Heading{
		title: "Query Editor",
		align: lipgloss.Center,
	},
		textArea,
		Heading{
			title: "Result",
			align: lipgloss.Center,
		},
		pager,
		Heading{
			title: "Execute Query: ctrl+x  |  Quit: ctrl+q/ctrl+c  | Toggle Focus: ctrl+t",
			align: lipgloss.Left,
		},
	}, []int{-1, 1, -1, 3, -1}), client}
	p := tea.NewProgram(
		c,
		//model{content: string(content)},
		//tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support, so we can track the mouse wheel
	)
	if err := p.Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
