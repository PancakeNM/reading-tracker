# Contract: публичный API пакета `domain`

**Package**: `github.com/PancakeNM/reading-tracker/internal/domain`
**Feature**: [spec.md](../spec.md) | **Data model**: [data-model.md](../data-model.md)

HTTP на этом этапе нет. Контракт ядра — экспортируемый Go API пакета, который используют будущие слои того же модуля. По принципу V конституции контракт покрывается тестами до реализации.

`now` в таблицах ниже — текущий момент на входе в метод, взятый у `ShelfItem` (см. [research.md](../research.md) R6).

## Ошибки

```go
var (
	ErrShelfItemMismatch     // запись полки другого пользователя или другой книги
	ErrReadingFinished       // статус полки Completed или Dropped
	ErrNegativePage          // page < 0
	ErrPageExceedsTotal      // page > totalPages
	ErrNegativeSeconds       // seconds < 0
	ErrSecondsExceedTotal    // seconds > totalSeconds
	ErrInvalidStatus         // статус вне четырёх допустимых
	ErrStartDateInFuture     // дата старта позже now
	ErrStartDateAfterFinish  // дата старта позже заданной FinishDate
	ErrFinishDateNotAllowed  // дата окончания при статусе не Completed / Dropped
	ErrFinishDateInFuture    // дата окончания позже now
	ErrStartDateMissing      // дата окончания без StartDate
	ErrFinishDateBeforeStart // дата окончания раньше StartDate
	ErrInvalidRating         // оценка вне 1..5
)
```

Все ошибки — сигнальные значения (`errors.New`), распознаются через `errors.Is`.

**Правило присваивания** (см. [research.md](../research.md) R9): там, где ниже написано «поле = значение» для поля-указателя, метод присваивает **новый указатель** (`item.FinishDate = &t`) и никогда не пишет через существующий (`*item.FinishDate = t`). Запись через `nil` вызовет панику, а запись через уже заданный указатель изменит и снимок состояния, который сделал тест.

## `func (s *ShelfItem) ChangeStatus(status Status) error`

| # | Условие | Ошибка |
|---|---------|--------|
| 1 | `status` не одно из четырёх значений | `ErrInvalidStatus` |

**Успех**: `s.Status == status`. Даты не меняются. Если `status` совпадает с текущим, ничего не меняется.

## `func (s *ShelfItem) SetStartDate(d time.Time) error`

| # | Условие | Ошибка |
|---|---------|--------|
| 1 | `d.After(now)` | `ErrStartDateInFuture` |
| 2 | `s.FinishDate != nil && d.After(*s.FinishDate)` | `ErrStartDateAfterFinish` |

**Успех**: `s.StartDate = &d`. Статус не важен и не меняется.

## `func (s *ShelfItem) SetFinishDate(d time.Time) error`

| # | Условие | Ошибка |
|---|---------|--------|
| 1 | `s.Status` не `StatusCompleted` и не `StatusDropped` | `ErrFinishDateNotAllowed` |
| 2 | `d.After(now)` | `ErrFinishDateInFuture` |
| 3 | `s.StartDate == nil` | `ErrStartDateMissing` |
| 4 | `d.Before(*s.StartDate)` | `ErrFinishDateBeforeStart` |

**Успех**: `s.FinishDate = &d`. Статус не меняется.

## `func (s *ShelfItem) ClearDates()`

Отказов нет, ошибку метод не возвращает. Работает при любом статусе.

**Результат**: `s.StartDate = nil`, `s.FinishDate = nil`. Статус не меняется.
**Назначение**: начать книгу заново после отката статуса. Старая дата окончания иначе не даёт поставить новую дату старта (`ErrStartDateAfterFinish`).

## `func (l *ReadingLog) UpdateProgress(item *ShelfItem, page, totalPages int) error`

| # | Условие | Ошибка |
|---|---------|--------|
| 1 | `item != nil` и (`item.UserID != l.UserID` или `item.BookID != l.BookID`) | `ErrShelfItemMismatch` |
| 2 | `item != nil` и `item.Status` — `StatusCompleted` или `StatusDropped` | `ErrReadingFinished` |
| 3 | `page < 0` | `ErrNegativePage` |
| 4 | `page > totalPages` | `ErrPageExceedsTotal` |

**Успех:**

- `l.Pages = &page`; `l.Seconds`, `l.Date`, `l.Comment` не меняются.
- Если `item == nil`, больше ничего не происходит.
- Если `item != nil` (`t := now`, одно значение на весь вызов):
  - `item.Status == StatusWantToRead` → `item.Status = StatusReading`;
  - `item.StartDate == nil` → `item.StartDate = &t`;
  - `page == totalPages` → `item.Status = StatusCompleted`, `item.FinishDate = &t` (новый указатель, даже если старая `FinishDate` была задана).

## `func (l *ReadingLog) UpdateAudioProgress(item *ShelfItem, seconds, totalSeconds int64) error`

То же, что `UpdateProgress`, с заменой:

