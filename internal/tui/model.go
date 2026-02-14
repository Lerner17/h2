package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"vpn/internal/hysteria/app/add_user"
	"vpn/internal/hysteria/app/get_connection_url"
	"vpn/internal/hysteria/app/list_users"
	"vpn/internal/hysteria/app/remove_user"
	"vpn/internal/hysteria/app/rotate_password"
)

type appState int

type userAction int

const (
	stateUsers appState = iota
	stateAddInput
	stateUserActions
	stateResult
	stateConnection
)

const (
	actRotate userAction = iota
	actRemove
	actConnection
	actBack
)

type usersLoadedMsg struct {
	users []string
	err   error
}

type operationMsg struct {
	title      string
	body       string
	err        error
	refresh    bool
	connection bool
}

type styles struct {
	header      lipgloss.Style
	headerDim   lipgloss.Style
	meterOn     lipgloss.Style
	meterOff    lipgloss.Style
	tableHead   lipgloss.Style
	row         lipgloss.Style
	rowActive   lipgloss.Style
	panel       lipgloss.Style
	panelError  lipgloss.Style
	success     lipgloss.Style
	error       lipgloss.Style
	muted       lipgloss.Style
	hotkeys     lipgloss.Style
	hotkeyLabel lipgloss.Style
	hotkeyValue lipgloss.Style
}

type model struct {
	state appState

	addUC        *add_user.UseCase
	rotateUC     *rotate_password.UseCase
	removeUC     *remove_user.UseCase
	listUC       *list_users.UseCase
	connectionUC *get_connection_url.UseCase

	users       []string
	usersCursor int
	loading     bool

	selectedUser  string
	actions       []string
	actionsDesc   []string
	actionsCursor int

	input textinput.Model

	resultTitle string
	resultBody  string
	resultErr   bool

	width  int
	height int

	styles styles
}

func newModel(deps *Dependencies) model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 42
	ti.Prompt = ""

	return model{
		state:        stateUsers,
		addUC:        deps.AddUser,
		rotateUC:     deps.RotatePassword,
		removeUC:     deps.RemoveUser,
		listUC:       deps.ListUsers,
		connectionUC: deps.Connection,
		loading:      true,
		actions: []string{
			"Rotate password",
			"Remove user",
			"Show connection URL + QR",
			"Back",
		},
		actionsDesc: []string{
			"Generate and apply a new password",
			"Delete user from auth.userpass",
			"Open dedicated connection view",
			"Return to users list",
		},
		input:  ti,
		styles: newStyles(),
	}
}
