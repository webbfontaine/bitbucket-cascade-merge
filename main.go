package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// initialize a buffered channel to process merges one at the time
	events := make(chan PullRequestEvent, 100)
	go worker(events)

	log.Printf("Server running on port %s, using token : %s", getEnv("PORT", "5000"), getEnv("TOKEN", ""))

	// start the hook listener
	handler := NewEventHandler(events)
	addr := fmt.Sprintf(":%s", getEnv("PORT", "5000"))
	http.Handle("/", handler.CheckToken(getEnv("TOKEN", ""), handler.Handle()))
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("cannot start server on %s", addr)
	}

	close(events)
}

func worker(event <-chan PullRequestEvent) {
	for e := range event {

		log.Printf("Processing cascade merge for %s from %s to %s", e.Repository.Name, e.PullRequest.Source.Branch.Name, e.PullRequest.Destination.Branch.Name)

		// retrieve auth from environment
		username := getEnv("BITBUCKET_USERNAME", "")
		password := getEnv("BITBUCKET_PASSWORD", "")

		// get the clone url which is not provided in the webhook
		api := NewBitbucket(username, password, e.Repository.Owner.UUID, e.Repository.Name)
		url, err := api.GetCloneURL("https")
		if err != nil {
			log.Printf("cannot read clone url of %s (owner=%s): %s", e.Repository.Name, e.Repository.Owner.UUID, err)
			continue
		}

		c, err := NewClient(&ClientOptions{
			Path: filepath.Join(os.TempDir(), e.Repository.Uuid),
			URL:  url,
			Credentials: &Credentials{
				Username: username,
				Password: password,
			},
		})

		if err != nil {
			log.Printf("failed to initialize git repository: %s", err)
		}

		// query repository branching model to know which branches are candidate for cascading
		opts, err := api.GetCascadeOptions(e.Repository.Owner.UUID, e.Repository.Name)
		if err != nil {
			log.Printf("cannot detect cascade options for %s, check branching model", e.Repository.Name)
			continue
		}

		// check destination branch is candidate for auto merge
		destination := e.PullRequest.Destination.Branch.Name
		if strings.HasPrefix(destination, opts.DevelopmentName) && !strings.HasPrefix(destination, opts.ReleasePrefix) {
			continue
		}

		// cascade merge the pull request
		state := c.CascadeMerge(e.PullRequest.Destination.Branch.Name, opts)
		if state != nil {

			if state.Source != "" && state.Target != "" {
				// create a new pull request when cascade fails
				err := api.CreatePullRequest(
					"Automatic merge failure",
					"There was a merge conflict automatically merging this branch",
					state.Source,
					state.Target)

				if err != nil {
					log.Printf("could not create a pull request %s to %s on %s", state.Source, state.Target, e.Repository.Name)
				} else {
					log.Printf("Pull request created from %s to %s on %s", state.Source, state.Target, e.Repository.Name)
				}
			}else{
        log.Printf("Error while doing cascade merge for repo %s : %s", e.Repository.Name, state.Error())
      }
		}
	}
}
