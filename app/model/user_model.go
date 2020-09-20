package model

type User struct {
	Id      string `json:"sub"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type UserLog struct {
	Time string `json:"time"`
	User User   `json:"user"`
}
