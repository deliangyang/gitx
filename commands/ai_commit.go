package commands

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

var (
	aiConfirm bool
)

//go:embed prompts/github.prompt
var githubPrompt string

//go:embed prompts/default.prompt
var defaultPrompt string

func init() {
	AICommitCmd.Flags().BoolVarP(&aiConfirm, "yes", "y", false, "Auto confirm AI generated commit message")
}

var AICommitCmd = &cobra.Command{
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
		sp := defaultPrompt
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