| # | Условие | Ошибка |
|---|---------|--------|
| 1–2 | как в `UpdateProgress` | `ErrShelfItemMismatch`, `ErrReadingFinished` |
| 3 | `seconds < 0` | `ErrNegativeSeconds` |
| 4 | `seconds > totalSeconds` | `ErrSecondsExceedTotal` |

**Успех**: `l.Seconds = &seconds`, `l.Pages` не меняется. Влияние на `item` такое же, но с условием `seconds == totalSeconds`.

## `func (r *Review) SetRating(rating int) error`

| # | Условие | Ошибка |
|---|---------|--------|
| 1 | `rating < 1 \|\| rating > 5` | `ErrInvalidRating` |

**Успех**: `r.Rating == rating`.

## Общее постусловие при ошибке

Все поля получателя и переданного `*ShelfItem` остаются без изменений.

## Обязательные тестовые случаи (минимум)

| Метод | Случай | Ожидание |
|-------|--------|----------|
| ChangeStatus | WantToRead → Reading; Completed → Reading; Dropped → WantToRead | `nil`, даты без изменений |
| ChangeStatus | тот же статус | `nil`, без изменений |
| ChangeStatus | `Status(99)`, `Status(-1)` | `ErrInvalidStatus` |
| SetStartDate | дата в прошлом; дата = now; дата = FinishDate | `nil` |
| SetStartDate | дата в 2030 году (now = 2026-10-15) | `ErrStartDateInFuture` |
| SetStartDate | дата позже FinishDate | `ErrStartDateAfterFinish` |
| SetStartDate | **порядок**: дата в 2030 году и FinishDate задана раньше неё | `ErrStartDateInFuture` |
| SetFinishDate | Completed / Dropped, дата между стартом и now; дата = StartDate | `nil` |
| SetFinishDate | статус Reading / WantToRead | `ErrFinishDateNotAllowed` |
| SetFinishDate | дата в 2030 году | `ErrFinishDateInFuture` |
| SetFinishDate | StartDate == nil | `ErrStartDateMissing` |
| SetFinishDate | дата раньше StartDate | `ErrFinishDateBeforeStart` |
| SetFinishDate | **порядок**: статус Reading и дата в 2030 году | `ErrFinishDateNotAllowed` |
| SetFinishDate | **порядок**: Completed без StartDate и дата в 2030 году | `ErrFinishDateInFuture` |
| ClearDates | Reading после отката, обе даты заданы | обе даты `nil`, статус Reading |
| ClearDates | Completed только со StartDate; Completed только с FinishDate | обе даты `nil`, статус Completed |
| UpdateProgress | Reading, 150 из 300; страница 0 | `nil`, статус Reading |
| UpdateProgress | 300 из 300 | `nil`, Completed, FinishDate = now |
| UpdateProgress | WantToRead, 20 из 300 | `nil`, Reading, StartDate = now |
| UpdateProgress | Reading без StartDate, 300 из 300 | `nil`, StartDate = FinishDate = now |
| UpdateProgress | `item == nil`, 300 из 300 | `nil`, только `Pages` |
| UpdateProgress | страница −1 | `ErrNegativePage` |
| UpdateProgress | страница 301 из 300 | `ErrPageExceedsTotal` |
| UpdateProgress | Completed / Dropped | `ErrReadingFinished` |
| UpdateProgress | чужой UserID / BookID | `ErrShelfItemMismatch` |
| UpdateProgress | Reading со старой FinishDate (после отката), 300 из 300 | `nil`, FinishDate = now (новый указатель) |
| UpdateProgress | **порядок**: Completed и страница −1 | `ErrReadingFinished` |
| UpdateProgress | **порядок**: чужая полка в статусе Completed | `ErrShelfItemMismatch` |
| UpdateAudioProgress | 12 000 из 36 000; 0 | `nil` |
| UpdateAudioProgress | 36 000 из 36 000 | `nil`, Completed, FinishDate = now |
| UpdateAudioProgress | WantToRead, 600 | `nil`, Reading, StartDate = now |
| UpdateAudioProgress | Reading без StartDate, 36 000 из 36 000 | `nil`, StartDate = FinishDate = now |
| UpdateAudioProgress | `item == nil` | `nil`, только `Seconds` |
| UpdateAudioProgress | −1 с | `ErrNegativeSeconds` |
| UpdateAudioProgress | 36 001 из 36 000 | `ErrSecondsExceedTotal` |
| UpdateAudioProgress | Completed / Dropped | `ErrReadingFinished` |
| UpdateAudioProgress | чужой UserID / BookID | `ErrShelfItemMismatch` |
| UpdateAudioProgress | **порядок**: Dropped и −1 с | `ErrReadingFinished` |
| UpdateAudioProgress | **порядок**: чужая полка в статусе Dropped | `ErrShelfItemMismatch` |
| SetRating | 1, 3, 5 | `nil` |
| SetRating | 0, 6, −1 | `ErrInvalidRating`, оценка без изменений |

Во всех случаях с ошибкой дополнительно проверяется, что получатель и `item` не изменились.
