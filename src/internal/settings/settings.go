package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"gopkg.in/yaml.v3"
)

const (
	appName      = "hashverifier"
	settingsFile = "settings.yaml"
)

type GenerateSettings struct {
	FollowSymbolicLinks bool     `yaml:"follow_symbolic_links"`
	SortPaths           bool     `yaml:"sort_paths"`
	Algorithm           string   `yaml:"algorithm"`
	ColumnOrder         []string `yaml:"column_order"`
	SortColumn          string   `yaml:"sort_column"`
	SortOrder           string   `yaml:"sort_order"`
}

type VerifySettings struct {
	VerifyOnOpen bool     `yaml:"verify_on_open"`
	ColumnOrder  []string `yaml:"column_order"`
	SortColumn   string   `yaml:"sort_column"`
	SortOrder    string   `yaml:"sort_order"`
}

type WindowSettings struct {
	TabOrder    []string `yaml:"tab_order"`
	CurrentPage int      `yaml:"current_page"`
}

type FlatpakSettings struct {
	SuppressSandboxWarning bool `yaml:"suppress_sandbox_warning"`
}

type Settings struct {
	Window   WindowSettings   `yaml:"window"`
	Generate GenerateSettings `yaml:"generate"`
	Verify   VerifySettings   `yaml:"verify"`
	Flatpak  FlatpakSettings  `yaml:"flatpak"`
}

func DefaultSettings() *Settings {
	return &Settings{
		Window: WindowSettings{
			TabOrder:    []string{"generate", "verify"},
			CurrentPage: 0,
		},
		Generate: GenerateSettings{
			FollowSymbolicLinks: true,
			SortPaths:           true,
			Algorithm:           ".md5",
			ColumnOrder:         []string{"idx", "path", "size", "hash", "note"},
			SortColumn:          "idx",
			SortOrder:           SortOrderAsc,
		},
		Verify: VerifySettings{
			VerifyOnOpen: true,
			ColumnOrder:  []string{"idx", "status", "path", "size", "hash", "expected_hash", "note"},
			SortColumn:   "status",
			SortOrder:    SortOrderDesc,
		},
		Flatpak: FlatpakSettings{
			SuppressSandboxWarning: false,
		},
	}
}

func getConfigDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable is not set")
		}

		return filepath.Join(appData, appName), nil

	case "linux":
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			return filepath.Join(xdgConfig, appName), nil
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}

		return filepath.Join(home, ".config", appName), nil

	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}

		return filepath.Join(home, "Library", "Application Support", appName), nil

	default:
		return os.Getwd()
	}
}

func getSettingsPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, settingsFile), nil
}

func ensureConfigDir() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	return nil
}

func Load() (*Settings, error) {
	settingsPath, err := getSettingsPath()
	if err != nil {
		return DefaultSettings(), err
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSettings(), nil
		}

		return DefaultSettings(), fmt.Errorf("failed to read settings file: %w", err)
	}

	settings := DefaultSettings()
	if err := yaml.Unmarshal(data, settings); err != nil {
		return DefaultSettings(), fmt.Errorf("failed to parse settings file: %w", err)
	}

	settings.fixColumnOrder()

	return settings, nil
}

func (s *Settings) fixColumnOrder() {
	defaultSettings := DefaultSettings()

	for _, col := range defaultSettings.Generate.ColumnOrder {
		if !slices.Contains(s.Generate.ColumnOrder, col) {
			s.Generate.ColumnOrder = defaultSettings.Generate.ColumnOrder
			break
		}
	}

	for _, col := range defaultSettings.Verify.ColumnOrder {
		if !slices.Contains(s.Verify.ColumnOrder, col) {
			s.Verify.ColumnOrder = defaultSettings.Verify.ColumnOrder
			break
		}
	}
}

func (s *Settings) Save() error {
	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("failed to ensure config directory: %w", err)
	}

	settingsPath, err := getSettingsPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

func GetSettingsPath() (string, error) {
	return getSettingsPath()
}

func Reset() error {
	defaultSettings := DefaultSettings()
	return defaultSettings.Save()
}
