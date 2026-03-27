package settings

import (
	"fmt"
	"strings"
)

type SettingInfo struct {
	Name        string
	Value       string
	Default     string
	Description string
}

type SettingsSection struct {
	Name     string
	Settings []SettingInfo
}

var descriptionsMap = map[string]string{
	"generate.follow_symbolic_links":   "Follow symbolic links when scanning directories",
	"generate.sort_paths":              "Sort paths before hashing",
	"generate.algorithm":               "Default hash algorithm (e.g., .sha256, .md5)",
	"generate.column_order":            "Order of columns in Generate tab",
	"generate.sort_column":             "Column to sort by in Generate tab",
	"generate.sort_order":              "Sort order in Generate tab (asc/desc)",
	"verify.verify_on_open":            "Auto-start verification when opening checksum file",
	"verify.column_order":              "Order of columns in Verify tab",
	"verify.sort_column":               "Column to sort by in Verify tab",
	"verify.sort_order":                "Sort order in Verify tab (asc/desc)",
	"window.tab_order":                 "Order of tabs in main window",
	"window.current_page":              "Currently active tab",
	"window.show_details":              "Show/hide details tables (Generate and Verify tabs)",
	"flatpak.suppress_sandbox_warning": "Suppress the Flatpak sandbox warning dialog on startup (Flatpak only)",
}

func GetAllSettingsInfo(cfg, defaults *Settings) []SettingsSection {
	return []SettingsSection{
		{
			Name: "Window",
			Settings: []SettingInfo{
				{
					Name:        "tab_order",
					Value:       formatSettingValueSlice(cfg.Window.TabOrder),
					Default:     formatSettingValueSlice(defaults.Window.TabOrder),
					Description: descriptionsMap["window.tab_order"],
				},
				{
					Name:        "current_page",
					Value:       formatSettingValue(cfg.Window.CurrentPage),
					Default:     formatSettingValue(defaults.Window.CurrentPage),
					Description: descriptionsMap["window.current_page"],
				},
				{
					Name:        "show_details",
					Value:       formatSettingValue(cfg.Window.ShowDetails),
					Default:     formatSettingValue(defaults.Window.ShowDetails),
					Description: descriptionsMap["window.show_details"],
				},
			},
		},
		{
			Name: "Generate",
			Settings: []SettingInfo{
				{
					Name:        "follow_symbolic_links",
					Value:       formatSettingValue(cfg.Generate.FollowSymbolicLinks),
					Default:     formatSettingValue(defaults.Generate.FollowSymbolicLinks),
					Description: descriptionsMap["generate.follow_symbolic_links"],
				},
				{
					Name:        "sort_paths",
					Value:       formatSettingValue(cfg.Generate.SortPaths),
					Default:     formatSettingValue(defaults.Generate.SortPaths),
					Description: descriptionsMap["generate.sort_paths"],
				},
				{
					Name:        "algorithm",
					Value:       formatSettingValue(cfg.Generate.Algorithm),
					Default:     formatSettingValue(defaults.Generate.Algorithm),
					Description: descriptionsMap["generate.algorithm"],
				},
				{
					Name:        "column_order",
					Value:       formatSettingValueSlice(cfg.Generate.ColumnOrder),
					Default:     formatSettingValueSlice(defaults.Generate.ColumnOrder),
					Description: descriptionsMap["generate.column_order"],
				},
				{
					Name:        "sort_column",
					Value:       formatSettingValue(cfg.Generate.SortColumn),
					Default:     formatSettingValue(defaults.Generate.SortColumn),
					Description: descriptionsMap["generate.sort_column"],
				},
				{
					Name:        "sort_order",
					Value:       formatSettingValue(cfg.Generate.SortOrder),
					Default:     formatSettingValue(defaults.Generate.SortOrder),
					Description: descriptionsMap["generate.sort_order"],
				},
			},
		},
		{
			Name: "Verify",
			Settings: []SettingInfo{
				{
					Name:        "verify_on_open",
					Value:       formatSettingValue(cfg.Verify.VerifyOnOpen),
					Default:     formatSettingValue(defaults.Verify.VerifyOnOpen),
					Description: descriptionsMap["verify.verify_on_open"],
				},
				{
					Name:        "column_order",
					Value:       formatSettingValueSlice(cfg.Verify.ColumnOrder),
					Default:     formatSettingValueSlice(defaults.Verify.ColumnOrder),
					Description: descriptionsMap["verify.column_order"],
				},
				{
					Name:        "sort_column",
					Value:       formatSettingValue(cfg.Verify.SortColumn),
					Default:     formatSettingValue(defaults.Verify.SortColumn),
					Description: descriptionsMap["verify.sort_column"],
				},
				{
					Name:        "sort_order",
					Value:       formatSettingValue(cfg.Verify.SortOrder),
					Default:     formatSettingValue(defaults.Verify.SortOrder),
					Description: descriptionsMap["verify.sort_order"],
				},
			},
		},
		{
			Name: "Flatpak",
			Settings: []SettingInfo{
				{
					Name:        "suppress_sandbox_warning",
					Value:       formatSettingValue(cfg.Flatpak.SuppressSandboxWarning),
					Default:     formatSettingValue(defaults.Flatpak.SuppressSandboxWarning),
					Description: descriptionsMap["flatpak.suppress_sandbox_warning"],
				},
			},
		},
	}
}

func formatSettingValue(v any) string {
	switch val := v.(type) {
	case bool:
		return fmt.Sprintf("%t", val)
	case string:
		if val == "" {
			return "-"
		}

		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatSettingValueSlice(slice []string) string {
	if len(slice) == 0 {
		return "-"
	}

	return strings.Join(slice, ", ")
}
