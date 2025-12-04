package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

var (
	regexpGitRepo    = regexp.MustCompile(`git@[^:]+:([^\.]+).git$`)
	regexpSplitSpace = regexp.MustCompile(`\s+`)
	regexpProject    = regexp.MustCompile(`((feat|online-fix|online-revision)-[^-]+)-(.+)$`)
	regexpSplitter   = regexp.MustCompile(`(feat:|test:|revert:|chore:|style:|refactor:|fix:|\n|\r)`)
	prefix           = []string{
		"feat",
		"online-fix",
		"online-revision",
	}
)

var (
	workspaceDir string
	isDebug      = os.Getenv("DEBUG") == "true"
	stableBranch string
	aiConfirm    bool
)

var cloneCmd = &cobra.Command{
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
	if !branchExists(repoPath, stableBranch) {
		errLog("Stable branch does not exist: %s, user can specify it with --branch|-b", stableBranch)
		os.Exit(1)
	}
	execCommand("git", "-C", repoPath, "checkout", stableBranch)
	execCommand("git", "-C", repoPath, "pull", "origin", stableBranch)
	if branchExists(repoPath, version) {
		execCommand("git", "-C", repoPath, "checkout", version)
		// pull latest changes
		execCommand("git", "-C", repoPath, "pull", "origin", version)
		// merge stable into feat branch
		execCommand("git", "-C", repoPath, "merge", "--no-ff", "-m",
			fmt.Sprintf("[Branch Merge] Merge %s into %s", stableBranch, version), stableBranch)
	} else {
		execCommand("git", "-C", repoPath, "checkout", "-b", version, stableBranch)
	}
	execCommand("git", "-C", repoPath, "push", "--set-upstream", "origin", version)
	successLog("project dir is: [" + repoPath + "]")
	if err := os.Chdir(repoPath); err != nil {
		errLog("Change directory to %s failed: %v", repoPath, err)
	}

	openByIDEA(repoPath)
}

func openByIDEA(repoPath string) {
	// select idea to open the project
	openPrompt := promptui.Select{
		Label: "Open project in IDEA?",
		Items: []string{"code", "goland", "pstorm", "no"},
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

var syncTargetBranchCmd = &cobra.Command{
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

var fetchCmd = &cobra.Command{
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

var renameProjectCmd = &cobra.Command{
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

		openByIDEA(newPath)
	},
}

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Switch to a specific project directory in the workspace",
	Run: func(cmd *cobra.Command, args []string) {
		prompt := promptui.Select{
			Label: "Select Project Directory",
			Items: func() []string {
				dirs, err := os.ReadDir(workspaceDir)
				if err != nil {
					errLog("failed to read workspace directory: %v", err)
				}
				var projectDirs []string
				for _, dir := range dirs {
					if dir.IsDir() {
						projectDirs = append(projectDirs, dir.Name())
					}
				}
				return projectDirs
			}(),
		}

		_, selectedDir, err := prompt.Run()
		if err != nil {
			errLog("Read selectedDir fail %v\n", err)
		}
		successLog("You selected %s", selectedDir)
		projectPath := path.Join(workspaceDir, selectedDir)
		if err := os.Chdir(projectPath); err != nil {
			errLog("Change directory to %s failed: %v", projectPath, err)
		}
		successLog("Changed directory to: %s", projectPath)

		openByIDEA(projectPath)
	},
}

var aiCommitCmd = &cobra.Command{
	Use:       "am [default|github]",
	Short:     "Generate AI-based commit messages, then push to remote, limit to 4000 characters diff",
	ValidArgs: []string{"default", "github"},
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("OPENAI_API_KEY") == "" {
			warningLog("OPENAI_API_KEY environment variable is not set.")
			return
		}
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

		diff := execCommandWithOutput("git", "diff", "--cached")
		if diff == "" {
			warningLog("use `git add .` first")
			errLog("No changes detected.")
		} else if len(diff) > 140000 {
			warningLog("diff is too large (>140000 characters), please commit manually")
			errLog("Diff too large.")
		}
		sp := sysPrompt
		var isGithub bool
		if len(args) > 0 && args[0] == "github" {
			sp = githubPrompt
			isGithub = true
		}
		userMessage := "以下是 git diff 内容：\n" + diff
		client := openai.NewClient() // defaults to os.LookupEnv("OPENAI_API_KEY")
		chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(sp),
				openai.UserMessage("正文内容一律使用中文"),
				openai.UserMessage(userMessage),
			},
			Model: openai.ChatModelGPT5_1,
		})
		if err != nil {
			errLog("OpenAI API error: %v", err)
		}
		commitMsg := strings.TrimSpace(chatCompletion.Choices[0].Message.Content)
		log.Println(commitMsg)

		if !aiConfirm {
			log.Print("Do you want to use this commit message and then push? (y/n): ")
			var input string
			fmt.Scanln(&input)
			if strings.ToLower(input) != "y" {
				warningLog("Commit aborted.")
				return
			}
		}
		commitArgs := formatCommitMessage(commitMsg, isGithub)
		execCommand("git", commitArgs...)
		successLog("Committed with AI-generated message.")
		execCommand("git", "push", "--set-upstream", "origin", version)
		successLog("Pushed to remote repository.")
	},
}

