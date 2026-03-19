package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name     string
	title    lipgloss.Style
	divider  lipgloss.Style
	selected lipgloss.Style
	normal   lipgloss.Style
	playing  lipgloss.Style
	help     lipgloss.Style
	input    lipgloss.Style
	err      lipgloss.Style
	viz      lipgloss.Style
	vizPause lipgloss.Style
	meta     lipgloss.Style
	filter   lipgloss.Style
	header   lipgloss.Style
	radio    lipgloss.Style
	dancer   lipgloss.Style
}

var themes = []Theme{
	themeDefault(),
	themeNord(),
	themeDracula(),
	themeEverforest(),
}

func themeDefault() Theme {
	return Theme{
		Name:     "Default",
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		divider:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		playing:  lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
		help:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		input:    lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
		err:      lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		viz:      lipgloss.NewStyle().Foreground(lipgloss.Color("82")),
		vizPause: lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		meta:     lipgloss.NewStyle().Foreground(lipgloss.Color("75")),
		filter:   lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true),
		header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")),
		radio:    lipgloss.NewStyle().Foreground(lipgloss.Color("178")),
		dancer:   lipgloss.NewStyle().Foreground(lipgloss.Color("213")),
	}
}

func themeNord() Theme {
	return Theme{
		Name:     "Nord",
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B48EAD")),
		divider:  lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Bold(true),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#D8DEE9")),
		playing:  lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")),
		help:     lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
		input:    lipgloss.NewStyle().Foreground(lipgloss.Color("#B48EAD")),
		err:      lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),
		viz:      lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
		vizPause: lipgloss.NewStyle().Foreground(lipgloss.Color("#D08770")),
		meta:     lipgloss.NewStyle().Foreground(lipgloss.Color("#81A1C1")),
		filter:   lipgloss.NewStyle().Foreground(lipgloss.Color("#8FBCBB")).Bold(true),
		header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5E81AC")),
		radio:    lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")),
		dancer:   lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")),
	}
}

func themeDracula() Theme {
	return Theme{
		Name:     "Dracula",
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")),
		divider:  lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Bold(true),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")),
		playing:  lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")),
		help:     lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		input:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")),
		err:      lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
		viz:      lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
		vizPause: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")),
		meta:     lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
		filter:   lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Bold(true),
		header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")),
		radio:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")),
		dancer:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")),
	}
}

func themeEverforest() Theme {
	return Theme{
		Name:     "Everforest",
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A7C080")),
		divider:  lipgloss.NewStyle().Foreground(lipgloss.Color("#859289")),
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#83C092")).Bold(true),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#D3C6AA")),
		playing:  lipgloss.NewStyle().Foreground(lipgloss.Color("#DBBC7F")),
		help:     lipgloss.NewStyle().Foreground(lipgloss.Color("#859289")),
		input:    lipgloss.NewStyle().Foreground(lipgloss.Color("#A7C080")),
		err:      lipgloss.NewStyle().Foreground(lipgloss.Color("#E67E80")),
		viz:      lipgloss.NewStyle().Foreground(lipgloss.Color("#A7C080")),
		vizPause: lipgloss.NewStyle().Foreground(lipgloss.Color("#E69875")),
		meta:     lipgloss.NewStyle().Foreground(lipgloss.Color("#7FBBB3")),
		filter:   lipgloss.NewStyle().Foreground(lipgloss.Color("#83C092")).Bold(true),
		header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7FBBB3")),
		radio:    lipgloss.NewStyle().Foreground(lipgloss.Color("#DBBC7F")),
		dancer:   lipgloss.NewStyle().Foreground(lipgloss.Color("#D699B6")),
	}
}
