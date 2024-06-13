package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type delegateKeyMap struct {
	toggleDone key.Binding
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		toggleDone: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter", "toggle done"),
			key.WithHelp(" ", "toggle done"),
		),
	}
}
