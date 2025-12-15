package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:       "config",
	Short:     "Configure gitx settings",
	ValidArgs: []string{"", "view", "set"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		switch args[0] {
		case "set":
			if len(args) != 3 {
				return fmt.Errorf("set requires exactly two arguments: <key> <value>")
			}

			if !validKeys[args[1]] {
				return fmt.Errorf("invalid config key: %s", args[1])
			}
		case "view":
			if len(args) != 1 {
				return fmt.Errorf("view does not take any arguments")
			}
		default:
			return fmt.Errorf("invalid argument: %s", args[0])
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		configJSON := strings.TrimSpace(configJSON)
		storePath := getConfigFilePath()
		dirname := path.Dir(storePath)
		_, err := os.Stat(dirname)
		isDirExist := err == nil
		if err != nil {
			if !os.IsNotExist(err) {
				errLog("failed to stat config directory: %v", err)
			}
		}
		if len(args) == 0 {
			if !isDirExist {
				if err := os.MkdirAll(dirname, 0755); err != nil {
					errLog("failed to create config directory: %v", err)
				}
			}
			_, err := os.Stat(storePath)
			if err == nil {
				successLog("config directory %s already exists", storePath)
				return
			} else if !os.IsNotExist(err) {
				errLog("failed to stat config file: %v", err)
			}
			if err := os.WriteFile(storePath, []byte(configJSON), 0644); err != nil {
				errLog("failed to write config file: %v", err)
			}
			fmt.Println(configJSON)
			successLog("Configuration written to %s", storePath)
			return
		}
		switch args[0] {
		case "view":
			if !isDirExist {
				warningLog("config file does not exist, please run 'gitx config' to create one")
				return
			}
			data, err := os.ReadFile(storePath)
			if err != nil {
				errLog("failed to read config file: %v", err)
			}
			fmt.Println(string(data))
		case "set":
			key := args[1]
			value := args[2]
			var cfg Config
			if !isDirExist {
				data, err := os.ReadFile(storePath)
				if err != nil {
					errLog("failed to read config file: %v", err)
				}
				if err := json.Unmarshal(data, &cfg); err != nil {
					errLog("failed to parse config file: %v", err)
				}
			}
			switch key {
			case "workspace_dir":
				cfg.WorkspaceDir = value
			case "default_ide":
				cfg.DefaultIDE = value
			case "open_in_ide_after_use":
				if strings.ToLower(value) == "true" {
					cfg.OpenInIDEAfterUse = true
				} else if strings.ToLower(value) == "false" {
					cfg.OpenInIDEAfterUse = false
				} else {
					errLog("invalid value for open_in_ide_after_use: %s", value)
				}
			case "common_projects":
				projects := strings.Split(value, ",")
				for i := range projects {
					projects[i] = strings.TrimSpace(projects[i])
				}
				cfg.CommonProjects = projects
			case "prefix":
				prefixes := strings.Split(value, ",")
				for i := range prefixes {
					prefixes[i] = strings.TrimSpace(prefixes[i])
				}
				cfg.Prefix = prefixes
			}
			if cfg.Prefix == nil {
				cfg.Prefix = []string{}
			}
			if cfg.CommonProjects == nil {
				cfg.CommonProjects = []string{}
			}
			updatedData, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				errLog("failed to serialize config: %v", err)
			}
			if err := os.WriteFile(storePath, updatedData, 0644); err != nil {
				errLog("failed to write config file: %v", err)
			}
			successLog("Configuration updated: %s set to %s", key, value)
		}
	},
}

var (
	config       Config
	workspaceDir string
	validKeys    = map[string]bool{
		"workspace_dir":         true,
		"default_ide":           true,
		"open_in_ide_after_use": true,
		"common_projects":       true,
		"prefix":                true,
	}
	prefix = []string{
		"feat",
		"online-fix",
		"online-revision",
	}
	regexpProject *regexp.Regexp
)

func init() {
	loadConfig()

	workspaceDir = getHomeDir()
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		errLog("Make work space dir fail")
	}
	if len(config.Prefix) > 0 {
		prefix = config.Prefix
	}

	regexpProject = regexp.MustCompile(`((` + strings.Join(prefix, "|") + `)-[^-]+)-(.+)$`)

	if config.DefaultIDE == "" {
		config.DefaultIDE = "code"
	}

}

func loadConfig() {
	configFilePath := getConfigFilePath()
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		warningLog("config file does not exist, please run 'gitx config' to create one")
		return
	}
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		errLog("failed to read config file: %v", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		errLog("failed to parse config file: %v", err)
	}
}

var (
	configJSON = `
{
	  "workspace_dir": "~/gitx_workspace",
	  "default_ide": "code",
	  "open_in_ide_after_use": true,
	  "common_projects": [],
	  "prefix": ["feat", "fix", "hotfix", "online", "release"]
}
`
)

type Config struct {
	WorkspaceDir      string   `json:"workspace_dir"`
	DefaultIDE        string   `json:"default_ide"`
	OpenInIDEAfterUse bool     `json:"open_in_ide_after_use"`
	CommonProjects    []string `json:"common_projects"`
	Prefix            []string `json:"prefix"`
}

func getConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		errLog("failed to get user home directory: %v", err)
	}
	return fmt.Sprintf("%s/.gitx/config.json", homeDir)
}
