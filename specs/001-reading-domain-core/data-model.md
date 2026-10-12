# Data Model: Доменное ядро учёта чтения

**Feature**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md) | **Date**: 2026-10-11 | **Revised**: 2026-10-12

Пакет `domain` (`internal/domain/`). Правила перевода из нотации `DESIGN.md` §4:

| `DESIGN.md` | Go |
|-------------|----|
| `long` | `int64` |
| `Date` | `time.Time` |
| поле со `*` (nullable) | указатель (`*string`, `*int`, `*int64`, `*time.Time`) |
| `text \| audio` | `BookType` со строковыми константами |

Комментарии из `DESIGN.md` («уникальный», «хэш», «секунды», связи) переносятся в код комментариями. Уникальность и ссылочная целостность — забота хранилища, а не этого этапа.

## Сущности только с данными

| Сущность | Поля | Файл |
|----------|------|------|
| User | `ID int64`, `Username string`, `Password string` (хэш) | `user.go` |
| Book | `ID int64`, `Type BookType`, `ISBN *string`, `Title string`, `Published *time.Time`, `SeriesID *int64`, `SeriesPos *int` | `book.go` |
| BookType | `BookTypeText = "text"`, `BookTypeAudio = "audio"` | `book.go` |
| TextBook | `ID int64` (= Book.ID), `EBook bool`, `TotalPages int` | `book.go` |
| AudioBook | `ID int64` (= Book.ID), `Narrator *string`, `Duration int64` (секунды) | `book.go` |
| Author | `ID int64`, `Name string` | `author.go` |
| BookAuthor | `AuthorID int64`, `BookID int64` | `author.go` |
| Series | `ID int64`, `Name string` | `series.go` |
| ReadingStatus | `ID int64`, `Name string` (справочник для хранилища) | `shelf.go` |
| AcquisitionType | `ID int64`, `Name string` | `ownership.go` |
| Currency | `ID int64`, `Name string` | `ownership.go` |
| BookOwnership | `ID int64`, `UserID int64`, `BookID int64`, `Acquired time.Time`, `AcquisitionTypeID int64`, `PriceCents *int64`, `CurrencyID *int64`, `ReturnDate *time.Time` | `ownership.go` |

## Сущности с поведением

### Status

`type Status int`, константы через `iota`, метод `String()`. Файл `shelf.go`.

| Константа | Значение | `String()` | Завершающий |
|-----------|----------|------------|-------------|
| `StatusWantToRead` | 0 (нулевое) | `"want_to_read"` | нет |
| `StatusReading` | 1 | `"reading"` | нет |
| `StatusCompleted` | 2 | `"completed"` | да |
| `StatusDropped` | 3 | `"dropped"` | да |

Любое другое целое значение недопустимо (`ErrInvalidStatus`).

### ShelfItem

Файл `shelf.go`. Расширяет `DESIGN.md` §4.9: `StatusID` заменён на `Status`, добавлены даты.

| Поле | Тип | Смысл |
|------|-----|-------|
| `ID` | `int64` | идентификатор |
| `UserID` | `int64` | many-to-1 к User |
| `BookID` | `int64` | many-to-1 к Book |
| `Status` | `Status` | текущий статус |
| `StartDate` | `*time.Time` | дата старта, `nil` — не задана |
| `FinishDate` | `*time.Time` | дата окончания, `nil` — не задана |
| `now` | `func() time.Time` | неэкспортируемое; источник текущего момента, `nil` → `time.Now` |

Методы: `ChangeStatus`, `SetStartDate`, `SetFinishDate`.

**Инварианты:**

1. `Status` — одно из четырёх значений.
2. Если заданы обе даты, `FinishDate` не раньше `StartDate`.
3. Даты, поставленные методами, не позже текущего момента на момент установки.

### ReadingLog

Файл `reading_log.go`. Поля как в `DESIGN.md` §4.10: `ID int64`, `UserID int64`, `BookID int64`, `Date time.Time`, `Pages *int`, `Seconds *int64`, `Comment *string`.
`Pages` и `Seconds` — **позиция** (страница, секунда), а не прирост.

Методы: `UpdateProgress` (текст), `UpdateAudioProgress` (аудио). Оба могут менять переданный `*ShelfItem`.

### Review

Файл `review.go`. Поля как в `DESIGN.md` §4.11: `ID int64`, `BookID int64`, `UserID int64`, `Rating int`, `Text string`. Метод: `SetRating`.

## Переходы статуса `ShelfItem`

```text
Ручная смена (ChangeStatus): любой статус → любой статус, даты не меняются.

Запись прогресса (UpdateProgress / UpdateAudioProgress), если item != nil:

  WantToRead ──(позиция < объём)──► Reading           StartDate = now, если nil
  WantToRead ──(позиция = объём)──► Completed         StartDate = now, если nil; FinishDate = now
  Reading    ──(позиция < объём)──► Reading           StartDate = now, если nil
  Reading    ──(позиция = объём)──► Completed         StartDate = now, если nil; FinishDate = now
  Completed / Dropped ──────────────► отказ ErrReadingFinished
```

## Правила валидации

Порядок проверок и ошибки для каждого метода — в [contracts/domain-api.md](./contracts/domain-api.md). При любой ошибке ни одно поле `ReadingLog`, `ShelfItem` или `Review` не меняется (FR-020).
