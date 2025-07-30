package models

const (
	Client = iota
	Admin
)

type UserRole int

type User struct {
	Id           int64
	Login        string
	PasswordHash string
	Role         UserRole
}
