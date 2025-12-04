package commands

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var RenameCmd = &cobra.Command{
	Use:   "rename feat-1.3.40",
	Short: "Rename current project directory",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("require exactly one argument")
		}
		if args[0] == "" {
			return fmt.Errorf("new project name cannot be empty")
		}
		var matched bool
		for _, p := range prefix {
			if strings.HasPrefix(args[0], p+"-") {
				matched = true
				break
			}
		}
		if !matched {
			eg := make([]string, 0, len(prefix))
			for _, p := range prefix {
				eg = append(eg, p+"-xxxx")
			}
			return fmt.Errorf("new project name must start with one of the prefixes: %s", strings.Join(eg, ", "))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		newVersion := args[0]
		pwd, err := os.Getwd()
		if err != nil {
			errLog("failed to get current working directory: %v", err)
		}
		dir := path.Base(pwd)
		if !regexpProject.MatchString(dir) {
			errLog("current directory is not a valid project directory")
		}
		matches := regexpProject.FindStringSubmatch(dir)
		if isDebug {
			log.Printf("matches: [%s]\n", strings.Join(matches, ", "))
		}
		version := matches[1]
		if newVersion == version {
			errLog("new project name is the same as the current one")
		}
		newDir := strings.ReplaceAll(dir, matches[1], newVersion)
		newPath := path.Join(path.Dir(pwd), newDir)
		if err := os.Rename(pwd, newPath); err != nil {
			errLog("failed to rename project directory: %v", err)
		}
		successLog("Renamed project directory to: %s", newPath)
		execCommand("git", "-C", newPath, "checkout", "-b", newVersion, version)
		execCommand("git", "-C", newPath, "push", "--set-upstream", "origin", newVersion)
		successLog("Created and pushed new branch: %s", newVersion)

		if config.OpenInIDEAfterUse {
			openByIDEA(newPath)
		}
	},
}
