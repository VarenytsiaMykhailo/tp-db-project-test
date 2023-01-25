package domain

type PostFull struct {
	Profile *User   `json:"author,omitempty"`
	Forum   *Forum  `json:"forum,omitempty"`
	Post    Post    `json:"post"`
	Thread  *Thread `json:"thread,omitempty"`
}
