package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	models = []tea.Model{
		NewModel(),
		NewForm(todo),
	}
	m := models[mainMenu]

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}

}
