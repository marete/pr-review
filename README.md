# PR Review Scripts with Claude

Thorough PR review tools that use Claude's API with extended thinking (ultrathink) mode enabled by default for deep code analysis.

## Features

- 🧠 **Extended Thinking Mode (Ultrathink)**: Deep analysis with Claude's extended reasoning
- 🔍 **Comprehensive Review**: Covers code quality, security, performance, testing, and more
- 📊 **Full Context**: Includes diff, changed files, and commit messages
- 🎯 **Flexible**: Multiple language implementations (Go, Python, Bash)
- ⚙️ **Configurable**: Control thinking budget, model selection, and more

## Prerequisites

1. **Anthropic API Key**: Set your API key as an environment variable:
   ```bash
   export ANTHROPIC_API_KEY='your-api-key-here'
   ```

2. **Git Repository**: Run from within a git repository with changes to review

## Go Version (Recommended)

### Installation

```bash
# Build the binary (from the source's top directory)
go build -v

# Optional: Install globally
sudo mv pr-review /usr/local/bin/
```

### Usage

```bash
# Review current branch against main (ultrathink enabled by default)
pr-review

# Review against a specific branch
pr-review -branch develop

# Compare two specific commits
pr-review -base abc123f

# Disable ultrathink mode
pr-review -no-ultrathink

# Use a different model
pr-review -model claude-opus-4-20250514

# Adjust thinking budget (default: 10000 tokens)
pr-review -thinking-budget 20000

# Include additional context files
pr-review -context README.md,ARCHITECTURE.md
```

### Options

- `-branch`: Target branch to compare against (default: main/master)
- `-base`: Base commit/branch to compare from
- `-model`: Claude model to use (default: claude-sonnet-4-5-20250929)
- `-no-ultrathink`: Disable extended thinking mode
- `-thinking-budget`: Token budget for extended thinking (default: 10000)
- `-context`: Comma-separated list of additional context files
