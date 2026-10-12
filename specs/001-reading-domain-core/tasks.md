# Tasks: Доменное ядро учёта чтения

**Feature**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md) | **Branch**: `001-reading-domain-core`
**Input**: design documents в `specs/001-reading-domain-core/`

**Tech stack** (из [plan.md](./plan.md)): Go 1.22+, только стандартная библиотека (`errors`, `time`; в тестах `testing`, `reflect`). Пакет `domain` в `internal/domain/`.

**Tests**: ОБЯЗАТЕЛЬНЫ. Конституция проекта требует TDD (принцип I) и прослеживаемость Red→Green (принцип II). На каждый из 6 методов — пара задач: сначала падающие тесты (коммит `test:`), затем реализация (коммит `feat:`).

## Условные обозначения

- `[P]` — задачу можно выполнять параллельно с другими `[P]` (разные файлы, нет зависимостей от незавершённых задач).
- `[US#]` — задача относится к пользовательской истории из [spec.md](./spec.md).
- Каждая задача указывает точный путь к файлу относительно корня репозитория.

**ВАЖНО про коммиты** (принцип II): в задачах с тестами коммит `test:` создаётся, когда тесты компилируются и ПАДАЮТ. В задачах реализации коммит `feat:` создаётся, когда те же тесты ПРОХОДЯТ. Запрещено схлопывать, amend-ить или переставлять эти коммиты.

**Приём «состояние не изменилось»** (FR-020): в негативных сценариях тест копирует получателя и переданный `*ShelfItem` до вызова, после ошибки сравнивает через `reflect.DeepEqual`, предварительно обнулив поле `now` в копиях (функции сравнимы только как `nil`, см. [research.md](./research.md) R6, R9).

---

## Phase 1: Setup

- [ ] T001 Создать `go.mod` в корне репозитория: `module github.com/PancakeNM/reading-tracker`, директива `go 1.22`, без секции `require` (ограничение этапа: только стандартная библиотека).
- [ ] T002 Создать каталог `internal/domain/` для пакета `domain`.

---

## Phase 2: Foundational (блокирует все истории)

**Цель**: описать все сущности `DESIGN.md` §4, тип `Status`, сигнальные ошибки и заглушки всех 6 методов. Заглушки ничего не делают и возвращают `nil` — это каркас для фазы Red. Вся фаза — ОДИН коммит `feat: add domain entities and method stubs` (поведения нет, тесты каркасу не нужны).

- [ ] T003 [P] Сущность `User` в `internal/domain/user.go`: `ID int64`, `Username string` (уникальный), `Password string` (хэш).
- [ ] T004 [P] Книги в `internal/domain/book.go`: `BookType` со строковыми константами `BookTypeText = "text"`, `BookTypeAudio = "audio"`; `Book` (`ID int64`, `Type BookType`, `ISBN *string`, `Title string`, `Published *time.Time`, `SeriesID *int64`, `SeriesPos *int`); `TextBook` (`ID int64`, `EBook bool`, `TotalPages int`); `AudioBook` (`ID int64`, `Narrator *string`, `Duration int64` — секунды).
- [ ] T005 [P] Авторство в `internal/domain/author.go`: `Author` (`ID int64`, `Name string`); `BookAuthor` (`AuthorID int64`, `BookID int64`).
- [ ] T006 [P] Серия в `internal/domain/series.go`: `Series` (`ID int64`, `Name string`).
- [ ] T007 [P] Владение в `internal/domain/ownership.go`: `AcquisitionType` (`ID int64`, `Name string`); `Currency` (`ID int64`, `Name string`); `BookOwnership` (`ID int64`, `UserID int64`, `BookID int64`, `Acquired time.Time`, `AcquisitionTypeID int64`, `PriceCents *int64`, `CurrencyID *int64`, `ReturnDate *time.Time`).
- [ ] T008 Полка и статус в `internal/domain/shelf.go`: справочник `ReadingStatus` (`ID int64`, `Name string`); тип `Status int` с константами через `iota` — `StatusWantToRead`=0, `StatusReading`=1, `StatusCompleted`=2, `StatusDropped`=3; метод `String()` (`"want_to_read"`, `"reading"`, `"completed"`, `"dropped"`); приватная проверка допустимости значения; структура `ShelfItem` (`ID int64`, `UserID int64`, `BookID int64`, `Status Status`, `StartDate *time.Time`, `FinishDate *time.Time`, неэкспортируемое `now func() time.Time`); приватный помощник `currentTime()` (возвращает `time.Now()`, если `now == nil`); заглушки `ChangeStatus(status Status) error`, `SetStartDate(d time.Time) error`, `SetFinishDate(d time.Time) error` — возвращают `nil`.
- [ ] T009 Запись о чтении в `internal/domain/reading_log.go`: структура `ReadingLog` (`ID int64`, `UserID int64`, `BookID int64`, `Date time.Time`, `Pages *int`, `Seconds *int64`, `Comment *string`; `Pages`/`Seconds` — позиция, а не прирост); заглушки `UpdateProgress(item *ShelfItem, page, totalPages int) error`, `UpdateAudioProgress(item *ShelfItem, seconds, totalSeconds int64) error` — возвращают `nil`.
- [ ] T010 Отзыв в `internal/domain/review.go`: структура `Review` (`ID int64`, `BookID int64`, `UserID int64`, `Rating int`, `Text string`); заглушка `SetRating(rating int) error` — возвращает `nil`.
- [ ] T011 Сигнальные ошибки в `internal/domain/errors.go` (через `errors.New`, см. [contracts/domain-api.md](./contracts/domain-api.md)): `ErrShelfItemMismatch`, `ErrReadingFinished`, `ErrNegativePage`, `ErrPageExceedsTotal`, `ErrNegativeSeconds`, `ErrSecondsExceedTotal`, `ErrInvalidStatus`, `ErrStartDateInFuture`, `ErrStartDateAfterFinish`, `ErrFinishDateNotAllowed`, `ErrFinishDateInFuture`, `ErrStartDateMissing`, `ErrFinishDateBeforeStart`, `ErrInvalidRating`.
- [ ] T012 Проверить сборку: `go build ./...`, `go vet ./...`, `gofmt -l .` (пусто). Создать коммит `feat: add domain entities and method stubs`.

