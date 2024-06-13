package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dawsonalex/todo"
)

// model wraps a list.Model to use to easily render the view
// and keeps this in like with a todo.List which acts as the source of
// truth for everything, as well as handling deserialization.
type model struct {
	sourceList *todo.List
	list       list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter", "space":
			item, ok := m.list.SelectedItem().(item)
			if ok {
				// mark the item as done in the todoList
				// TODO: this doesn't cause the UI to update for some reason.
				item.Done = !item.Done
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func newModel(sourceList *todo.List) model {
	m := model{
		sourceList: sourceList,
		list:       list.New(todoListItemsToViewItems(sourceList.GetAll()), newItemDelegate(), 0, 0),
	}
	m.list.Title = "Todo"
	m.list.Styles.Title = titleStyle
	return m
}

// todoListItemsToViewItems
func todoListItemsToViewItems(listItems []todo.Item) []list.Item {
	viewItems := make([]list.Item, len(listItems))
	for i, listItem := range listItems {
		viewItems[i] = item{listItem}
	}
	return viewItems
}
