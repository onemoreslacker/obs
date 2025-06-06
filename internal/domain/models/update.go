package models

import (
	"fmt"
	"strings"
)

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
