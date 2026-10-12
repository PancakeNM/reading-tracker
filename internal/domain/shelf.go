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

// ChangeStatus меняет статус записи. Любой переход допустим, даты не меняются.
func (s *ShelfItem) ChangeStatus(status Status) error {
	if !status.valid() {
		return ErrInvalidStatus
	}
	s.Status = status
	return nil
}

// SetStartDate задаёт дату старта. Статус не меняется.
func (s *ShelfItem) SetStartDate(d time.Time) error {
	if d.After(s.currentTime()) {
		return ErrStartDateInFuture
	}
	if s.FinishDate != nil && d.After(*s.FinishDate) {
		return ErrStartDateAfterFinish
	}
	s.StartDate = &d
	return nil
}

// SetFinishDate задаёт дату окончания. Заглушка каркаса.
func (s *ShelfItem) SetFinishDate(d time.Time) error {
	return errNotImplemented
}

// ClearDates очищает даты записи. Заглушка каркаса.
func (s *ShelfItem) ClearDates() {}
