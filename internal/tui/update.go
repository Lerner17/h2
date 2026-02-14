package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Init() tea.Cmd { return loadUsersCmd(m.listUC, m.statsUC) }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.Width = max(20, min(60, msg.Width-18))
		return m, nil
	case usersLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.resultTitle = "Load users failed"
			m.resultBody = msg.err.Error()
			m.resultErr = true
			m.state = stateResult
			return m, nil
		}
		m.users = msg.users
		m.userStats = msg.stats
		if m.usersCursor >= len(m.users) {
			m.usersCursor = max(0, len(m.users)-1)
		}
		return m, nil
	case operationMsg:
		if msg.connection {
			if msg.err != nil {
				m.resultTitle = "Connection failed"
				m.resultBody = msg.err.Error()
				m.resultErr = true
				m.state = stateResult
				return m, nil
			}
			m.resultTitle = msg.title
			m.resultBody = msg.body
			m.resultErr = false
			m.state = stateConnection
			return m, nil
		}

		if msg.err != nil {
			m.resultTitle = "Operation failed"
			m.resultBody = msg.err.Error()
			m.resultErr = true
		} else {
			m.resultTitle = msg.title
			m.resultBody = msg.body
			m.resultErr = false
		}
		m.state = stateResult
		if msg.refresh {
			m.loading = true
			return m, loadUsersCmd(m.listUC, m.statsUC)
		}
		return m, nil
	case tea.KeyMsg:
		switch m.state {
		case stateUsers:
			return m.updateUsers(msg)
		case stateAddInput:
			return m.updateAddInput(msg)
		case stateUserActions:
			return m.updateUserActions(msg)
		case stateResult, stateConnection:
			if msg.String() == "q" || msg.String() == "ctrl+c" || msg.String() == "f10" {
				return m, tea.Quit
			}
			m.state = stateUsers
			return m, nil
		}
	}
	return m, nil
}

func (m model) updateUsers(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "f10":
		return m, tea.Quit
	case "a", "f2":
		m.input.SetValue("")
		m.state = stateAddInput
		return m, nil
	case "r", "f5":
		m.loading = true
		return m, loadUsersCmd(m.listUC, m.statsUC)
	case "up", "k":
		if m.usersCursor > 0 {
			m.usersCursor--
		}
	case "down", "j":
		if m.usersCursor < len(m.users)-1 {
			m.usersCursor++
		}
	case "enter", "f6":
		if len(m.users) == 0 {
			return m, nil
		}
		m.selectedUser = m.users[m.usersCursor]
		m.actionsCursor = 0
		m.state = stateUserActions
	}
	return m, nil
}

func (m model) updateAddInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateUsers
		return m, nil
	case "enter":
		username := strings.TrimSpace(m.input.Value())
		if username == "" {
			m.resultTitle = "Validation"
			m.resultBody = "username is required"
			m.resultErr = true
			m.state = stateResult
			return m, nil
		}
		return m, addUserCmd(m.addUC, username)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateUserActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateUsers
		return m, nil
	case "up", "k":
		if m.actionsCursor > 0 {
			m.actionsCursor--
		}
	case "down", "j":
		if m.actionsCursor < len(m.actions)-1 {
			m.actionsCursor++
		}
	case "enter":
		switch userAction(m.actionsCursor) {
		case actRotate:
			return m, rotatePasswordCmd(m.rotateUC, m.selectedUser)
		case actRemove:
			return m, removeUserCmd(m.removeUC, m.selectedUser)
		case actConnection:
			return m, connectionCmd(m.connectionUC, m.selectedUser)
		case actBack:
			m.state = stateUsers
			return m, nil
		}
	}
	return m, nil
}
