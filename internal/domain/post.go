package domain

import "time"

type Post struct {
	Id              uint64    `json:"id"`
	ProfileId       uint32    `json:"-"`
	ProfileNickname string    `json:"author"`
	Created         time.Time `json:"created"`
	ForumSlug       string    `json:"forum"`
	IsEdited        bool      `json:"isEdited"`
	Message         string    `json:"message"`
	ParentPost      uint64    `json:"parent,omitempty"`
	ThreadId        int32     `json:"thread"`
}
