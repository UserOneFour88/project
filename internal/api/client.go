package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pipelineapp/internal/model"
)

const baseURL = "https://jsonplaceholder.typicode.com/posts"

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) FetchPosts() ([]model.Post, error) {
	resp, err := c.httpClient.Get(baseURL)
	if err != nil {
		return nil, fmt.Errorf("request posts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var posts []model.Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, fmt.Errorf("decode posts response: %w", err)
	}
	return posts, nil
}
