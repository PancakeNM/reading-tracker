package domain

// Author — автор книги.
type Author struct {
	ID   int64
	Name string
}

// BookAuthor — связь many-to-many между Book и Author.
type BookAuthor struct {
	AuthorID int64
	BookID   int64
}
