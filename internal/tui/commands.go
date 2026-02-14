package tui

import (
	"context"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"vpn/internal/hysteria/app/add_user"
	"vpn/internal/hysteria/app/get_connection_url"
	"vpn/internal/hysteria/app/get_user_stats"
	"vpn/internal/hysteria/app/list_users"
	"vpn/internal/hysteria/app/remove_user"
	"vpn/internal/hysteria/app/rotate_password"
)

func loadUsersCmd(listUC *list_users.UseCase, statsUC *get_user_stats.UseCase) tea.Cmd {
	return func() tea.Msg {
		users, err := listUC.Execute(context.Background())
		if err != nil {
			return usersLoadedMsg{users: nil, stats: nil, err: err}
		}
		stats := map[string]get_user_stats.UserStats{}
		if statsUC != nil {
			stats, _ = statsUC.Execute(context.Background(), users)
		}
		return usersLoadedMsg{users: users, stats: stats, err: nil}
	}
}

func addUserCmd(uc *add_user.UseCase, username string) tea.Cmd {
	return func() tea.Msg {
		password, err := uc.Execute(context.Background(), username)
		if err != nil {
			return operationMsg{err: err}
		}
		return operationMsg{title: "User created", body: fmt.Sprintf("User: %s\nPassword: %s", username, password), refresh: true}
	}
}

func rotatePasswordCmd(uc *rotate_password.UseCase, username string) tea.Cmd {
	return func() tea.Msg {
		password, err := uc.Execute(context.Background(), username)
		if err != nil {
			return operationMsg{err: err}
		}
		return operationMsg{title: "Password rotated", body: fmt.Sprintf("User: %s\nNew password: %s", username, password), refresh: true}
	}
}

func removeUserCmd(uc *remove_user.UseCase, username string) tea.Cmd {
	return func() tea.Msg {
		if err := uc.Execute(context.Background(), username); err != nil {
			return operationMsg{err: err}
		}
		return operationMsg{title: "User removed", body: fmt.Sprintf("User %s removed", username), refresh: true}
	}
}

func connectionCmd(uc *get_connection_url.UseCase, username string) tea.Cmd {
	return func() tea.Msg {
		url, err := uc.Execute(context.Background(), username)
		if err != nil {
			return operationMsg{connection: true, err: err}
		}
		return operationMsg{title: "Connection", body: url + "\n\n" + renderQRCode(url), connection: true}
	}
}

func renderQRCode(content string) string {
	path, err := exec.LookPath("qrencode")
	if err != nil {
		return "qrencode not found. Install it to render QR."
	}
	cmd := exec.Command(path, "-t", "ANSIUTF8", content)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "failed to render qr: " + err.Error()
	}
	return string(out)
}
