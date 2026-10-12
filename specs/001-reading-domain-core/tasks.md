# Tasks: Доменное ядро учёта чтения

**Feature**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md) | **Branch**: `001-reading-domain-core`
**Input**: design documents в `specs/001-reading-domain-core/`

**Tech stack** (из [plan.md](./plan.md)): Go 1.22+, только стандартная библиотека (`errors`, `time`; в тестах `testing`, `reflect`). Пакет `domain` в `internal/domain/`.

**Tests**: ОБЯЗАТЕЛЬНЫ. Конституция проекта требует TDD (принцип I) и прослеживаемость Red→Green (принцип II). На каждый из 7 методов — пара задач: сначала падающие тесты (коммит `test:`), затем реализация (коммит `feat:`).

## Условные обозначения

- `[P]` — задачу можно выполнять параллельно с другими `[P]` (разные файлы, нет зависимостей от незавершённых задач).
- `[US#]` — задача относится к пользовательской истории из [spec.md](./spec.md).
- Каждая задача указывает точный путь к файлу относительно корня репозитория.

**ВАЖНО про коммиты** (принцип II): в задачах с тестами коммит `test:` создаётся, когда тесты компилируются и ПАДАЮТ. В задачах реализации коммит `feat:` создаётся, когда те же тесты ПРОХОДЯТ. Запрещено схлопывать, amend-ить или переставлять эти коммиты.

**Приём «состояние не изменилось»** (FR-020): в негативных сценариях тест копирует получателя и переданный `*ShelfItem` до вызова, после ошибки сравнивает через `reflect.DeepEqual`, предварительно обнулив поле `now` в копиях (функции сравнимы только как `nil`, см. [research.md](./research.md) R6, R9).

**Правило присваивания указателей** (R9): методы присваивают полю-указателю **новый** указатель (`item.FinishDate = &t`) и никогда не пишут через существующий (`*item.FinishDate = t`): через `nil` это паника, через заданный указатель — порча снимка в тесте.

**Порядок проверок** (FR-017, FR-018): в каждой таблице есть случаи «нарушено сразу несколько правил», помеченные ниже как **порядок**. Они доказывают, что возвращается первая по порядку ошибка.

---

## Phase 1: Setup

- [X] T001 Создать `go.mod` в корне репозитория: `module github.com/PancakeNM/reading-tracker`, директива `go 1.22`, без секции `require` (ограничение этапа: только стандартная библиотека).
- [X] T002 Создать каталог `internal/domain/` для пакета `domain`.

---

## Phase 2: Foundational (блокирует все истории)

**Цель**: описать все сущности `DESIGN.md` §4, тип `Status`, сигнальные ошибки и заглушки всех 7 методов. Заглушки без логики: методы с ошибкой возвращают неэкспортируемую `errNotImplemented`, `ClearDates` ничего не делает. Так на заглушке падает **каждый** подтест (R11). Вся фаза — ОДИН коммит `chore: scaffold domain entities and method stubs` (реализации поведения нет, поэтому `chore:`, а не `feat:`).

