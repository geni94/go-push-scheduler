package main

import (
	"log"
	"os/exec"
)

func gitCommit(message string) {
	cmd := exec.Command("git", "commit", "-m", message)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func gitPush(repoPath string, setUpstream bool) {
	args := []string{"-C", repoPath, "push"}
	if setUpstream {
		args = append(args, "-u")
	}
	args = append(args, "origin", "HEAD")

	cmd := exec.Command("git", args...)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func hasUpstreamBranch(repoPath string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil // No upstream branch
		}
		return false, err // Some other error occurred
	}
	return true, nil // Upstream branch exists
}
