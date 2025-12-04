package commands

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	FetchCmd.Flags().StringVarP(&stableBranch, "branch", "b", "stable", "Stable branch name, default is 'stable'")
}

var FetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Merge stable branch into current feat branch, like merge stable into feat-3.4.0",
	Run: func(cmd *cobra.Command, args []string) {
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
		execCommand("git", "fetch", "--all")
		execCommand("git", "checkout", stableBranch)
		execCommand("git", "pull", "origin", stableBranch)
		execCommand("git", "checkout", version)
		execCommand("git", "pull", "origin", version)
		execCommand("git", "merge", "--no-ff", "-m",
			fmt.Sprintf("[Branch Merge] Merge %s into %s", stableBranch, version), stableBranch)
		execCommand("git", "push", "--set-upstream", "origin", version)
		successLog("Fetched updates and merged [%s] into [%s]", stableBranch, version)
	},
}
