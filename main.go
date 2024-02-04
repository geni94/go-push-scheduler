package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

var (
	tokenFlag = flag.String("token", "", "GitHub Personal Access Token")
	dateFlag  = flag.String("date", "", "When to push the commit, in the format: dd-mm-yyyy hh:mm")
	pathFlag  = flag.String("path", "", "Path to the repository to commit")
)

func gitCommit(message string) {
	cmd := exec.Command("git", "commit", "-m", message)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func gitPush(repoPath string) {
	cmd := exec.Command("git", "-C", repoPath, "push", "origin", "HEAD")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func parseDateTime(dateTimeStr string) (time.Time, error) {
	layout := "02-01-2006 15:04"
	return time.Parse(layout, dateTimeStr)
}

func main() {
	flag.Parse()

	if *tokenFlag == "" {
		fmt.Println("GitHub Personal Access Token is required")
		return
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *tokenFlag},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	// The `client` variable is now authenticated and can be used to interact with GitHub.
	// print the authenticated user
	user, _, err := client.Users.Get(ctx, "")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(*user.Login)

	if *pathFlag == "" {
		log.Fatal("Path to the repository is required")
	}

	if *dateFlag != "" {
		scheduledTime, err := parseDateTime(*dateFlag)
		if err != nil {
			log.Fatalf("Invalid date format: %v", err)
		}

		// wait until the scheduled time in a separate goroutine
		go func() {
			time.Sleep(time.Until(scheduledTime))
			gitPush(*pathFlag)
		}()
	} else {
		// return an error if the date is not provided
		log.Fatal("A date to schedule the commit is required")
	}
}
