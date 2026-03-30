package model

// Post is the source shape returned by jsonplaceholder.
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// PostSummary is the shape produced by the pipeline.
type PostSummary struct {
	ID        int
	UserID    int
	Title     string
	BodyWords int
}
