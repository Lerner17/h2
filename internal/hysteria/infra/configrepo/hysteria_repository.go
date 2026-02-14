package configrepo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"vpn/internal/hysteria/domain"
)

type Repository struct {
	path string
	mu   sync.Mutex
}

func NewRepository(path string) *Repository {
	return &Repository{path: path}
}

func (r *Repository) AddUser(_ context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, root, err := r.readRoot()
	if err != nil {
		return err
	}

	auth := ensureMappingValue(root, "auth")

	authType := findMappingValue(auth, "type")
	if authType == nil || authType.Value != "userpass" {
		return errors.New("auth.type must be userpass")
	}

	userPass := ensureMappingValue(auth, "userpass")

	if userPass.Kind != yaml.MappingNode {
		return errors.New("auth.userpass must be a map")
	}

	if findMappingValue(userPass, user.Username) != nil {
		return domain.ErrUserAlreadyExists
	}

	userPass.Content = append(userPass.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: user.Username, Tag: "!!str"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: user.Password, Tag: "!!str"},
	)

	result, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(r.path, result, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func (r *Repository) RotatePassword(_ context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, root, err := r.readRoot()
	if err != nil {
		return err
	}

	auth := findMappingValue(root, "auth")
	if auth == nil {
		return errors.New("auth section not found")
	}
	authType := findMappingValue(auth, "type")
	if authType == nil || authType.Value != "userpass" {
		return errors.New("auth.type must be userpass")
	}

	userPass := findMappingValue(auth, "userpass")
	if userPass == nil || userPass.Kind != yaml.MappingNode {
		return errors.New("auth.userpass must be a map")
	}

	passwordNode := findMappingValue(userPass, user.Username)
	if passwordNode == nil {
		return domain.ErrUserNotFound
	}
	passwordNode.Value = user.Password
	passwordNode.Tag = "!!str"

	result, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(r.path, result, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func (r *Repository) RemoveUser(_ context.Context, username string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, root, err := r.readRoot()
	if err != nil {
		return err
	}

	auth := findMappingValue(root, "auth")
	if auth == nil {
		return errors.New("auth section not found")
	}
	userPass := findMappingValue(auth, "userpass")
	if userPass == nil || userPass.Kind != yaml.MappingNode {
		return errors.New("auth.userpass must be a map")
	}

	if !deleteMappingKey(userPass, username) {
		return domain.ErrUserNotFound
	}

	result, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(r.path, result, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (r *Repository) ListUsers(_ context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, root, err := r.readRoot()
	if err != nil {
		return nil, err
	}

	auth := findMappingValue(root, "auth")
	if auth == nil {
		return nil, errors.New("auth section not found")
	}
	userPass := findMappingValue(auth, "userpass")
	if userPass == nil || userPass.Kind != yaml.MappingNode {
		return nil, errors.New("auth.userpass must be a map")
	}

	users := make([]string, 0, len(userPass.Content)/2)
	for i := 0; i < len(userPass.Content)-1; i += 2 {
		users = append(users, userPass.Content[i].Value)
	}
	sort.Strings(users)
	return users, nil
}

func (r *Repository) GetConnectionConfig(_ context.Context, username string) (domain.ConnectionConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, root, err := r.readRoot()
	if err != nil {
		return domain.ConnectionConfig{}, err
	}

	auth := findMappingValue(root, "auth")
	if auth == nil {
		return domain.ConnectionConfig{}, errors.New("auth section not found")
	}
	userPass := findMappingValue(auth, "userpass")
	if userPass == nil || userPass.Kind != yaml.MappingNode {
		return domain.ConnectionConfig{}, errors.New("auth.userpass must be a map")
	}

	passwordNode := findMappingValue(userPass, username)
	if passwordNode == nil {
		return domain.ConnectionConfig{}, domain.ErrUserNotFound
	}

	host, err := readHost(root)
	if err != nil {
		return domain.ConnectionConfig{}, err
	}

	port, err := readPort(root)
	if err != nil {
		return domain.ConnectionConfig{}, err
	}

	obfsType, obfsPassword := readObfs(root)
	sni := host

	return domain.ConnectionConfig{
		Username:     username,
		Password:     passwordNode.Value,
		Host:         host,
		Port:         port,
		SNI:          sni,
		ObfsType:     obfsType,
		ObfsPassword: obfsPassword,
	}, nil
}

func (r *Repository) readRoot() (*yaml.Node, *yaml.Node, error) {
	raw, err := os.ReadFile(r.path)
	if err != nil {
		return nil, nil, fmt.Errorf("read config: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, nil, errors.New("invalid hysteria config format")
	}

	return &doc, doc.Content[0], nil
}

func findMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}

	return nil
}

func ensureMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	value := findMappingValue(mapping, key)
	if value != nil {
		return value
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}
	valueNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	mapping.Content = append(mapping.Content, keyNode, valueNode)
	return valueNode
}

func deleteMappingKey(mapping *yaml.Node, key string) bool {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
			return true
		}
	}
	return false
}

func readHost(root *yaml.Node) (string, error) {
	acme := findMappingValue(root, "acme")
	if acme == nil {
		return "", errors.New("acme section not found")
	}
	domains := findMappingValue(acme, "domains")
	if domains == nil || domains.Kind != yaml.SequenceNode || len(domains.Content) == 0 {
		return "", errors.New("acme.domains[0] is required")
	}
	host := strings.TrimSpace(domains.Content[0].Value)
	if host == "" {
		return "", errors.New("acme.domains[0] is empty")
	}
	return host, nil
}

func readPort(root *yaml.Node) (int, error) {
	listen := findMappingValue(root, "listen")
	if listen == nil || strings.TrimSpace(listen.Value) == "" {
		return 443, nil
	}

	value := strings.TrimSpace(listen.Value)
	if strings.HasPrefix(value, ":") {
		p, err := strconv.Atoi(strings.TrimPrefix(value, ":"))
		if err != nil {
			return 0, fmt.Errorf("invalid listen port: %q", value)
		}
		return p, nil
	}

	_, portText, err := net.SplitHostPort(value)
	if err != nil {
		if p, convErr := strconv.Atoi(value); convErr == nil {
			return p, nil
		}
		return 0, fmt.Errorf("invalid listen value: %q", value)
	}

	p, err := strconv.Atoi(portText)
	if err != nil {
		return 0, fmt.Errorf("invalid listen port: %q", value)
	}
	return p, nil
}

func readObfs(root *yaml.Node) (string, string) {
	obfs := findMappingValue(root, "obfs")
	if obfs == nil {
		return "", ""
	}

	obfsTypeNode := findMappingValue(obfs, "type")
	if obfsTypeNode == nil || strings.TrimSpace(obfsTypeNode.Value) == "" {
		return "", ""
	}
	obfsType := strings.TrimSpace(obfsTypeNode.Value)

	obfsValueNode := findMappingValue(obfs, obfsType)
	if obfsValueNode == nil {
		return obfsType, ""
	}

	if obfsValueNode.Kind == yaml.ScalarNode {
		return obfsType, strings.TrimSpace(obfsValueNode.Value)
	}
	if obfsValueNode.Kind == yaml.MappingNode {
		passwordNode := findMappingValue(obfsValueNode, "password")
		if passwordNode != nil {
			return obfsType, strings.TrimSpace(passwordNode.Value)
		}
	}

	return obfsType, ""
}
