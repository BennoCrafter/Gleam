// Package git provides a wrapper for Git command line operations
package git

import (
	"bytes"
	"os/exec"
	"strings"
)

// GitCommand represents a Git command executor with a working directory
type GitCommand struct {
	WorkingDir string
}

// NewGitCommand creates a new GitCommand instance with the specified working directory
func NewGitCommand(workingDir string) *GitCommand {
	return &GitCommand{WorkingDir: workingDir}
}

// runCommand executes a git command with the given arguments and returns its output
func (g *GitCommand) runCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.WorkingDir

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

// GetDiff returns the diff of all changes in the working directory
func (g *GitCommand) GetDiff() (string, error) {
	return g.runCommand("diff")
}

// GetFileDiff returns the diff for a specific file
func (g *GitCommand) GetFileDiff(file string) (string, error) {
	return g.runCommand("diff", file)
}

// GetStagedFiles returns a list of files that are staged for commit
func (g *GitCommand) GetStagedFiles() ([]string, error) {
	output, err := g.runCommand("--name-only", "--cached")
	if output == "" {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// GetUnstagedFiles returns a list of files that have changes but are not staged
func (g *GitCommand) GetUnstagedFiles() ([]string, error) {
	output, err := g.runCommand("ls-files", "--others", "--modified", "--exclude-standard")
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// Commit creates a new commit with the given message
func (g *GitCommand) Commit(message string) error {
	_, err := g.runCommand("commit", "-m", message)
	return err
}

// Add stages the specified files for commit
func (g *GitCommand) Add(files []string) error {
	args := append([]string{"add"}, files...)
	_, err := g.runCommand(args...)
	return err
}

// Reset unstages the specified files
func (g *GitCommand) Reset(files []string) error {
	args := append([]string{"reset", "HEAD", "--"}, files...)
	_, err := g.runCommand(args...)
	return err
}

// Push pushes commits to the remote repository
func (g *GitCommand) Push() error {
	_, err := g.runCommand("push")
	return err
}

// Pull fetches and merges changes from the remote repository
func (g *GitCommand) Pull() error {
	_, err := g.runCommand("pull")
	return err
}

// Fetch downloads objects and refs from the remote repository
func (g *GitCommand) Fetch() error {
	_, err := g.runCommand("fetch")
	return err
}

// Stage adds the specified files to the staging area
func (g *GitCommand) Stage(files []string) error {
	args := append([]string{"add"}, files...)
	_, err := g.runCommand(args...)
	return err
}

// Unstage removes the specified files from the staging area
func (g *GitCommand) Unstage(files []string) error {
	args := append([]string{"reset", "HEAD", "--"}, files...)
	_, err := g.runCommand(args...)
	return err
}

// StageFile adds a single file to the staging area
func (g *GitCommand) StageFile(file string) error {
	return g.Stage([]string{file})
}

// UnstageFile removes a single file from the staging area
func (g *GitCommand) UnstageFile(file string) error {
	return g.Unstage([]string{file})

}
