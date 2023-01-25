package domain

import "time"

type Thread struct {
	Id              int32     `json:"id"`
	Title           string    `json:"title"`
	ProfileNickname string    `json:"author"`
	ProfileId       int32     `json:""`
	ForumSlug       string    `json:"forum"`
	Message         string    `json:"message"`
	Votes           int32     `json:"votes"`
	Slug            string    `json:"slug"`
	Created         time.Time `json:"created"`
}
