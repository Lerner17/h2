package tui

import "github.com/charmbracelet/lipgloss"

func newStyles() styles {
	return styles{
		header:    lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
		headerDim: lipgloss.NewStyle().Foreground(lipgloss.Color("70")),
		meterOn:   lipgloss.NewStyle().Foreground(lipgloss.Color("46")),
		meterOff:  lipgloss.NewStyle().Foreground(lipgloss.Color("236")),

		tableHead: lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true),
		row:       lipgloss.NewStyle().Foreground(lipgloss.Color("120")),
		rowActive: lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("46")).Bold(true),

		panel:      lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("40")).Padding(0, 1),
		panelError: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("196")).Padding(0, 1),

		success: lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
		error:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		muted:   lipgloss.NewStyle().Foreground(lipgloss.Color("243")),

		hotkeys:     lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		hotkeyLabel: lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("81")).Bold(true).Padding(0, 1),
		hotkeyValue: lipgloss.NewStyle().Foreground(lipgloss.Color("120")),
	}
}
