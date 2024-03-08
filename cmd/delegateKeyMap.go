package main

import "github.com/charmbracelet/bubbles/key"

type delegateKeyMap struct {
	toggleDone key.Binding
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		toggleDone: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("enter", "toggle done"),
			key.WithHelp("space", "toggle done"),
		),
	}
}
