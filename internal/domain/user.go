package domain

type User struct {
	Id       int32  `json:"-"`
	Nickname string `json:"nickname"`
	Fullname string `json:"fullname"`
	About    string `json:"about"`
	Email    string `json:"email"`
}
