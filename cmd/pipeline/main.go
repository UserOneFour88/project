package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"pipelineapp/internal/api"
	"pipelineapp/internal/format"
	"pipelineapp/internal/model"
	"pipelineapp/internal/pipeline"
)

func main() {
	userID := flag.Int("user", 0, "filter posts by userId (0 = all users)")
	limit := flag.Int("limit", 10, "max number of posts to process (0 = no limit)")
	workers := flag.Int("workers", 3, "number of fan-out workers")
	timeout := flag.Duration("timeout", 7*time.Second, "HTTP timeout, e.g. 5s")
	flag.Parse()

	if *workers < 1 {
		exitErr(fmt.Errorf("invalid -workers: must be >= 1"))
	}
	if *limit < 0 {
		exitErr(fmt.Errorf("invalid -limit: must be >= 0"))
	}
	if *userID < 0 {
		exitErr(fmt.Errorf("invalid -user: must be >= 0"))
	}

	client := api.NewClient(*timeout)
	posts, err := client.FetchPosts()
	if err != nil {
		exitErr(err)
	}

	summaries, err := runPipeline(posts, *userID, *limit, *workers)
	if err != nil {
		exitErr(err)
	}

	fmt.Print(format.Summaries(summaries))
}

func runPipeline(posts []model.Post, userID int, limit int, workers int) ([]model.PostSummary, error) {
	stage1In := pipeline.Source(posts)
	stage1Out := pipeline.FilterAndMap(stage1In, userID, limit)
	workerOut := pipeline.FanOutBodyWordCount(posts, stage1Out, workers)
	merged := pipeline.FanIn(workerOut...)

	var result []model.PostSummary
	for item := range merged {
		result = append(result, item)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no posts matched your filters")
	}
	return result, nil
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