**Checkpoint**: пакет компилируется, все структуры и заглушки на месте. Можно начинать любую историю.

---

## Phase 3: User Story 1 — Прогресс по текстовой книге (P1) 🎯 MVP

**Goal**: `ReadingLog.UpdateProgress` записывает позицию-страницу, проверяет соответствие полки, статус и границы, при достижении конца автоматически завершает чтение на полке.

**Independent Test**: создать `ReadingLog` и `ShelfItem` (статус «читаю», `now` фиксирован) с известным `totalPages`, передать разные страницы и проверить `*l.Pages`, статус полки и даты; негативные случаи — ошибку и неизменность состояния.

- [ ] T013 [US1] Табличные тесты `TestReadingLog_UpdateProgress` в `internal/domain/reading_log_test.go` (`package domain`, каждый случай через `t.Run`, в подтестах `t.Parallel()`, ошибки через `errors.Is`). Случаи из [contracts/domain-api.md](./contracts/domain-api.md): 150 из 300 (Reading, `FinishDate==nil`); страница 0; 300 из 300 (`StartDate` задан → Completed, `*FinishDate==now`); WantToRead 20 из 300 (→ Reading, `*StartDate==now`); Reading без `StartDate`, 300 из 300 (`StartDate==FinishDate==now`); `item==nil`, 300 из 300 (меняется только `Pages`); страница −1 → `ErrNegativePage`; 301 из 300 → `ErrPageExceedsTotal`; статус Completed и Dropped → `ErrReadingFinished`; `item` с чужим `UserID`/`BookID` → `ErrShelfItemMismatch`. В негативных случаях проверять неизменность `*l` и `*item` через `reflect.DeepEqual` (с обнулением `now`). Прогнать `go test ./internal/domain/ -run TestReadingLog_UpdateProgress` — тесты ПАДАЮТ. Коммит `test: add failing tests for ReadingLog.UpdateProgress`.
- [ ] T014 [US1] Реализовать `UpdateProgress` в `internal/domain/reading_log.go` по порядку проверок FR-018: (1) `item!=nil` и несовпадение `UserID`/`BookID` → `ErrShelfItemMismatch`; (2) `item!=nil` и завершающий статус → `ErrReadingFinished`; (3) `page<0` → `ErrNegativePage`; (4) `page>totalPages` → `ErrPageExceedsTotal`. Все проверки ДО изменений. При успехе: `l.Pages = &page`; если `item!=nil` — `WantToRead`→`Reading`, при `StartDate==nil` установить `*StartDate = item.currentTime()`, при `page==totalPages` → `Completed` и `*FinishDate = item.currentTime()`. Прогнать тесты — ЗЕЛЁНЫЕ. Коммит `feat: implement ReadingLog.UpdateProgress logic`.

**Checkpoint**: US1 реализована и независимо проверяема.

---

## Phase 4: User Story 2 — Прогресс по аудиокниге (P1)

**Goal**: `ReadingLog.UpdateAudioProgress` — то же, что US1, но с секундами и длительностью.

**Independent Test**: как US1, но `seconds`/`totalSeconds` и влияние на полку по условию `seconds==totalSeconds`.