- [X] T003 [P] Сущность `User` в `internal/domain/user.go`: `ID int64`, `Username string` (уникальный), `Password string` (хэш).
- [X] T004 [P] Книги в `internal/domain/book.go`: `BookType` со строковыми константами `BookTypeText = "text"`, `BookTypeAudio = "audio"`; `Book` (`ID int64`, `Type BookType`, `ISBN *string`, `Title string`, `Published *time.Time`, `SeriesID *int64`, `SeriesPos *int`); `TextBook` (`ID int64`, `EBook bool`, `TotalPages int`); `AudioBook` (`ID int64`, `Narrator *string`, `Duration int64` — секунды).
- [X] T005 [P] Авторство в `internal/domain/author.go`: `Author` (`ID int64`, `Name string`); `BookAuthor` (`AuthorID int64`, `BookID int64`).
- [X] T006 [P] Серия в `internal/domain/series.go`: `Series` (`ID int64`, `Name string`).
- [X] T007 [P] Владение в `internal/domain/ownership.go`: `AcquisitionType` (`ID int64`, `Name string`); `Currency` (`ID int64`, `Name string`); `BookOwnership` (`ID int64`, `UserID int64`, `BookID int64`, `Acquired time.Time`, `AcquisitionTypeID int64`, `PriceCents *int64`, `CurrencyID *int64`, `ReturnDate *time.Time`).
- [X] T008 Полка и статус в `internal/domain/shelf.go`: справочник `ReadingStatus` (`ID int64`, `Name string`); тип `Status int` с константами через `iota` — `StatusWantToRead`=0, `StatusReading`=1, `StatusCompleted`=2, `StatusDropped`=3; метод `String()` (`"want_to_read"`, `"reading"`, `"completed"`, `"dropped"`); приватная проверка допустимости значения; структура `ShelfItem` (`ID int64`, `UserID int64`, `BookID int64`, `Status Status`, `StartDate *time.Time`, `FinishDate *time.Time`, неэкспортируемое `now func() time.Time`); приватный помощник `currentTime()` (возвращает `time.Now()`, если `now == nil`); заглушки `ChangeStatus(status Status) error`, `SetStartDate(d time.Time) error`, `SetFinishDate(d time.Time) error` — возвращают `errNotImplemented`; заглушка `ClearDates()` — пустое тело.
- [X] T009 Запись о чтении в `internal/domain/reading_log.go`: структура `ReadingLog` (`ID int64`, `UserID int64`, `BookID int64`, `Date time.Time`, `Pages *int`, `Seconds *int64`, `Comment *string`; `Pages`/`Seconds` — позиция, а не прирост); заглушки `UpdateProgress(item *ShelfItem, page, totalPages int) error`, `UpdateAudioProgress(item *ShelfItem, seconds, totalSeconds int64) error` — возвращают `errNotImplemented`.
- [X] T010 Отзыв в `internal/domain/review.go`: структура `Review` (`ID int64`, `BookID int64`, `UserID int64`, `Rating int`, `Text string`); заглушка `SetRating(rating int) error` — возвращает `errNotImplemented`.
- [X] T011 Ошибки в `internal/domain/errors.go` (через `errors.New`, см. [contracts/domain-api.md](./contracts/domain-api.md)): экспортируемые `ErrShelfItemMismatch`, `ErrReadingFinished`, `ErrNegativePage`, `ErrPageExceedsTotal`, `ErrNegativeSeconds`, `ErrSecondsExceedTotal`, `ErrInvalidStatus`, `ErrStartDateInFuture`, `ErrStartDateAfterFinish`, `ErrFinishDateNotAllowed`, `ErrFinishDateInFuture`, `ErrStartDateMissing`, `ErrFinishDateBeforeStart`, `ErrInvalidRating`; временная неэкспортируемая `errNotImplemented` (удаляется в T028).
- [X] T012 Проверить сборку: `go build ./...`, `go vet ./...`, `gofmt -l .` (пусто). Создать коммит `chore: scaffold domain entities and method stubs`.

**Checkpoint**: пакет компилируется, все структуры и заглушки на месте. Можно начинать любую историю.

---

## Phase 3: User Story 1 — Прогресс по текстовой книге (P1) 🎯 MVP

**Goal**: `ReadingLog.UpdateProgress` записывает позицию-страницу, проверяет соответствие полки, статус и границы, при достижении конца автоматически завершает чтение на полке.

**Independent Test**: создать `ReadingLog` и `ShelfItem` (статус «читаю», `now` фиксирован) с известным `totalPages`, передать разные страницы и проверить `*l.Pages`, статус полки и даты; негативные случаи — ошибку и неизменность состояния.

- [X] T013 [US1] Табличные тесты `TestReadingLog_UpdateProgress` в `internal/domain/reading_log_test.go` (`package domain`, каждый случай через `t.Run`, в подтестах `t.Parallel()`, ошибки через `errors.Is`). Случаи из [contracts/domain-api.md](./contracts/domain-api.md): 150 из 300 (Reading, `FinishDate==nil`); страница 0; 300 из 300 (`StartDate` задан → Completed, `*FinishDate==now`); WantToRead 20 из 300 (→ Reading, `*StartDate==now`); Reading без `StartDate`, 300 из 300 (`*StartDate==*FinishDate==now`); Reading со старой `FinishDate` (после отката), 300 из 300 → `*FinishDate==now`, а снимок полки, сделанный до вызова, по-прежнему хранит старую дату (проверка правила указателей); `item==nil`, 300 из 300 (меняется только `Pages`); страница −1 → `ErrNegativePage`; 301 из 300 → `ErrPageExceedsTotal`; статус Completed и Dropped → `ErrReadingFinished`; `item` с чужим `UserID`/`BookID` → `ErrShelfItemMismatch`; **порядок**: Completed и страница −1 → `ErrReadingFinished`; **порядок**: чужая полка в статусе Completed → `ErrShelfItemMismatch`. В негативных случаях проверять неизменность `*l` и `*item` через `reflect.DeepEqual` (с обнулением `now`). Прогнать `go test ./internal/domain/ -run TestReadingLog_UpdateProgress` — тесты ПАДАЮТ. Коммит `test: add failing tests for ReadingLog.UpdateProgress`.
- [ ] T014 [US1] Реализовать `UpdateProgress` в `internal/domain/reading_log.go` по порядку проверок FR-018: (1) `item!=nil` и несовпадение `UserID`/`BookID` → `ErrShelfItemMismatch`; (2) `item!=nil` и завершающий статус → `ErrReadingFinished`; (3) `page<0` → `ErrNegativePage`; (4) `page>totalPages` → `ErrPageExceedsTotal`. Все проверки ДО изменений. При успехе: `l.Pages = &page`; если `item!=nil` — взять `t := item.currentTime()` один раз; `WantToRead`→`Reading`; при `item.StartDate==nil` — `item.StartDate = &t`; при `page==totalPages` — `item.Status = StatusCompleted` и `item.FinishDate = &t` (новый указатель, не запись через старый). Прогнать тесты — ЗЕЛЁНЫЕ. Коммит `feat: implement ReadingLog.UpdateProgress logic`.

