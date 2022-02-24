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
	"github.com/mathaou/termdbms/viewer"

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

type table struct {
	termdbmsTable viewer.TuiModel
	keyboardFocus bool
}

func (t *table) Init() tea.Cmd {
	t.termdbmsTable = viewer.GetNewModel("", nil)
	return t.termdbmsTable.Init()
}

func (t *table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case StringResultMsg:
		// update table
		t.termdbmsTable.UI.RenderSelection = true
		t.termdbmsTable.Data().EditTextBuffer = string(m)
		return t, nil
	case TableResultMsg:
		i := 0
		t.termdbmsTable.UI.RenderSelection = false
		t.termdbmsTable.Data().EditTextBuffer = ""
		t.termdbmsTable.QueryData = &viewer.UIData{
			TableHeaders:      make(map[string][]string),
			TableHeadersSlice: []string{},
			TableSlices:       make(map[string][]interface{}),
			TableIndexMap:     make(map[int]string),
		}
		t.termdbmsTable.QueryResult = &viewer.TableState{ // perform query
			Database: t.termdbmsTable.Table().Database,
			Data:     make(map[string]interface{}),
		}
		t.termdbmsTable.PopulateDataForResult(m, &i, "0")
		t.termdbmsTable.UI.CurrentTable = 1
		_ = t.termdbmsTable.NumHeaders() // to set maxHeaders global var, for side effect
		t.termdbmsTable.SetViewSlices()
		return t, nil
	case tea.KeyMsg:
		if m.Type == tea.KeyCtrlT {
			t.keyboardFocus = !t.keyboardFocus
			return t, nil
		}
		if !t.keyboardFocus {
			return t, nil
		}
	case tea.WindowSizeMsg:
		if m.Height > 0 {
			m.Height += 1 // to eliminate termdbs global sizing staff (footer height).
		}
		msg = m
	}
	tmp, cmd := t.termdbmsTable.Update(msg)
	t.termdbmsTable = tmp.(viewer.TuiModel)
	return t, cmd
}

func (t *table) View() string {
	//return viewer.AssembleTable(&t.termdbmsTable)
	done := make(chan bool, 2)
	defer close(done) // close
	var header, content string
	// body
	go func(c *string) {
		*c = viewer.AssembleTable(&t.termdbmsTable)
		done <- true
	}(&content)

	// header
	go viewer.HeaderAssembly(&t.termdbmsTable, &header, &done)
	<-done
	<-done
	if content == "" {
		content = strings.Repeat("\n", t.termdbmsTable.Viewport.Height)
	}
	return fmt.Sprintf("%s\n%s", header, content)
}

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

type Separator int

func (s Separator) Init() tea.Cmd {
	return nil
}

func (s Separator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s = Separator(msg.Width - 2)
	}
	return s, nil
}

func (s Separator) View() string {
	return strings.Repeat("─", max(0, int(s)))
}

func (c controller) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case textarea.SubmitMsg:
		return c, func() tea.Msg {
			lt := strings.TrimSpace(string(m))
			var w bytes.Buffer
			//if _, err := execSQL(c.client, lt, &w); err != nil {
			//	w.WriteString(err.Error())
			//}
			rows, err := execSQL(c.client, lt, &w)
			if err != nil {
				return StringResultMsg(err.Error())
			}
			return TableResultMsg(rows)
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
	var s Separator
	//pager := &model{}
	//var termdbmsModel viewer.TuiModel
	textArea := textarea.InitTextArea()
	c := &controller{vertical.InitialModel([]tea.Model{
		Heading{
			title: "Result",
			align: lipgloss.Center,
		},
		&table{},
		s,
		textArea,
		Heading{
			title: "Execute C-x  Quit C-q/C-c  Toggle Focus C-t",
			align: lipgloss.Left,
		},
	}, []int{-1, 3, -1, 1, -1}), client}
	p := tea.NewProgram(
		c,
		tea.WithMouseCellMotion(), // turn on mouse support, so we can track the mouse wheel
	)
	if err := p.Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