- [ ] T015 [US2] Табличные тесты `TestReadingLog_UpdateAudioProgress` в `internal/domain/reading_log_test.go`. Случаи из контракта: 12 000 из 36 000; 0; 36 000 из 36 000 (Completed, `*FinishDate==now`); WantToRead 600 (→ Reading, `*StartDate==now`); `item==nil` (меняется только `Seconds`); −1 с → `ErrNegativeSeconds`; 36 001 из 36 000 → `ErrSecondsExceedTotal`; Completed/Dropped → `ErrReadingFinished`; чужой `UserID`/`BookID` → `ErrShelfItemMismatch`. Проверять, что `l.Pages` не меняется и неизменность состояния в негативных случаях. `go test -run TestReadingLog_UpdateAudioProgress` ПАДАЕТ. Коммит `test: add failing tests for ReadingLog.UpdateAudioProgress`.
- [ ] T016 [US2] Реализовать `UpdateAudioProgress` в `internal/domain/reading_log.go`: порядок как в US1, границы (3) `seconds<0` → `ErrNegativeSeconds`, (4) `seconds>totalSeconds` → `ErrSecondsExceedTotal`. При успехе `l.Seconds = &seconds`, влияние на `item` по условию `seconds==totalSeconds`. Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement ReadingLog.UpdateAudioProgress logic`.

**Checkpoint**: оба метода прогресса готовы.

---

## Phase 5: User Story 3 — Статус и даты на полке (P1)

**Goal**: `ShelfItem.ChangeStatus`, `SetStartDate`, `SetFinishDate` с проверками FR-015 – FR-017.

**Independent Test**: создать `ShelfItem` в разных статусах и с разными датами (`now` фиксирован), вызвать методы и проверить результат и неизменность при ошибке.

- [ ] T017 [US3] Табличные тесты `TestShelfItem_ChangeStatus` в `internal/domain/shelf_test.go`. Случаи: WantToRead→Reading, Completed→Reading, Dropped→WantToRead (успех, даты без изменений); смена на тот же статус (успех, без изменений); `Status(99)` и `Status(-1)` → `ErrInvalidStatus` (состояние не изменилось). `go test -run TestShelfItem_ChangeStatus` ПАДАЕТ. Коммит `test: add failing tests for ShelfItem.ChangeStatus`.
- [ ] T018 [US3] Реализовать `ChangeStatus` в `internal/domain/shelf.go`: недопустимое значение → `ErrInvalidStatus`; иначе `s.Status = status` (любой переход разрешён, даты не трогаются; смена на текущий статус — тоже успех). Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement ShelfItem.ChangeStatus logic`.
- [ ] T019 [US3] Табличные тесты `TestShelfItem_SetStartDate` в `internal/domain/shelf_test.go`. Случаи: дата в прошлом; дата `==now`; дата `==FinishDate` (успех); дата в 2030 году при `now`=2026-10-15 → `ErrStartDateInFuture`; дата позже заданной `FinishDate` → `ErrStartDateAfterFinish`. Негативные — неизменность. `go test -run TestShelfItem_SetStartDate` ПАДАЕТ. Коммит `test: add failing tests for ShelfItem.SetStartDate`.
- [ ] T020 [US3] Реализовать `SetStartDate` в `internal/domain/shelf.go` по порядку: (1) `d.After(currentTime())` → `ErrStartDateInFuture`; (2) `FinishDate!=nil && d.After(*FinishDate)` → `ErrStartDateAfterFinish`; иначе `s.StartDate = &d` (статус не важен и не меняется). Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement ShelfItem.SetStartDate logic`.
- [ ] T021 [US3] Табличные тесты `TestShelfItem_SetFinishDate` в `internal/domain/shelf_test.go`. Случаи: Completed/Dropped с датой между стартом и `now`; дата `==StartDate` (успех); статус Reading/WantToRead → `ErrFinishDateNotAllowed`; дата в 2030 году → `ErrFinishDateInFuture`; `StartDate==nil` → `ErrStartDateMissing`; дата раньше `StartDate` → `ErrFinishDateBeforeStart`. Негативные — неизменность. `go test -run TestShelfItem_SetFinishDate` ПАДАЕТ. Коммит `test: add failing tests for ShelfItem.SetFinishDate`.
- [ ] T022 [US3] Реализовать `SetFinishDate` в `internal/domain/shelf.go` по порядку FR-017: (1) статус не Completed и не Dropped → `ErrFinishDateNotAllowed`; (2) `d.After(currentTime())` → `ErrFinishDateInFuture`; (3) `StartDate==nil` → `ErrStartDateMissing`; (4) `d.Before(*StartDate)` → `ErrFinishDateBeforeStart`; иначе `s.FinishDate = &d`. Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement ShelfItem.SetFinishDate logic`.

**Checkpoint**: полка полностью управляема.

---

## Phase 6: User Story 4 — Оценка в отзыве (P2)