func formatCommitMessage(commitMsg string, isGithub bool) []string {
	if isGithub {
		msg := strings.Split(commitMsg, "\n")
		args := []string{"commit"}
		for _, line := range msg {
			line = strings.TrimSpace(line)
			if line != "" {
				args = append(args, "-m", line)
			}
		}
		return args
	}
	// find multiple lines starting with feat:|test:|revert:|chore:|style:|refactor:|fix:
	matchedPrefix := regexpSplitter.FindAllString(commitMsg, -1)
	if len(matchedPrefix) == 0 {
		errLog("No valid commit message prefix found in AI response.")
	}
	prefixCounter := make(map[string]int)
	for _, prefix := range matchedPrefix {
		if prefix != "\n" && prefix != "\r" {
			prefix = strings.TrimSuffix(prefix, ":")
			prefixCounter[prefix]++
		}
	}
	var finalPrefix string
	maxCount := 0
	for prefix, count := range prefixCounter {
		if count > maxCount {
			maxCount = count
			finalPrefix = prefix
		}
	}
	if isDebug {
		log.Printf("finalPrefix: [%s]\n", finalPrefix)
	}

	splitCommitMsg := regexpSplitter.Split(commitMsg, -1)
	filteredLines := []string{}
	for _, line := range splitCommitMsg {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			filteredLines = append(filteredLines, trimmedLine)
		}
	}
	splitCommitMsg = filteredLines

	commitArgs := []string{}
	commitArgs = append(commitArgs, "commit")
	for idx, line := range splitCommitMsg {
		line = strings.TrimSpace(line)
		if idx == 0 && !strings.HasPrefix(line, finalPrefix) {
			line = fmt.Sprintf("%s: %s", finalPrefix, line)
		}
		commitArgs = append(commitArgs, "-m", line)
	}
	return commitArgs
}

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select common projects to clone, like `gitx select` or `gitx select -b main`",
	Run: func(cmd *cobra.Command, args []string) {
		prompt := promptui.Select{
			Label: "Select Repository to Clone",
			Items: commonProjects,
		}

		_, repoUrl, err := prompt.Run()
		if err != nil {
			errLog("Read repoUrl fail %v\n", err)
		}
		successLog("You selected %s", repoUrl)
		prefixPrompt := promptui.Select{
			Label: "Enter prefix for version",
			Items: prefix,
		}
		_, prefix, err := prefixPrompt.Run()
		if err != nil {
			errLog("Read prefix fail %v\n", err)
		}
		versionPrompt := promptui.Prompt{
			Label: "Enter version (e.g., 20250101 or 1.2.3, not containing prefix and -)",
		}
		version, err := versionPrompt.Run()
		if err != nil {
			errLog("Read version fail %v\n", err)
		}
		branchPrompt := promptui.Prompt{
			Label: "Enter develop branch",
		}
		branch, err := branchPrompt.Run()
		if err != nil {
			errLog("Read branch fail %v\n", err)
		} else if branch == "" {
			errLog("Branch cannot be empty")
		}
		cloneRepository(repoUrl, prefix+"-"+version, branch)
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install gitx tool",
	Run: func(cmd *cobra.Command, args []string) {
		execCommand("go", "install", "github.com/deliangyang/gitx@latest")
		successLog("gitx installed successfully.")
	},
}

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Show documentation",
	Run: func(cmd *cobra.Command, args []string) {
		successLog(`Documentation is available at: https://github.com/deliangyang/gitx/tree/main`)
	},
}

