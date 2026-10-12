package domain

import "time"

// BookType — дискриминатор подтипа книги.
type BookType string

const (
	BookTypeText  BookType = "text"
	BookTypeAudio BookType = "audio"
)

// Book — супертип текстовой и аудиокниги.
type Book struct {
	ID        int64
	Type      BookType
	ISBN      *string
	Title     string
	Published *time.Time
	SeriesID  *int64 // many-to-1 к Series
	SeriesPos *int
}

// TextBook — подтип Book, связь is-a.
type TextBook struct {
	ID         int64 // совпадает с Book.ID
	EBook      bool
	TotalPages int
}

// AudioBook — подтип Book, связь is-a.
type AudioBook struct {
	ID       int64 // совпадает с Book.ID
	Narrator *string
	Duration int64 // секунды
}