**Checkpoint**: US1 реализована и независимо проверяема.

---

## Phase 4: User Story 2 — Прогресс по аудиокниге (P1)

**Goal**: `ReadingLog.UpdateAudioProgress` — то же, что US1, но с секундами и длительностью.

**Independent Test**: как US1, но `seconds`/`totalSeconds` и влияние на полку по условию `seconds==totalSeconds`.

- [ ] T015 [US2] Табличные тесты `TestReadingLog_UpdateAudioProgress` в `internal/domain/reading_log_test.go`. Случаи из контракта: 12 000 из 36 000; 0; 36 000 из 36 000 (Completed, `*FinishDate==now`); WantToRead 600 (→ Reading, `*StartDate==now`); Reading без `StartDate`, 36 000 из 36 000 (`*StartDate==*FinishDate==now`); `item==nil` (меняется только `Seconds`); −1 с → `ErrNegativeSeconds`; 36 001 из 36 000 → `ErrSecondsExceedTotal`; Completed/Dropped → `ErrReadingFinished`; чужой `UserID`/`BookID` → `ErrShelfItemMismatch`; **порядок**: Dropped и −1 с → `ErrReadingFinished`; **порядок**: чужая полка в статусе Dropped → `ErrShelfItemMismatch`. Проверять, что `l.Pages` не меняется, и неизменность состояния в негативных случаях. `go test -run TestReadingLog_UpdateAudioProgress` ПАДАЕТ. Коммит `test: add failing tests for ReadingLog.UpdateAudioProgress`.
- [ ] T016 [US2] Реализовать `UpdateAudioProgress` в `internal/domain/reading_log.go`: порядок как в US1, границы (3) `seconds<0` → `ErrNegativeSeconds`, (4) `seconds>totalSeconds` → `ErrSecondsExceedTotal`. При успехе `l.Seconds = &seconds`; влияние на `item` как в T014 (`t := item.currentTime()`, `item.StartDate = &t`, `item.FinishDate = &t`) по условию `seconds==totalSeconds`. Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement ReadingLog.UpdateAudioProgress logic`.

**Checkpoint**: оба метода прогресса готовы.

---

## Phase 5: User Story 3 — Статус и даты на полке (P1)

**Goal**: `ShelfItem.ChangeStatus`, `SetStartDate`, `SetFinishDate`, `ClearDates` с правилами FR-015 – FR-017 и FR-026.

**Independent Test**: создать `ShelfItem` в разных статусах и с разными датами (`now` фиксирован), вызвать методы и проверить результат и неизменность при ошибке.

- [ ] T017 [US3] Табличные тесты `TestShelfItem_ChangeStatus` в `internal/domain/shelf_test.go`. Случаи: WantToRead→Reading, Completed→Reading, Dropped→WantToRead (успех, даты без изменений); смена на тот же статус (успех, без изменений); `Status(99)` и `Status(-1)` → `ErrInvalidStatus` (состояние не изменилось). `go test -run TestShelfItem_ChangeStatus` ПАДАЕТ. Коммит `test: add failing tests for ShelfItem.ChangeStatus`.
- [ ] T018 [US3] Реализовать `ChangeStatus` в `internal/domain/shelf.go`: недопустимое значение → `ErrInvalidStatus`; иначе `s.Status = status` (любой переход разрешён, даты не трогаются; смена на текущий статус — тоже успех). Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement ShelfItem.ChangeStatus logic`.
- [ ] T019 [US3] Табличные тесты `TestShelfItem_SetStartDate` в `internal/domain/shelf_test.go`. Случаи: дата в прошлом; дата `==now`; дата `==FinishDate` (успех); дата в 2030 году при `now`=2026-10-15 → `ErrStartDateInFuture`; дата позже заданной `FinishDate` → `ErrStartDateAfterFinish`; **порядок**: дата в 2030 году при заданной `FinishDate` раньше неё → `ErrStartDateInFuture`. Негативные — неизменность. `go test -run TestShelfItem_SetStartDate` ПАДАЕТ. Коммит `test: add failing tests for ShelfItem.SetStartDate`.
- [ ] T020 [US3] Реализовать `SetStartDate` в `internal/domain/shelf.go` по порядку: (1) `d.After(s.currentTime())` → `ErrStartDateInFuture`; (2) `s.FinishDate!=nil && d.After(*s.FinishDate)` → `ErrStartDateAfterFinish`; иначе `s.StartDate = &d` (статус не важен и не меняется). Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement ShelfItem.SetStartDate logic`.
- [ ] T021 [US3] Табличные тесты `TestShelfItem_SetFinishDate` в `internal/domain/shelf_test.go`. Случаи: Completed/Dropped с датой между стартом и `now`; дата `==StartDate` (успех); статус Reading/WantToRead → `ErrFinishDateNotAllowed`; дата в 2030 году → `ErrFinishDateInFuture`; `StartDate==nil` → `ErrStartDateMissing`; дата раньше `StartDate` → `ErrFinishDateBeforeStart`; **порядок**: статус Reading и дата в 2030 году → `ErrFinishDateNotAllowed`; **порядок**: Completed без `StartDate` и дата в 2030 году → `ErrFinishDateInFuture`. Негативные — неизменность. `go test -run TestShelfItem_SetFinishDate` ПАДАЕТ. Коммит `test: add failing tests for ShelfItem.SetFinishDate`.
- [ ] T022 [US3] Реализовать `SetFinishDate` в `internal/domain/shelf.go` по порядку FR-017: (1) статус не Completed и не Dropped → `ErrFinishDateNotAllowed`; (2) `d.After(s.currentTime())` → `ErrFinishDateInFuture`; (3) `s.StartDate==nil` → `ErrStartDateMissing`; (4) `d.Before(*s.StartDate)` → `ErrFinishDateBeforeStart`; иначе `s.FinishDate = &d`. Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement ShelfItem.SetFinishDate logic`.
- [ ] T023 [US3] Табличные тесты `TestShelfItem_ClearDates` в `internal/domain/shelf_test.go`. Случаи (у каждого хотя бы одна дата задана, чтобы подтест падал на пустой заглушке): Reading после отката, обе даты заданы → обе `nil`, статус Reading; Completed только со `StartDate` → обе `nil`, статус Completed; Completed только с `FinishDate` → обе `nil`, статус Completed; сценарий перечитывания: Reading со `StartDate`=2026-09-01 и `FinishDate`=2026-09-10 → `ClearDates()` → `SetStartDate(2026-10-01)` возвращает `nil` (до очистки тот же вызов дал бы `ErrStartDateAfterFinish`). `go test -run TestShelfItem_ClearDates` ПАДАЕТ. Коммит `test: add failing tests for ShelfItem.ClearDates`.
- [ ] T024 [US3] Реализовать `ClearDates` в `internal/domain/shelf.go`: `s.StartDate = nil`, `s.FinishDate = nil`; статус не трогается; ошибки нет (FR-026). Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement ShelfItem.ClearDates logic`.

**Checkpoint**: полка полностью управляема, перечитывание возможно.

---

## Phase 6: User Story 4 — Оценка в отзыве (P2)

**Goal**: `Review.SetRating` принимает только 1–5, иначе `ErrInvalidRating` (инвариант `DESIGN.md` §7.5).

**Independent Test**: вызвать с граничными и недопустимыми значениями, проверить результат и неизменность при ошибке.

- [ ] T025 [US4] Табличные тесты `TestReview_SetRating` в `internal/domain/review_test.go`. Случаи: 1, 3, 5 (успех, `r.Rating` установлен); 0, 6, −1 → `ErrInvalidRating` (оценка не изменилась). `go test -run TestReview_SetRating` ПАДАЕТ. Коммит `test: add failing tests for Review.SetRating`.
- [ ] T026 [US4] Реализовать `SetRating` в `internal/domain/review.go`: `rating<1 || rating>5` → `ErrInvalidRating`; иначе `r.Rating = rating`. Тесты ЗЕЛЁНЫЕ. Коммит `feat: implement Review.SetRating logic`.

**Checkpoint**: все 7 методов реализованы по TDD.

---

## Phase 7: User Story 5 — Модель предметной области (P3)

**Goal**: убедиться, что модель ядра соответствует `DESIGN.md` §4 (кроме задокументированных изменений `ShelfItem`).

**Independent Test**: ручная сверка + сборка без внешних зависимостей.

- [ ] T027 [US5] Сверить сущности в `internal/domain/*.go` с таблицами [data-model.md](./data-model.md) и `DESIGN.md` §4: 14 сущностей, поля и обязательность (`*` → указатель) совпадают; единственное допустимое отличие — `ShelfItem` (`Status` вместо `StatusID`, добавлены `StartDate`, `FinishDate`). Зафиксировать расхождения, если найдены (SC-006).

---

## Phase 8: Polish & Cross-Cutting

- [ ] T028 Удалить временную `errNotImplemented` из `internal/domain/errors.go` (после T026 она нигде не используется). Тесты ЗЕЛЁНЫЕ. Коммит `refactor: remove errNotImplemented`.
- [ ] T029 [P] `gofmt -l .` (пусто) и `go vet ./...` (чисто) по всему модулю.
- [ ] T030 [P] Детерминированность и гонки: `go test ./internal/domain/ -count=100` и `go test ./internal/domain/ -race` зелёные (SC-003); прогон укладывается в 5 с (SC-004).
- [ ] T031 Проверить ограничение «только стандартная библиотека»: `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/domain/` выводит только `github.com/PancakeNM/reading-tracker/internal/domain`; в `go.mod` нет `require`, файла `go.sum` нет (FR-003).
- [ ] T032 Проверить прослеживаемость Red→Green по `git log --oneline -- internal/domain/`: для каждого из 7 методов коммит `test:` предшествует `feat:`; первый коммит — `chore: scaffold …` (SC-005, принцип II).

---

## Dependencies & Execution Order

- **Setup (Phase 1)** → **Foundational (Phase 2)** → истории.
- **Phase 2 блокирует всё**: структуры, тип `Status`, ошибки и заглушки нужны всем методам.
- После Phase 2 истории технически независимы (каждый метод в своём методе, тесты в своём файле или своей функции):
  - US1 (T013–T014) и US2 (T015–T016) правят `reading_log.go` / `reading_log_test.go` — вести последовательно, один файл.
  - US3 (T017–T024) правит `shelf.go` / `shelf_test.go` — не конфликтует с US1/US2, может идти параллельно. Внутри US3: T023 (тест перечитывания) вызывает `SetStartDate`, поэтому ClearDates делается после T020.
  - US4 (T025–T026) правит `review.go` / `review_test.go` — независима.
- **Внутри каждого метода**: задача `test:` строго раньше задачи `feat:` (принцип II, нарушать нельзя).
- T028 — только после реализации всех 7 методов (иначе заглушки перестанут компилироваться).
- US5 (T027) и остальной Polish (T029–T032) — после T028.

## Parallel Opportunities

- **Phase 2**: T003–T007 помечены `[P]` — разные файлы, можно писать одновременно (но коммит общий, T012).
- **Между историями**: US3 и US4 можно вести параллельно US1/US2 (разные файлы). US1 и US2 — последовательно (общий файл `reading_log.go`).
- **Phase 8**: T029 и T030 параллельны.

## Implementation Strategy

- **MVP = User Story 1** (T001–T014): каркас + запись прогресса по текстовой книге с автозавершением. Уже даёт проверяемую бизнес-ценность.
- **Инкременты**: US2 (аудио) → US3 (полка, включая очистку дат) → US4 (оценка). Каждый добавляет методы, не ломая предыдущие.
- **Правило коммитов на каждом методе**: `test:` (красный) → `feat:` (зелёный). Это и есть доказательство TDD для ревью.

## Notes

- Шаблон `tasks-template` в `.specify/templates/` отсутствует — использована стандартная структура Spec Kit.
- Go на машине пока не установлен; установка — предусловие из [quickstart.md](./quickstart.md) (`winget install GoLang.Go`), нужна начиная с T012.
- Порядок методов здесь совпадает с [research.md](./research.md) R11.
