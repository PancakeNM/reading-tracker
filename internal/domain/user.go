package domain

// User — пользователь приложения.
type User struct {
	ID       int64
	Username string // уникальный
	Password string // хэш
}
