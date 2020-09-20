package model

type User struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Token   string `json:"token"`
}

type UserLog struct {
	Time string `json:"time"`
	User User   `json:"user"`
}
