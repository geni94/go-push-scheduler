package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/google/go-github/v39/github"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
)

var (
	tokenFlag    = flag.String("token", "", "Your GitHub Personal Access Token. Provide either this, or the username and password.")
	usernameFlag = flag.String("username", "", "Your GitHub username. Provide either this, or the token.")
	dateFlag     = flag.String("date", "", "When to push the commit, in the format: dd-mm-yyyy hh:mm")
	pathFlag     = flag.String("path", "", "Absolute path to the repository to push.")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Git Push Scheduler: Commit now, push whenever you want.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n")
		flag.PrintDefaults()
	}
}

func parseDateTime(dateTimeStr string) (time.Time, error) {
	layout := "02-01-2006 15:04"
	return time.Parse(layout, dateTimeStr)
}

func askForConfirmation() bool {
	var response string
	fmt.Print("No upstream branch set. Would you like to set it automatically? [y/n]: ")
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func promptForPassword() string {
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Error reading password: %v", err)
	}
	fmt.Println()
	return string(bytePassword)
}

func main() {
	flag.Parse()

	if flag.NFlag() == 0 {
		// return Usage func if no arguments are passed
		flag.Usage()
		return
	}

	if *tokenFlag == "" && *usernameFlag == "" {
		fmt.Println("A GitHub Username or a Personal Access Token is required")
		return
	}

	if *pathFlag == "" {
		log.Fatal("Path to the repository is required")
		return
	}
	var tc *http.Client

	ctx := context.Background()

	if *tokenFlag != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *tokenFlag},
		)
		tc = oauth2.NewClient(ctx, ts)
	} else if *usernameFlag != "" {
		password := promptForPassword()
		tp := github.BasicAuthTransport{
			Username: *usernameFlag,
			Password: password,
		}
		tc = tp.Client()
	} else {
		log.Fatal("GitHub Personal Access Token or username is required")
	}
	client := github.NewClient(tc)
	// The `client` variable is now authenticated and can be used to interact with GitHub.
	// print the authenticated user
	user, _, err := client.Users.Get(ctx, "")

	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(*user.Login)

	// schedule the push
	if *dateFlag != "" {
		scheduledTime, err := parseDateTime(*dateFlag)
		if err != nil {
			log.Fatalf("Invalid date format: %v", err)
		}

		upstreamExists, err := hasUpstreamBranch(*pathFlag)
		if err != nil {
			log.Fatal(err)
		}

		setUpstream := false
		if !upstreamExists {
			fmt.Println("The current branch does not have an upstream branch set.")
			setUpstream = askForConfirmation()
		}

		// Schedule the push in a separate goroutine
		go func() {
			time.Sleep(time.Until(scheduledTime))
			gitPush(*pathFlag, setUpstream)
		}()
	} else {
		log.Fatal("A date to schedule the commit is required")
	}
}
