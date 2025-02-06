package git

import (
	"bytes"
	"os/exec"
	"strings"
)

type GitCommand struct {
	WorkingDir string
}

func NewGitCommand(workingDir string) *GitCommand {
	return &GitCommand{WorkingDir: workingDir}
}

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

func (g *GitCommand) GetDiff() (string, error) {
	return g.runCommand("diff")
}

func (g *GitCommand) GetStagedFiles() ([]string, error) {
	output, err := g.runCommand("ls-files", "--staged", "--modified", "--exclude-standard")
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

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

func (g *GitCommand) Commit(message string) error {
	_, err := g.runCommand("commit", "-m", message)
	return err
}

func (g *GitCommand) Stage(files []string) error {
	args := append([]string{"add"}, files...)
	_, err := g.runCommand(args...)
	return err
}

func (g *GitCommand) Unstage(files []string) error {
	args := append([]string{"reset", "HEAD", "--"}, files...)
	_, err := g.runCommand(args...)
	return err
}

func (g *GitCommand) StageFile(file string) error {
	return g.Stage([]string{file})
}

func (g *GitCommand) UnstageFile(file string) error {
	return g.Unstage([]string{file})
}
