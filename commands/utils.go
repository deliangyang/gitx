package commands

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"
)

var (
	regexpGitRepo    = regexp.MustCompile(`git@[^:]+:([^\.]+).git$`)
	regexpSplitSpace = regexp.MustCompile(`\s+`)
	regexpSplitter   = regexp.MustCompile(`(feat:|test:|revert:|chore:|style:|refactor:|fix:|\n|\r)`)
)

var (
	isDebug = os.Getenv("DEBUG") == "true"
	ideas   = []string{"code", "goland", "pstorm", "no"}
)

func openByIDEA(repoPath string) {
	sort.Slice(ideas, func(i, j int) bool {
		return ideas[i] != config.DefaultIDE && ideas[j] != config.DefaultIDE
	})
	openPrompt := promptui.Select{
		Label: "Open project in IDEA?",
		Items: ideas,
	}
	_, openCmd, err := openPrompt.Run()
	if err != nil {
		errLog("Read openCmd fail %v\n", err)
	}
	if openCmd != "no" {
		if !commandExists(openCmd) {
			errLog("command [%s] not found in PATH", openCmd)
		}
		execCommand(openCmd, repoPath)
	}
}

func errLog(format string, a ...interface{}) {
	log.Printf("\033[31m"+format+"\033[0m\n", a...)
	os.Exit(1)
}

func successLog(format string, a ...interface{}) {
	log.Printf("\033[32m"+format+"\033[0m\n", a...)
}

func warningLog(format string, a ...interface{}) {
	log.Printf("\033[33m"+format+"\033[0m\n", a...)
}

func getHomeDir() string {
	workspaceDir := os.Getenv("WORKSPACE_DIR")
	if workspaceDir != "" {
		return workspaceDir
	} else if config.WorkspaceDir != "" {
		return config.WorkspaceDir
	}
	u, err := user.Current()
	if err != nil {
		errLog("failed to get current user: %v", err)
	}
	return path.Join(u.HomeDir, "work")
}

func execCommand(name string, args ...string) {
	// Placeholder for command execution logic
	cmd := exec.Command(name, args...)
	data, err := cmd.CombinedOutput()
	if isDebug {
		log.Println(cmd.String())
	}
	log.Println(strings.TrimSpace(string(data)))
	if err != nil {
		errLog("something went wrong: %s [%s]", err.Error(), cmd.String())
	}
}

func execCommandWithOutput(name string, args ...string) string {
	// Placeholder for command execution logic with output
	cmd := exec.Command(name, args...)
	data, err := cmd.CombinedOutput()
	if isDebug {
		log.Println(cmd.String())
	}
	log.Println(strings.TrimSpace(string(data)))
	if err != nil {
		errLog("something went wrong: %s [%s]", err.Error(), cmd.String())
	}
	return strings.TrimSpace(string(data))
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func branchExists(repoPath, branch string) bool {
	output := execCommandWithOutput("git", "-C", repoPath, "branch", "--list", branch)
	if output != "" {
		return true
	}
	output = execCommandWithOutput("git", "-C", repoPath, "ls-remote", "--heads", "origin", branch)
	return output != ""
}
