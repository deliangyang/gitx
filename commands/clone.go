package commands

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var (
	mainBranch string
)

func init() {
	CloneCmd.Flags().StringVarP(&mainBranch, "branch", "b", "main", "Main branch name, default is 'main'")
	CloneCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", workspaceDir, "Workspace directory")

}

var CloneCmd = &cobra.Command{
	Use:   "clone git@github.com:deliangyang/gitx.git feat-3.4.0 new-dev -b main",
	Short: "Clone a repository",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("require exactly three argument")
		}
		if !regexpGitRepo.MatchString(args[0]) {
			if _, err := url.Parse(args[0]); err != nil {
				return fmt.Errorf("invalid repository URL: [%s], error: %s", args[0], err.Error())
			}
		}
		if args[1] == "" {
			eg := make([]string, 0, len(prefix))
			for _, p := range prefix {
				eg = append(eg, p+"-xxxx")
			}
			return fmt.Errorf("prefix cannot be empty, e.g., %s", strings.Join(eg, ", "))
		} else {
			var matched bool
			for _, p := range prefix {
				if strings.HasPrefix(args[1], p+"-") {
					matched = true
					break
				}
			}
			if !matched {
				eg := make([]string, 0, len(prefix))
				for _, p := range prefix {
					eg = append(eg, p+"-xxxx")
				}
				return fmt.Errorf("version must start with one of the prefixes: %s", strings.Join(eg, ", "))
			}
		}
		if args[2] == "" {
			return fmt.Errorf("develop branch cannot be empty")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		repoURL := args[0]
		version := args[1]
		branch := args[2]
		cloneRepository(repoURL, version, branch)
	},
}

func cloneRepository(repoURL, version, branch string) {
	repoName := regexpGitRepo.FindStringSubmatch(repoURL)[1]
	repoName = strings.TrimLeft(repoName, "/")
	repoName = strings.ReplaceAll(repoName, "/", "-") + "-" + version + "-" + branch
	repoPath := path.Join(workspaceDir, repoName)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		execCommand("git", "clone", repoURL, repoPath)
	}
	execCommand("git", "-C", repoPath, "fetch", "--all")
	if !branchExists(repoPath, mainBranch) {
		errLog("Main branch does not exist: %s, user can specify it with --branch|-b", mainBranch)
		os.Exit(1)
	}
	execCommand("git", "-C", repoPath, "checkout", mainBranch)
	execCommand("git", "-C", repoPath, "pull", "origin", mainBranch)
	if branchExists(repoPath, version) {
		execCommand("git", "-C", repoPath, "checkout", version)
		// pull latest changes
		execCommand("git", "-C", repoPath, "pull", "origin", version)
		// merge main into feat branch
		execCommand("git", "-C", repoPath, "merge", "--no-ff", "-m",
			fmt.Sprintf("[Branch Merge] Merge %s into %s", mainBranch, version), mainBranch)
	} else {
		execCommand("git", "-C", repoPath, "checkout", "-b", version, mainBranch)
	}
	execCommand("git", "-C", repoPath, "push", "--set-upstream", "origin", version)
	successLog("project dir is: [" + repoPath + "]")
	if err := os.Chdir(repoPath); err != nil {
		errLog("Change directory to %s failed: %v", repoPath, err)
	}

	if config.OpenInIDEAfterUse {
		openByIDEA(repoPath)
	}

}
