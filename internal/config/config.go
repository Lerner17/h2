package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	envConfigPath = "VPN_CONFIG_PATH"
)

type Config struct {
	HysteriaConfigPath     string `yaml:"hysteria_config_path"`
	HysteriaServiceName    string `yaml:"hysteria_service_name"`
	HysteriaRestartEnabled bool   `yaml:"hysteria_restart_enabled"`
	HysteriaRestartCommand string `yaml:"hysteria_restart_command"`
}

type CLILoadResult struct {
	Config     Config
	Path       string
	WasCreated bool
}

func Load() (Config, error) {
	cfg := defaultConfig()
	if err := applyEnvOverrides(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func LoadCLI() (CLILoadResult, error) {
	path, err := resolveCLIConfigPath()
	if err != nil {
		return CLILoadResult{}, err
	}

	created, err := ensureDefaultConfigFile(path)
	if err != nil {
		return CLILoadResult{}, err
	}

	cfg, err := readConfigFile(path)
	if err != nil {
		return CLILoadResult{}, err
	}
	if err := applyEnvOverrides(&cfg); err != nil {
		return CLILoadResult{}, err
	}

	return CLILoadResult{
		Config:     cfg,
		Path:       path,
		WasCreated: created,
	}, nil
}

func defaultConfig() Config {
	return Config{
		HysteriaConfigPath:     "/etc/hysteria/config.yaml",
		HysteriaServiceName:    "hysteria-server",
		HysteriaRestartEnabled: true,
		HysteriaRestartCommand: "",
	}
}

func readConfigFile(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", err)
	}

	return cfg, nil
}

func ensureDefaultConfigFile(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("stat config file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("create config dir: %w", err)
	}

	raw, err := yaml.Marshal(defaultConfig())
	if err != nil {
		return false, fmt.Errorf("marshal default config: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return false, fmt.Errorf("write default config: %w", err)
	}

	return true, nil
}

func resolveCLIConfigPath() (string, error) {
	if path := os.Getenv(envConfigPath); path != "" {
		return path, nil
	}

	const systemPath = "/etc/vpn/config.yaml"
	if _, err := os.Stat(systemPath); err == nil {
		return systemPath, nil
	}
	if canCreateUnder("/etc/vpn") {
		return systemPath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "vpn", "config.yaml"), nil
}

func canCreateUnder(dir string) bool {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	return true
}

func applyEnvOverrides(cfg *Config) error {
	if v, ok := os.LookupEnv("HYSTERIA_CONFIG_PATH"); ok {
		cfg.HysteriaConfigPath = v
	}
	if v, ok := os.LookupEnv("HYSTERIA_SERVICE_NAME"); ok {
		cfg.HysteriaServiceName = v
	}
	if v, ok := os.LookupEnv("HYSTERIA_RESTART_COMMAND"); ok {
		cfg.HysteriaRestartCommand = v
	}
	if v, ok := os.LookupEnv("HYSTERIA_RESTART_ENABLED"); ok {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("parse HYSTERIA_RESTART_ENABLED: %w", err)
		}
		cfg.HysteriaRestartEnabled = parsed
	}
	return nil
}
