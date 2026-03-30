package pipeline

import (
	"strings"
	"sync"

	"pipelineapp/internal/model"
)

func Source(posts []model.Post) <-chan model.Post {
	out := make(chan model.Post)
	go func() {
		defer close(out)
		for _, p := range posts {
			out <- p
		}
	}()
	return out
}

// FilterAndMap is the first stage of pipeline.
func FilterAndMap(in <-chan model.Post, userID int, limit int) <-chan model.PostSummary {
	out := make(chan model.PostSummary)
	go func() {
		defer close(out)
		count := 0
		for p := range in {
			if userID > 0 && p.UserID != userID {
				continue
			}
			out <- model.PostSummary{
				ID:     p.ID,
				UserID: p.UserID,
				Title:  strings.TrimSpace(p.Title),
			}
			count++
			if limit > 0 && count >= limit {
				return
			}
		}
	}()
	return out
}

// FanOutBodyWordCount starts N workers and enriches summaries with word counts.
func FanOutBodyWordCount(posts []model.Post, in <-chan model.PostSummary, workers int) []<-chan model.PostSummary {
	if workers < 1 {
		workers = 1
	}

	lookup := make(map[int]string, len(posts))
	for _, p := range posts {
		lookup[p.ID] = p.Body
	}

	workerOutputs := make([]<-chan model.PostSummary, 0, workers)
	for i := 0; i < workers; i++ {
		out := make(chan model.PostSummary)
		workerOutputs = append(workerOutputs, out)
		go func(workerOut chan<- model.PostSummary) {
			defer close(workerOut)
			for s := range in {
				body := lookup[s.ID]
				s.BodyWords = len(strings.Fields(body))
				workerOut <- s
			}
		}(out)
	}
	return workerOutputs
}

// FanIn merges many channels into one.
func FanIn(inputs ...<-chan model.PostSummary) <-chan model.PostSummary {
	out := make(chan model.PostSummary)
	var wg sync.WaitGroup
	wg.Add(len(inputs))

	for _, ch := range inputs {
		go func(c <-chan model.PostSummary) {
			defer wg.Done()
			for item := range c {
				out <- item
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
