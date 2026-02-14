package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	appconfig "vpn/internal/config"
	"vpn/internal/hysteria/app/add_user"
	"vpn/internal/hysteria/app/get_connection_url"
	"vpn/internal/hysteria/app/list_users"
	"vpn/internal/hysteria/app/remove_user"
	"vpn/internal/hysteria/app/rotate_password"
)

const (
	exitUsage = 2
	exitError = 1
)

func main() {
	loadResult, err := appconfig.LoadCLI()
	if err != nil {
		fatalf("load config: %v", err)
	}
	cfg := loadResult.Config
	if loadResult.WasCreated {
		fmt.Fprintf(os.Stderr, "Config created: %s\n", loadResult.Path)
	}

	addUserUseCase, err := add_user.BuildUseCase(cfg)
	if err != nil {
		fatalf("build add-user usecase: %v", err)
	}

	rotatePasswordUseCase, err := rotate_password.BuildUseCase(cfg)
	if err != nil {
		fatalf("build rotate-password usecase: %v", err)
	}

	removeUserUseCase, err := remove_user.BuildUseCase(cfg)
	if err != nil {
		fatalf("build remove-user usecase: %v", err)
	}

	listUsersUseCase, err := list_users.BuildUseCase(cfg)
	if err != nil {
		fatalf("build list-users usecase: %v", err)
	}

	connectionURLUseCase, err := get_connection_url.BuildUseCase(cfg)
	if err != nil {
		fatalf("build connection usecase: %v", err)
	}

	if err := run(os.Args[1:], addUserUseCase, rotatePasswordUseCase, removeUserUseCase, listUsersUseCase, connectionURLUseCase, cfg, os.Stdin, os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		var codeErr exitCodeError
		if errors.As(err, &codeErr) {
			os.Exit(codeErr.code)
		}
		fatalf("%v", err)
	}
}