var rootCmd = &cobra.Command{
	Use:   "gitx",
	Short: "A CLI tool for advanced Git operations",
	Long: `A CLI tool for advanced Git operations.
It provides various commands to manage Git repositories efficiently.`,
}

func getHomeDir() string {
	workspaceDir := os.Getenv("WORKSPACE_DIR")
	if workspaceDir != "" {
		return workspaceDir
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

var (
	sysPrompt = `
你是一个资深的代码提交信息生成助手，能够根据 git diff 内容生成简洁且准确的中文提交信息。
请严格按照以下要求生成提交信息：
1. 将以下 git diff，联系上下文信息，总结为一行或者多行中文提交消息，如果有多个内容的提交，请用列出 1,2,3,4 点等，换行分隔
2. 注意根据内容仅给提交消息添加一个前缀 (feat|test|revert|chore|style|refactor|fix):等，后面的任何内容不需要添加
3. 注意 diff 内容中，每行前缀 "+++" 表示新增，前缀 "---" 表示删除，前缀 " " 表示未改动
4. 仅总结代码改动的行，可以联系上下文，不要添加多余的内容
5. 忽略新增或者删除注释，空行，格式化等无意义，多余，不必要的改动
6. 没有内容可以总结时，回复 "style: 格式化代码"
7. 最后对总结的提交消息列表进行去重，重新编号，确保每一行内容大概意思不重复
8. 返回内容去掉 diff 信息，不能包含 diff 的代码
9. 只需要总结出一个前缀 (feat|test|revert|chore|style|refactor|fix):开头的提交消息，不能再内容中添加多余的 (feat|test|revert|chore|style|refactor|fix):前缀
10. 提交消息内容使用中文描述，越简洁越好`
)

//go:embed github.prompt
var githubPrompt string

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

var (
	version        = "v0.1.2"
	repoPrefix     = "git@github.com:"
	commonProjects = []string{}
)

func init() {

	for idx, project := range commonProjects {
		repoURL := repoPrefix + project + ".git"
		commonProjects[idx] = repoURL
	}

	workspaceDir = getHomeDir()
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		errLog("Make work space dir fail")
	}

	cloneCmd.Flags().StringVarP(&stableBranch, "branch", "b", "stable", "Stable branch name, default is 'stable'")
	cloneCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", workspaceDir, "Workspace directory")

	fetchCmd.Flags().StringVarP(&stableBranch, "branch", "b", "stable", "Stable branch name, default is 'stable'")

	aiCommitCmd.Flags().BoolVarP(&aiConfirm, "yes", "y", false, "Auto confirm AI generated commit message")

	selectCmd.Flags().StringVarP(&stableBranch, "branch", "b", "stable", "Stable branch name, default is 'stable'")
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

func main() {
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(syncTargetBranchCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(aiCommitCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(selectCmd)
	rootCmd.AddCommand(docCmd)
	rootCmd.AddCommand(renameProjectCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		errLog("Error: %s", err.Error())
	}
}
