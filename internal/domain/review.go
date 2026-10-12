package domain

// Review — отзыв на книгу. Один на пару User и Book. С ShelfItem не связан.
type Review struct {
	ID     int64
	BookID int64 // many-to-1 к Book
	UserID int64 // many-to-1 к User
	Rating int   // от 1 до 5
	Text   string
}

// SetRating задаёт оценку. Заглушка каркаса.
func (r *Review) SetRating(rating int) error {
	return errNotImplemented
}
