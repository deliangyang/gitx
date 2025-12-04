package commands

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Merge feat branch into target branch, like merge feat-3.4.0 into new-dev",
	Run: func(cmd *cobra.Command, args []string) {
		pwd, err := os.Getwd()
		if err != nil {
			errLog("failed to get current working directory: %v", err)

		}
		dir := path.Base(pwd)
		if !regexpProject.MatchString(dir) {
			errLog("current directory is not a valid project directory")
			os.Exit(1)
		}
		matches := regexpProject.FindStringSubmatch(dir)
		if isDebug {
			log.Printf("matches: [%s]\n", strings.Join(matches, ", "))
		}
		version := matches[1]
		branch := matches[3]
		execCommand("git", "fetch", "--all")
		if !branchExists(pwd, branch) {
			errLog("branch [%s] does not exist", branch)
			os.Exit(1)
		}
		execCommand("git", "checkout", version)
		execCommand("git", "pull", "origin", version)
		execCommand("git", "checkout", branch)
		execCommand("git", "pull", "origin", branch)
		// merge feat branch into target branch
		execCommand("git", "merge", "--no-ff", "-m",
			fmt.Sprintf("[Branch Merge] Merge %s into %s", version, branch), version)
		execCommand("git", "push", "--set-upstream", "origin", branch)
		execCommand("git", "checkout", version)
		successLog("Synced branch [%s] with feat branch [%s]", branch, version)

		pipelineURL := fmt.Sprintf("%s/-/pipelines", getRepoURL())
		successLog("You can check the pipeline status at: [%s]", pipelineURL)
	},
}

func getRepoURL() string {
	data := execCommandWithOutput("git", "remote", "-v")
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		split := regexpSplitSpace.Split(line, -1)
		if len(split) >= 2 && split[0] == "origin" {
			if regexpGitRepo.MatchString(split[1]) {
				return "https://" + strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(split[1], "git@"), ".git"), ":", "/")
			} else {
				parsedURL, err := url.Parse(split[1])
				if err != nil {
					errLog("Invalid repository URL: [%s], error: %s", split[1], err.Error())
				}
				return strings.TrimSuffix(parsedURL.String(), ".git")
			}
		}
	}
	errLog("Failed to get repository URL from git remote")
	return ""
}
