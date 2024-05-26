package models

type User struct {
	ID           int64
	Username     string
	Email        string
	Role         string
	Activated    bool
	PasswordHash Password
}

type Password struct {
	PlainText *string
	Hash      []byte
}

type App struct {
	ID     int
	Name   string
	Secret string
}
