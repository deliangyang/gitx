package commands

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

var (
	aiConfirm   bool
	autoAdd     bool
	limitLength int
)

//go:embed prompts/github.prompt
var githubPrompt string

//go:embed prompts/default.prompt
var defaultPrompt string

func init() {
	AICommitCmd.Flags().BoolVarP(&aiConfirm, "yes", "y", false, "Auto confirm AI generated commit message")
	AICommitCmd.Flags().BoolVarP(&autoAdd, "add", "a", false, "Auto git add . before generating commit message")
	AICommitCmd.Flags().IntVarP(&limitLength, "limit", "l", 10000, "Set the maximum length of git diff to be processed")
}

var AICommitCmd = &cobra.Command{
	Use:       "am [default|github]",
	Short:     fmt.Sprintf("Generate AI-based commit messages, then push to remote, limit to %d characters diff", limitedLen),
	ValidArgs: []string{"default", "github"},
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("OPENAI_API_KEY") == "" {
			warningLog("OPENAI_API_KEY environment variable is not set.")
			return
		}
		if autoAdd {
			execCommand("git", "add", ".")
			successLog("Auto git add . executed.")
		}
		diff := execCommandWithOutput("git", "diff", "--cached")
		if diff == "" {
			if !autoAdd {
				warningLog("use `git add .` first")
			}
			errLog("No changes detected.")
		} else if len(diff) > limitLength {
			warningLog(fmt.Sprintf("diff is too large (>%d characters), please commit manually", limitLength))
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
		// pull first, auto merge
		execCommand("git", "pull", "--no-edit")
		// then push
		execCommand("git", "push")
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
