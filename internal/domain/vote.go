package domain

type Vote struct {
	ProfileId       uint32 `json:"-"`
	ProfileNickname string `json:"nickname"`
	ThreadId        uint32 `json:"-"`
	Voice           int8   `json:"voice"`
}
