package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	claudeAPIURL = "https://api.anthropic.com/v1/messages"
	apiVersion   = "2023-06-01"
)

type ClaudeRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature,omitempty"`
	Messages    []Message `json:"messages"`
	Thinking    *Thinking `json:"thinking,omitempty"`
}

type Thinking struct {
	Type   string `json:"type"`
	Budget int    `json:"budget_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeResponse struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
	Usage   Usage          `json:"usage"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func main() {
	// Command line flags
	branch := flag.String("branch", "", "Target branch to compare against (default: main or master)")
	base := flag.String("base", "", "Base branch/commit to compare from")
	model := flag.String("model", "claude-sonnet-4-5-20250929", "Claude model to use")
	noThinking := flag.Bool("no-ultrathink", false, "Disable extended thinking mode")
	thinkingBudget := flag.Int("thinking-budget", 10000, "Extended thinking token budget")
	contextFiles := flag.String("context", "", "Comma-separated list of additional context files to include")
	flag.Parse()

	// Get API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	// Determine target branch
	targetBranch := *branch
	if targetBranch == "" {
		targetBranch = getDefaultBranch()
	}

	// Get current branch
	currentBranch := getCurrentBranch()
	fmt.Printf("🔍 Reviewing changes on '%s' against '%s'\n\n", currentBranch, targetBranch)

	// Get the diff
	var diff string
	var err error
	if *base != "" {
		diff, err = getDiff(*base, "HEAD")
	} else {
		diff, err = getDiff(targetBranch, "HEAD")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting diff: %v\n", err)
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println("No changes found.")
		os.Exit(0)
	}

	// Get changed files summary
	changedFiles := getChangedFiles(targetBranch)

	// Get recent commit messages
	commitMessages := getRecentCommits(targetBranch)

	// Get additional context files if specified
	additionalContext := ""
	if *contextFiles != "" {
		files := strings.Split(*contextFiles, ",")
		for _, file := range files {
			file = strings.TrimSpace(file)
			content, err := os.ReadFile(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not read context file %s: %v\n", file, err)
				continue
			}
			additionalContext += fmt.Sprintf("\n\n--- Context from %s ---\n%s\n", file, string(content))
		}
	}

	// Build the prompt
	prompt := buildReviewPrompt(diff, changedFiles, commitMessages, additionalContext)

	// Call Claude API
	fmt.Println("🤖 Analyzing PR with Claude (ultrathink mode: enabled)...")
	fmt.Println("⏳ This may take a moment for deep analysis...\n")

	review, usage, err := callClaude(apiKey, *model, prompt, !*noThinking, *thinkingBudget)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling Claude API: %v\n", err)
		os.Exit(1)
	}

	// Print the review
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println("CODE REVIEW")
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println()
	fmt.Println(review)
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Printf("📊 Token Usage: Input: %d | Output: %d | Total: %d\n",
		usage.InputTokens, usage.OutputTokens, usage.InputTokens+usage.OutputTokens)
	fmt.Println("=" + strings.Repeat("=", 78))
}

func buildReviewPrompt(diff, changedFiles, commitMessages, additionalContext string) string {
	prompt := `You are an expert code reviewer. Please perform a thorough and comprehensive review of this Pull Request.

Your review should cover:

1. **Code Quality & Best Practices**
   - Design patterns and architecture
   - Code organization and structure
   - Naming conventions and readability
   - DRY principle adherence
   - SOLID principles where applicable

2. **Potential Issues**
   - Bugs or logic errors
   - Edge cases not handled
   - Race conditions or concurrency issues
   - Memory leaks or performance problems
   - Security vulnerabilities

3. **Testing**
   - Test coverage adequacy
   - Missing test cases
   - Test quality and effectiveness

4. **Performance**
   - Algorithmic complexity
   - Database query efficiency
   - Resource usage (memory, CPU, network)
   - Caching opportunities

5. **Security**
   - Input validation
   - Authentication/authorization issues
   - SQL injection, XSS, or other vulnerabilities
   - Secrets or sensitive data exposure

6. **Maintainability**
   - Documentation quality
   - Code complexity
   - Technical debt introduced
   - Future extensibility

7. **Specific Suggestions**
   - Concrete code improvements
   - Alternative approaches
   - Refactoring opportunities

Please be thorough but constructive. Highlight both concerns and things done well.

---

## Changed Files
` + "```\n" + changedFiles + "\n```\n\n"

	if commitMessages != "" {
		prompt += "## Recent Commit Messages\n```\n" + commitMessages + "\n```\n\n"
	}

	prompt += "## Full Diff\n```diff\n" + diff + "\n```\n"

	if additionalContext != "" {
		prompt += "\n## Additional Context\n" + additionalContext + "\n"
	}

	prompt += "\n\nPlease provide your comprehensive code review."

	return prompt
}

func callClaude(apiKey, model, prompt string, useThinking bool, thinkingBudget int) (string, Usage, error) {
	req := ClaudeRequest{
		Model:       model,
		MaxTokens:   16000,
		Temperature: 1.0,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Enable extended thinking if requested
	if useThinking {
		req.Thinking = &Thinking{
			Type:   "enabled",
			Budget: thinkingBudget,
		}
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("error marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", claudeAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", Usage{}, fmt.Errorf("error creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", Usage{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", Usage{}, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", Usage{}, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", Usage{}, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Combine all text content blocks
	var reviewText strings.Builder
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			reviewText.WriteString(block.Text)
		}
	}

	return reviewText.String(), claudeResp.Usage, nil
}

func getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func getDefaultBranch() string {
	// Try to get the default branch from remote
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		branch = strings.TrimPrefix(branch, "refs/remotes/origin/")
		if branch != "" {
			return branch
		}
	}

	// Fallback: check if main exists, otherwise use master
	cmd = exec.Command("git", "rev-parse", "--verify", "main")
	if cmd.Run() == nil {
		return "main"
	}

	return "master"
}

func getDiff(base, head string) (string, error) {
	cmd := exec.Command("git", "diff", base+"..."+head)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getChangedFiles(baseBranch string) string {
	cmd := exec.Command("git", "diff", "--name-status", baseBranch+"...HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "Error getting changed files"
	}
	return strings.TrimSpace(string(output))
}

func getRecentCommits(baseBranch string) string {
	cmd := exec.Command("git", "log", baseBranch+"..HEAD", "--pretty=format:%h - %s (%an, %ar)")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
