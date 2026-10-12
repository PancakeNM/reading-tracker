package domain

import "errors"

var (
	ErrShelfItemMismatch     = errors.New("запись полки другого пользователя или другой книги")
	ErrReadingFinished       = errors.New("статус полки Completed или Dropped")
	ErrNegativePage          = errors.New("page < 0")
	ErrPageExceedsTotal      = errors.New("page > totalPages")
	ErrNegativeSeconds       = errors.New("seconds < 0")
	ErrSecondsExceedTotal    = errors.New("seconds > totalSeconds")
	ErrInvalidStatus         = errors.New("статус вне четырёх допустимых")
	ErrStartDateInFuture     = errors.New("дата старта позже now")
	ErrStartDateAfterFinish  = errors.New("дата старта позже заданной FinishDate")
	ErrFinishDateNotAllowed  = errors.New("дата окончания при статусе не Completed и не Dropped")
	ErrFinishDateInFuture    = errors.New("дата окончания позже now")
	ErrStartDateMissing      = errors.New("дата окончания без StartDate")
	ErrFinishDateBeforeStart = errors.New("дата окончания раньше StartDate")
	ErrInvalidRating         = errors.New("оценка вне 1..5")
)
