package domain

type Status struct {
	Forum  uint32 `json:"forum"`
	Post   uint64 `json:"post"`
	Thread uint32 `json:"thread"`
	User   uint32 `json:"user"`
}
