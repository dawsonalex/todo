package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"io"
	"strings"
)

var (
	term           = termenv.EnvColorProfile()
	selectedItemFg = termenv.Style{}.Foreground(term.Color("170")).Styled
	doneItemFg     = termenv.Style{}.Foreground(term.Color("#00ff00")).Styled
)

type itemDelegate struct {
	keys *delegateKeyMap
}

func (delegate itemDelegate) Height() int  { return 1 }
func (delegate itemDelegate) Spacing() int { return 0 }

func (delegate itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	_, ok := m.SelectedItem().(item)
	if !ok {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, delegate.keys.toggleDone):
			// TODO: This needs to toggle the item, I think

		}
	}
	return nil
}

func makeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(term.Color(color)).Styled
}

func (delegate itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Description)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	if i.Done {
		str = fmt.Sprintf("%s %s", str, doneItemFg("(done)"))
	}

	fmt.Fprint(w, fn(str))
}

func (delegate itemDelegate) ShortHelp() []key.Binding {
	return []key.Binding{
		delegate.keys.toggleDone,
	}
}

func (delegate itemDelegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			delegate.keys.toggleDone,
		},
	}
}

func newItemDelegate() list.ItemDelegate {
	return itemDelegate{keys: newDelegateKeyMap()}
}
