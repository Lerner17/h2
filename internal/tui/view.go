package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	header := m.renderHeader()
	var body string
	switch m.state {
	case stateUsers:
		body = m.renderUsers()
	case stateAddInput:
		body = m.renderAddInput()
	case stateUserActions:
		body = m.renderUserActions()
	case stateResult:
		body = m.renderResult()
	case stateConnection:
		body = m.renderConnection()
	}
	footer := m.renderFooter()

	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	if m.width > 0 && m.height > 0 {
		contentHeight := lipgloss.Height(content)
		footerHeight := lipgloss.Height(footer)
		spacerLines := max(0, m.height-contentHeight-footerHeight)
		spacer := ""
		if spacerLines > 0 {
			spacer = strings.Repeat("\n", spacerLines)
		}
		layout := lipgloss.JoinVertical(lipgloss.Left, header, body, spacer, footer)
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, layout)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m model) renderHeader() string {
	mode := "USERS"
	switch m.state {
	case stateAddInput:
		mode = "ADD"
	case stateUserActions:
		mode = "ACTIONS"
	case stateResult:
		mode = "RESULT"
	case stateConnection:
		mode = "CONNECTION"
	}
	usersCount := len(m.users)
	meter := renderMeter(m.styles, usersCount)
	onlineCount, totalRx, totalTx := m.aggregateStats()

	line1 := m.styles.header.Render("HY2-CTL") + " " + m.styles.headerDim.Render("mode=") + m.styles.header.Render(mode)
	line2 := m.styles.headerDim.Render("users") + " " + meter + "  " + m.styles.headerDim.Render(fmt.Sprintf("count=%d online=%d rx=%s tx=%s", usersCount, onlineCount, formatBytes(totalRx), formatBytes(totalTx)))
	line3 := m.styles.headerDim.Render("a:add  f2:add  f5:refresh  enter/f6:actions  f10:quit")
	return lipgloss.JoinVertical(lipgloss.Left, line1, line2, line3)
}

func renderMeter(s styles, n int) string {
	maxBars := 20
	on := min(maxBars, n)
	off := maxBars - on
	return s.meterOn.Render(strings.Repeat("|", on)) + s.meterOff.Render(strings.Repeat(".", off))
}

func (m model) renderUsers() string {
	panelWidth := m.contentWidth()
	if m.loading {
		return m.styles.panel.Copy().Width(panelWidth).Render("Loading users...")
	}
	idxW := 4
	onlineW := 7
	rxW := 9
	txW := 9
	totalW := 9
	userW := max(8, panelWidth-idxW-onlineW-rxW-txW-totalW-10)
	head := m.styles.tableHead.Render(
		fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s", idxW, "ID", userW, "USER", onlineW, "ONLINE", rxW, "RX", txW, "TX", totalW, "TOTAL"),
	)
	rows := []string{head}

	if len(m.users) == 0 {
		rows = append(rows, m.styles.muted.Render("-- no users --  (press F2 or A to add)"))
	} else {
		for i, u := range m.users {
			stat := m.userStats[u]
			online := "no"
			if stat.Online {
				online = "yes"
			}
			line := fmt.Sprintf(
				"%-*d %-*s %-*s %-*s %-*s %-*s",
				idxW, i+1,
				userW, truncate(u, userW),
				onlineW, online,
				rxW, formatBytes(stat.RxBytes),
				txW, formatBytes(stat.TxBytes),
				totalW, formatBytes(stat.TotalBytes),
			)
			if i == m.usersCursor {
				rows = append(rows, m.styles.rowActive.Render(line))
			} else {
				rows = append(rows, m.styles.row.Render(line))
			}
		}
	}
	return m.styles.panel.Copy().Width(panelWidth).Render(strings.Join(rows, "\n"))
}

func (m model) renderAddInput() string {
	panelWidth := m.contentWidth()
	body := strings.Join([]string{
		"Command: add-user",
		"",
		"Enter username:",
		m.input.View(),
	}, "\n")
	return m.styles.panel.Copy().Width(panelWidth).Render(body)
}

func (m model) renderUserActions() string {
	panelWidth := m.contentWidth()
	lines := []string{m.styles.tableHead.Render("User: " + m.selectedUser)}
	for i := range m.actions {
		line := fmt.Sprintf("%-26s  %s", m.actions[i], m.actionsDesc[i])
		if i == m.actionsCursor {
			lines = append(lines, m.styles.rowActive.Render(line))
		} else {
			lines = append(lines, m.styles.row.Render(line))
		}
	}
	return m.styles.panel.Copy().Width(panelWidth).Render(strings.Join(lines, "\n"))
}

func (m model) renderResult() string {
	title := m.styles.success.Render(m.resultTitle)
	panel := m.styles.panel
	if m.resultErr {
		title = m.styles.error.Render(m.resultTitle)
		panel = m.styles.panelError
	}
	return lipgloss.JoinVertical(lipgloss.Left, title, panel.Copy().Width(m.contentWidth()).Render(m.resultBody))
}

func (m model) renderConnection() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.styles.success.Render(m.resultTitle),
		m.styles.panel.Copy().Width(m.contentWidth()).Render(m.resultBody),
	)
}

func (m model) renderFooter() string {
	parts := []string{
		m.styles.hotkeyLabel.Render("F2") + m.styles.hotkeyValue.Render(" Add"),
		m.styles.hotkeyLabel.Render("F5") + m.styles.hotkeyValue.Render(" Refresh"),
		m.styles.hotkeyLabel.Render("F6") + m.styles.hotkeyValue.Render(" Actions"),
		m.styles.hotkeyLabel.Render("Esc") + m.styles.hotkeyValue.Render(" Back"),
		m.styles.hotkeyLabel.Render("F10") + m.styles.hotkeyValue.Render(" Quit"),
	}
	return m.styles.hotkeys.Copy().Width(m.contentWidth()).Render(strings.Join(parts, "  "))
}
