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

// UpdateProgress записывает позицию-страницу и при достижении конца завершает чтение на полке.
func (l *ReadingLog) UpdateProgress(item *ShelfItem, page, totalPages int) error {
	if item != nil && (item.UserID != l.UserID || item.BookID != l.BookID) {
		return ErrShelfItemMismatch
	}
	if item != nil && (item.Status == StatusCompleted || item.Status == StatusDropped) {
		return ErrReadingFinished
	}
	if page < 0 {
		return ErrNegativePage
	}
	if page > totalPages {
		return ErrPageExceedsTotal
	}

	l.Pages = &page
	if item == nil {
		return nil
	}

	t := item.currentTime()
	if item.Status == StatusWantToRead {
		item.Status = StatusReading
	}
	if item.StartDate == nil {
		item.StartDate = &t
	}
	if page == totalPages {
		item.Status = StatusCompleted
		item.FinishDate = &t
	}
	return nil
}

// UpdateAudioProgress записывает позицию в секундах и при достижении конца завершает чтение на полке.
func (l *ReadingLog) UpdateAudioProgress(item *ShelfItem, seconds, totalSeconds int64) error {
	if item != nil && (item.UserID != l.UserID || item.BookID != l.BookID) {
		return ErrShelfItemMismatch
	}
	if item != nil && (item.Status == StatusCompleted || item.Status == StatusDropped) {
		return ErrReadingFinished
	}
	if seconds < 0 {
		return ErrNegativeSeconds
	}
	if seconds > totalSeconds {
		return ErrSecondsExceedTotal
	}

	l.Seconds = &seconds
	if item == nil {
		return nil
	}

	t := item.currentTime()
	if item.Status == StatusWantToRead {
		item.Status = StatusReading
	}
	if item.StartDate == nil {
		item.StartDate = &t
	}
	if seconds == totalSeconds {
		item.Status = StatusCompleted
		item.FinishDate = &t
	}
	return nil
}