**Goal**: `Review.SetRating` принимает только 1–5, иначе `ErrInvalidRating` (инвариант `DESIGN.md` §7.5).

**Independent Test**: вызвать с граничными и недопустимыми значениями, проверить результат и неизменность при ошибке.

- [ ] T023 [US4] Табличные тесты `TestReview_SetRating` в `internal/domain/review_test.go`. Случаи: 1, 3, 5 (успех, `r.Rating` установлен); 0, 6, −1 → `ErrInvalidRating` (оценка не изменилась). `go test -run TestReview_SetRating` ПАДАЕТ. Коммит `test: add failing tests for Review.SetRating`.
- [ ] T024 [US4] Реализовать `SetRating` в `internal/domain/review.go`: `rating<1 || rating>5` → `ErrInvalidRating`; иначе `r.Rating = rating`. Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement Review.SetRating logic`.

**Checkpoint**: все 6 методов реализованы по TDD.

---

## Phase 7: User Story 5 — Модель предметной области (P3)

**Goal**: убедиться, что модель ядра соответствует `DESIGN.md` §4 (кроме задокументированных изменений `ShelfItem`).

**Independent Test**: ручная сверка + сборка без внешних зависимостей.

- [ ] T025 [US5] Сверить сущности в `internal/domain/*.go` с таблицами [data-model.md](./data-model.md) и `DESIGN.md` §4: 14 сущностей, поля и обязательность (`*` → указатель) совпадают; единственное допустимое отличие — `ShelfItem` (`Status` вместо `StatusID`, добавлены `StartDate`, `FinishDate`). Зафиксировать расхождения, если найдены (SC-006).

---

## Phase 8: Polish & Cross-Cutting

- [ ] T026 [P] `gofmt -l .` (пусто) и `go vet ./...` (чисто) по всему модулю.
- [ ] T027 [P] Детерминированность и гонки: `go test ./internal/domain/ -count=100` и `go test ./internal/domain/ -race` зелёные (SC-003); прогон укладывается в 5 с (SC-004).
- [ ] T028 Проверить ограничение «только стандартная библиотека»: `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/domain/` выводит только `github.com/PancakeNM/reading-tracker/internal/domain`; в `go.mod` нет `require`, файла `go.sum` нет (FR-003).
- [ ] T029 Проверить прослеживаемость Red→Green по `git log --oneline -- internal/domain/`: для каждого из 6 методов коммит `test:` предшествует `feat:` (SC-005, принцип II).

---

## Dependencies & Execution Order

- **Setup (Phase 1)** → **Foundational (Phase 2)** → истории.
- **Phase 2 блокирует всё**: структуры, тип `Status`, ошибки и заглушки нужны всем методам.
- После Phase 2 истории технически независимы (каждый метод в своём методе, тесты в своём файле или своей функции):
  - US1 (T013–T014) и US2 (T015–T016) правят `reading_log.go` / `reading_log_test.go` — вести последовательно, один файл.
  - US3 (T017–T022) правит `shelf.go` / `shelf_test.go` — не конфликтует с US1/US2, может идти параллельно другому разработчику.
  - US4 (T023–T024) правит `review.go` / `review_test.go` — независима.
- **Внутри каждого метода**: задача `test:` строго раньше задачи `feat:` (принцип II, нарушать нельзя).
- US5 (T025) и Polish (Phase 8) — после реализации всех методов.

## Parallel Opportunities

- **Phase 2**: T003–T007 помечены `[P]` — разные файлы, можно писать одновременно (но коммит общий, T012).
- **Между историями**: US3 и US4 можно вести параллельно US1/US2 (разные файлы). US1 и US2 — последовательно (общий файл `reading_log.go`).
- **Phase 8**: T026 и T027 параллельны.

## Implementation Strategy

- **MVP = User Story 1** (T001–T014): каркас + запись прогресса по текстовой книге с автозавершением. Уже даёт проверяемую бизнес-ценность.
- **Инкременты**: US2 (аудио) → US3 (полка) → US4 (оценка). Каждый добавляет один-два метода, не ломая предыдущие.
- **Правило коммитов на каждом методе**: `test:` (красный) → `feat:` (зелёный). Это и есть доказательство TDD для ревью.

## Notes

- Шаблон `tasks-template` в `.specify/templates/` отсутствует — использована стандартная структура Spec Kit.
- Go на машине пока не установлен; установка — предусловие из [quickstart.md](./quickstart.md) (`winget install GoLang.Go`), нужна начиная с T012.
- Порядок историй здесь следует приоритетам [spec.md](./spec.md) (US1 первой как MVP). В [research.md](./research.md) R11 приведён альтернативный порядок коммитов (сначала `ShelfItem`); оба допустимы, ключевой инвариант — `test:` раньше `feat:` на каждом методе.
