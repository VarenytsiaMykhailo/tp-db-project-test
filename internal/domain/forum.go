package domain

type Forum struct {
	Title           string `json:"title"`
	Slug            string `json:"slug"`
	ProfileId       int32  `json:"-"`
	ProfileNickname string `json:"user"`
	Posts           int64  `json:"posts"`
	Threads         int32  `json:"threads"`
}