func run(
	args []string,
	addUserUseCase *add_user.UseCase,
	rotatePasswordUseCase *rotate_password.UseCase,
	removeUserUseCase *remove_user.UseCase,
	listUsersUseCase *list_users.UseCase,
	connectionUseCase *get_connection_url.UseCase,
	cfg appconfig.Config,
	in io.Reader,
	out, errOut io.Writer,
) error {
	if len(args) == 0 {
		printRootHelp(errOut)
		return exitWithCode(exitUsage)
	}

	switch args[0] {
	case "help", "-h", "--help":
		printRootHelp(out)
		return nil
	case "init":
		return runInit(args[1:], cfg, out, errOut)
	case "add-user":
		return runAddUser(args[1:], addUserUseCase, cfg, in, out, errOut)
	case "rotate-password":
		return runRotatePassword(args[1:], rotatePasswordUseCase, cfg, in, out, errOut)
	case "remove-user":
		return runRemoveUser(args[1:], removeUserUseCase, cfg, in, out, errOut)
	case "list-users":
		return runListUsers(args[1:], listUsersUseCase, out, errOut)
	case "connection":
		return runConnection(args[1:], connectionUseCase, in, out, errOut)
	default:
		printRootHelp(errOut)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runAddUser(args []string, useCase *add_user.UseCase, cfg appconfig.Config, in io.Reader, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("add-user", flag.ContinueOnError)
	fs.SetOutput(errOut)

	username := fs.String("username", "", "username to add")
	yes := fs.Bool("yes", false, "skip confirmation")
	output := fs.String("output", "text", "output format: text|json")

	fs.Usage = func() {
		fmt.Fprintf(errOut, "Usage:\n")
		fmt.Fprintf(errOut, "  %s add-user [flags]\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Examples:\n")
		fmt.Fprintf(errOut, "  %s add-user --username alice\n", os.Args[0])
		fmt.Fprintf(errOut, "  %s add-user --username alice --output json --yes\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *output != "text" && *output != "json" {
		return fmt.Errorf("invalid --output %q (allowed: text|json)", *output)
	}

	reader := bufio.NewReader(in)
	interactive := isInteractiveInput()

	if *username == "" && interactive {
		value, err := promptRequired(reader, out, "Username")
		if err != nil {
			return err
		}
		*username = value
	}

	if *username == "" {
		fs.Usage()
		return exitWithCode(exitUsage)
	}

	if !*yes {
		if !confirm(reader, out, fmt.Sprintf("Add user %q into %s? [y/N]: ", *username, cfg.HysteriaConfigPath)) {
			return errors.New("operation canceled")
		}
	}

	password, err := useCase.Execute(context.Background(), *username)
	if err != nil {
		return fmt.Errorf("add user: %w", err)
	}

	if *output == "json" {
		return json.NewEncoder(out).Encode(map[string]any{
			"status":   "ok",
			"username": *username,
			"password": password,
			"config":   cfg.HysteriaConfigPath,
		})
	}

	fmt.Fprintf(out, "User %q added to %s\nPassword: %s\n", *username, cfg.HysteriaConfigPath, password)
	return nil
}

func runConnection(args []string, useCase *get_connection_url.UseCase, in io.Reader, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("connection", flag.ContinueOnError)
	fs.SetOutput(errOut)

	username := fs.String("username", "", "existing username")
	output := fs.String("output", "text", "output format: text|json")
	showQR := fs.Bool("qr", true, "print QR code in terminal")

	fs.Usage = func() {
		fmt.Fprintf(errOut, "Usage:\n")
		fmt.Fprintf(errOut, "  %s connection [flags]\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Examples:\n")
		fmt.Fprintf(errOut, "  %s connection --username valera\n", os.Args[0])
		fmt.Fprintf(errOut, "  %s connection --username valera --output json --qr=false\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("invalid --output %q (allowed: text|json)", *output)
	}

	reader := bufio.NewReader(in)
	if *username == "" && isInteractiveInput() {
		value, err := promptRequired(reader, out, "Username")
		if err != nil {
			return err
		}
		*username = value
	}
	if *username == "" {
		fs.Usage()
		return exitWithCode(exitUsage)
	}

	connectionURL, err := useCase.Execute(context.Background(), *username)
	if err != nil {
		return fmt.Errorf("build connection url: %w", err)
	}

	if *output == "json" {
		if err := json.NewEncoder(out).Encode(map[string]any{
			"status": "ok",
			"url":    connectionURL,
		}); err != nil {
			return err
		}
		return nil
	}

	fmt.Fprintln(out, connectionURL)
	if *showQR {
		fmt.Fprintln(out)
		printQRCode(out, connectionURL)
	}
	return nil
}

func runRotatePassword(args []string, useCase *rotate_password.UseCase, cfg appconfig.Config, in io.Reader, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("rotate-password", flag.ContinueOnError)
	fs.SetOutput(errOut)

	username := fs.String("username", "", "existing username")
	yes := fs.Bool("yes", false, "skip confirmation")
	output := fs.String("output", "text", "output format: text|json")

	fs.Usage = func() {
		fmt.Fprintf(errOut, "Usage:\n")
		fmt.Fprintf(errOut, "  %s rotate-password [flags]\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Examples:\n")
		fmt.Fprintf(errOut, "  %s rotate-password --username alice\n", os.Args[0])
		fmt.Fprintf(errOut, "  %s rotate-password --username alice --output json --yes\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("invalid --output %q (allowed: text|json)", *output)
	}

	reader := bufio.NewReader(in)
	interactive := isInteractiveInput()

	if *username == "" && interactive {
		value, err := promptRequired(reader, out, "Username")
		if err != nil {
			return err
		}
		*username = value
	}

	if *username == "" {
		fs.Usage()
		return exitWithCode(exitUsage)
	}

	if !*yes {
		if !confirm(reader, out, fmt.Sprintf("Rotate password for %q in %s? [y/N]: ", *username, cfg.HysteriaConfigPath)) {
			return errors.New("operation canceled")
		}
	}

	password, err := useCase.Execute(context.Background(), *username)
	if err != nil {
		return fmt.Errorf("rotate password: %w", err)
	}

	if *output == "json" {
		return json.NewEncoder(out).Encode(map[string]any{
			"status":   "ok",
			"username": *username,
			"password": password,
			"config":   cfg.HysteriaConfigPath,
		})
	}

	fmt.Fprintf(out, "Password rotated for %q in %s\nNew password: %s\n", *username, cfg.HysteriaConfigPath, password)
	return nil
}

func runRemoveUser(args []string, useCase *remove_user.UseCase, cfg appconfig.Config, in io.Reader, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("remove-user", flag.ContinueOnError)
	fs.SetOutput(errOut)

	username := fs.String("username", "", "existing username")
	yes := fs.Bool("yes", false, "skip confirmation")
	output := fs.String("output", "text", "output format: text|json")
	fs.Usage = func() {
		fmt.Fprintf(errOut, "Usage:\n")
		fmt.Fprintf(errOut, "  %s remove-user [flags]\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Examples:\n")
		fmt.Fprintf(errOut, "  %s remove-user --username alice\n", os.Args[0])
		fmt.Fprintf(errOut, "  %s remove-user --username alice --output json --yes\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Flags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("invalid --output %q (allowed: text|json)", *output)
	}

	reader := bufio.NewReader(in)
	if *username == "" && isInteractiveInput() {
		value, err := promptRequired(reader, out, "Username")
		if err != nil {
			return err
		}
		*username = value
	}
	if *username == "" {
		fs.Usage()
		return exitWithCode(exitUsage)
	}
	if !*yes {
		if !confirm(reader, out, fmt.Sprintf("Remove user %q from %s? [y/N]: ", *username, cfg.HysteriaConfigPath)) {
			return errors.New("operation canceled")
		}
	}
	if err := useCase.Execute(context.Background(), *username); err != nil {
		return fmt.Errorf("remove user: %w", err)
	}
	if *output == "json" {
		return json.NewEncoder(out).Encode(map[string]any{
			"status":   "ok",
			"username": *username,
			"config":   cfg.HysteriaConfigPath,
		})
	}
	fmt.Fprintf(out, "User %q removed from %s\n", *username, cfg.HysteriaConfigPath)
	return nil
}

func runListUsers(args []string, useCase *list_users.UseCase, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("list-users", flag.ContinueOnError)
	fs.SetOutput(errOut)
	output := fs.String("output", "text", "output format: text|json")
	fs.Usage = func() {
		fmt.Fprintf(errOut, "Usage:\n")
		fmt.Fprintf(errOut, "  %s list-users [flags]\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Flags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("invalid --output %q (allowed: text|json)", *output)
	}
	users, err := useCase.Execute(context.Background())
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}
	if *output == "json" {
		return json.NewEncoder(out).Encode(map[string]any{
			"status": "ok",
			"users":  users,
		})
	}
	for _, u := range users {
		fmt.Fprintln(out, u)
	}
	return nil
}

func runInit(args []string, cfg appconfig.Config, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(errOut)

	host := fs.String("host", "", "target server host or IP")
	inventory := fs.String("inventory", "", "ansible inventory path or host list")
	sshUser := fs.String("user", "root", "ssh username")
	sshKey := fs.String("ssh-key", "", "path to private ssh key")
	sshPort := fs.Int("port", 22, "ssh port")
	playbookPath := fs.String("playbook", "resources/ansible/hysteria_init.yml", "path to ansible playbook")
	varsPath := fs.String("vars", "resources/ansible/vars.yaml", "path to ansible vars yaml")
	configPath := fs.String("config", "resources/config.yaml", "path to local hysteria config.yaml")
	become := fs.Bool("become", true, "run tasks with privilege escalation")
	check := fs.Bool("check", false, "ansible check mode")

	fs.Usage = func() {
		fmt.Fprintf(errOut, "Usage:\n")
		fmt.Fprintf(errOut, "  %s init [flags]\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Examples:\n")
		fmt.Fprintf(errOut, "  %s init --host 1.2.3.4 --user root --ssh-key ~/.ssh/id_rsa\n", os.Args[0])
		fmt.Fprintf(errOut, "  %s init --inventory inventories/prod.ini --config resources/config.yaml\n\n", os.Args[0])
		fmt.Fprintf(errOut, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *inventory == "" && *host == "" {
		fs.Usage()
		return fmt.Errorf("either --host or --inventory is required")
	}

	ansiblePath, err := exec.LookPath("ansible-playbook")
	if err != nil {
		return errors.New("ansible-playbook not found in PATH")
	}

	absPlaybook, err := filepath.Abs(*playbookPath)
	if err != nil {
		return fmt.Errorf("resolve playbook path: %w", err)
	}
	absConfig, err := filepath.Abs(*configPath)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}
	absVars, err := filepath.Abs(*varsPath)
	if err != nil {
		return fmt.Errorf("resolve vars path: %w", err)
	}

	inv := *inventory
	if inv == "" {
		inv = *host + ","
	}

	cmdArgs := []string{
		"-i", inv,
		absPlaybook,
		"-u", *sshUser,
		"-e", "@" + absVars,
		"-e", fmt.Sprintf("ansible_port=%d", *sshPort),
		"-e", fmt.Sprintf("hysteria_service_name=%s", cfg.HysteriaServiceName),
		"-e", fmt.Sprintf("hysteria_config_src=%s", absConfig),
	}

	if *sshKey != "" {
		cmdArgs = append(cmdArgs, "--private-key", *sshKey)
	}
	if *become {
		cmdArgs = append(cmdArgs, "--become")
	}
	if *check {
		cmdArgs = append(cmdArgs, "--check")
	}

	fmt.Fprintf(out, "Running: %s %s\n", ansiblePath, strings.Join(cmdArgs, " "))
	cmd := exec.Command(ansiblePath, cmdArgs...)
	cmd.Stdout = out
	cmd.Stderr = errOut
	return cmd.Run()
}

func printQRCode(out io.Writer, content string) {
	path, err := exec.LookPath("qrencode")
	if err != nil {
		fmt.Fprintln(out, "qrencode not found. Install it to print terminal QR.")
		return
	}
	cmd := exec.Command(path, "-t", "ANSIUTF8", content)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(out, "failed to render qr: %v\n", err)
	}
}

func printRootHelp(w io.Writer) {
	fmt.Fprintf(w, "VPN CLI\n\n")
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "  %s <command> [flags]\n\n", os.Args[0])
	fmt.Fprintf(w, "Commands:\n")
	fmt.Fprintf(w, "  init         Run ansible playbook to bootstrap Hysteria on server\n")
	fmt.Fprintf(w, "  add-user     Add user to hysteria auth.userpass\n")
	fmt.Fprintf(w, "  remove-user  Remove existing user from hysteria auth.userpass\n")
	fmt.Fprintf(w, "  list-users   List users from hysteria auth.userpass\n")
	fmt.Fprintf(w, "  rotate-password Rotate password for existing user\n")
	fmt.Fprintf(w, "  connection   Print hy2 URL and QR code for a user\n")
	fmt.Fprintf(w, "  help         Show this help\n\n")
	fmt.Fprintf(w, "Use \"%s <command> --help\" for command flags.\n", os.Args[0])
}

func promptRequired(reader *bufio.Reader, out io.Writer, label string) (string, error) {
	for {
		fmt.Fprintf(out, "%s: ", label)
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		value := strings.TrimSpace(line)
		if value != "" {
			return value, nil
		}
		if errors.Is(err, io.EOF) {
			return "", errors.New("input ended")
		}
	}
}

func confirm(reader *bufio.Reader, out io.Writer, prompt string) bool {
	fmt.Fprint(out, prompt)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false
	}
	v := strings.TrimSpace(strings.ToLower(line))
	return v == "y" || v == "yes"
}

func isInteractiveInput() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(exitError)
}

type exitCodeError struct {
	code int
}

func (e exitCodeError) Error() string {
	return ""
}

func exitWithCode(code int) error {
	if code == 0 {
		return nil
	}
	return exitCodeError{code: code}
}
