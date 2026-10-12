package domain

import "time"

// ReadingLog — запись о чтении. Не зависит от ShelfItem.
// Pages и Seconds — позиция (страница, секунда), а не прирост за подход.
type ReadingLog struct {
	ID      int64
	UserID  int64 // many-to-1 к User
	BookID  int64 // many-to-1 к Book
	Date    time.Time
	Pages   *int
	Seconds *int64 // секунды, позиция
	Comment *string
}

// UpdateProgress записывает позицию по текстовой книге. Заглушка каркаса.
func (l *ReadingLog) UpdateProgress(item *ShelfItem, page, totalPages int) error {
	return errNotImplemented
}

// UpdateAudioProgress записывает позицию по аудиокниге. Заглушка каркаса.
func (l *ReadingLog) UpdateAudioProgress(item *ShelfItem, seconds, totalSeconds int64) error {
	return errNotImplemented
}
