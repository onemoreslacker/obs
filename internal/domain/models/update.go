package models

import (
	"fmt"
	"strings"
)

type StackOverflowUpdate struct {
	Type  string
	Owner struct {
		Username string `json:"display_name"`
	} `json:"owner"`
	CreatedAt int64  `json:"creation_date"`
	Body      string `json:"body"`
}

type GitHubUpdate struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	User  struct {
		Login string `json:"login"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
}

type Update struct {
	Title     string
	CreatedAt string
	Author    string
	Preview   string
}

func NewUpdate(title, createdAt, author, preview string) Update {
	return Update{
		Title:     title,
		CreatedAt: createdAt,
		Author:    author,
		Preview:   preview,
	}
}

func (u *Update) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "📌 New %s\n", u.Title)
	fmt.Fprintf(&b, "🕒 Date: %s\n", u.CreatedAt)
	fmt.Fprintf(&b, "👤 Author: %s\n", u.Author)
	fmt.Fprintf(&b, "🔗 View: %s\n\n", u.Preview)

	return b.String()
}
