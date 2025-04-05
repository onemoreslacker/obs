package entities

// StackOverflowUpdate represents answer/comment activity via tracking link.
type StackOverflowUpdate struct {
	Owner struct {
		Username string `json:"display_name"`
	} `json:"owner"`
	CreatedAt int64  `json:"creation_date"`
	Body      string `json:"body"`
}

// GitHubUpdate represents PR/issue activity via tracking link.
type GitHubUpdate struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	User  struct {
		Login string `json:"login"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
}
