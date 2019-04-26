package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// cmdErr will return and error with the output from stderr from the error if
// it exists.
func cmdErr(err error) error {
	if eerr, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(eerr.Stderr))
		if stderr != "" {
			return errors.New(stderr)
		}
	}
	return err
}

func gitRemoteURL(path string) (string, error) {
	cmd := exec.Command("git", "config", "remote.origin.url")
	cmd.Dir = path

	output, err := cmd.Output()

	return strings.TrimSpace(string(output)), cmdErr(err)
}

func gitClone(uri, dest string) error {
	cmd := exec.Command("git", "clone",
		"-q", "--config", "push.default=upstream", "--recursive",
		uri, dest)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmdErr(cmd.Run())
}

// var errGitAlreadyInitalized = errors.New("git already initialized")

func gitInit(path string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	return cmdErr(cmd.Run())
}

func gitRemoteExists(path, name string) bool {
	existingRemote, _ := gitConfig(path, "remote."+name+".url")
	return existingRemote != ""
}

func gitRemoteAdd(path, name, url string) error {
	cmd := exec.Command("git", "remote", "add", name, url)
	cmd.Dir = path
	return cmdErr(cmd.Run())
}

func gitConfig(path, opt string) (string, error) {
	cmd := exec.Command("git", "config", opt)
	cmd.Dir = path
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), cmdErr(err)
}

func isGitDir(path string) bool {
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		return true
	}
	return false
}

func gitCommitAll(path, msg string) error {
	args := []string{"commit", "-a"}
	if msg != "" {
		args = append(args, "-m", msg)
	}
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Dir = path
	return cmdErr(cmd.Run())

}

func gitDiff(path string) error {
	cmd := exec.Command("git", "diff")
	cmd.Stdout = os.Stdout
	cmd.Dir = path
	return cmdErr(cmd.Run())
}

func gitPull(path string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = path
	return cmdErr(cmd.Run())
}

func gitPush(path string) error {
	cmd := exec.Command("git", "push")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Dir = path
	return cmdErr(cmd.Run())
}

func gitStatus(path string) error {
	cmd := exec.Command("git", "status")
	cmd.Stdout = os.Stdout
	cmd.Dir = path
	return cmdErr(cmd.Run())

}
