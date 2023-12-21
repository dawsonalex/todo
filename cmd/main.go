package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dawsonalex/todo"
	"os"
)

var (
	todoList = todo.NewListFromDeserialised("")

	docStyle          = lipgloss.NewStyle().Margin(1, 2)
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item struct {
	todo.Item
}

func (i item) Description() string { return i.Message }
func (i item) FilterValue() string { return i.Message }

func main() {
	todoList.Add(todo.Message("Go Shopping"), todo.Done())
	todoList.Add(todo.Message("Another item"))
	todoList.Add(todo.Message("Third Item"))

	listItems := todoList.GetAll()
	viewItems := make([]list.Item, len(listItems))
	for i, listItem := range listItems {
		viewItems[i] = item{listItem}
	}

	m := Model{list: list.New(viewItems, itemDelegate{}, 0, 0)}
	m.list.Title = "Todo"
	m.list.Styles.Title = titleStyle

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
