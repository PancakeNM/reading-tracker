package domain

import "time"

// ReadingStatus — справочник статуса записи на полке. 1-to-many к ShelfItem.
type ReadingStatus struct {
	ID   int64
	Name string
}

// Status — статус записи на полке.
type Status int

const (
	StatusWantToRead Status = iota // 0
	StatusReading                  // 1
	StatusCompleted                // 2
	StatusDropped                  // 3
)

// String возвращает имя статуса: want_to_read, reading, completed, dropped.
func (s Status) String() string {
	switch s {
	case StatusWantToRead:
		return "want_to_read"
	case StatusReading:
		return "reading"
	case StatusCompleted:
		return "completed"
	case StatusDropped:
		return "dropped"
	default:
		return ""
	}
}

func (s Status) valid() bool {
	switch s {
	case StatusWantToRead, StatusReading, StatusCompleted, StatusDropped:
		return true
	default:
		return false
	}
}

// ShelfItem — книга на полке пользователя. Одна пара User и Book — одна запись.
type ShelfItem struct {
	ID         int64
	UserID     int64 // many-to-1 к User
	BookID     int64 // many-to-1 к Book
	Status     Status
	StartDate  *time.Time
	FinishDate *time.Time
	now        func() time.Time
}

func (s *ShelfItem) currentTime() time.Time {
	if s.now == nil {
		return time.Now()
	}
	return s.now()
}

// ChangeStatus меняет статус записи. Заглушка каркаса.
func (s *ShelfItem) ChangeStatus(status Status) error {
	return errNotImplemented
}

// SetStartDate задаёт дату старта. Заглушка каркаса.
func (s *ShelfItem) SetStartDate(d time.Time) error {
	return errNotImplemented
}

// SetFinishDate задаёт дату окончания. Заглушка каркаса.
func (s *ShelfItem) SetFinishDate(d time.Time) error {
	return errNotImplemented
}

// ClearDates очищает даты записи. Заглушка каркаса.
func (s *ShelfItem) ClearDates() {}
