# Implementation Plan: Доменное ядро учёта чтения

**Branch**: `001-reading-domain-core` | **Date**: 2026-10-11 | **Revised**: 2026-10-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/001-reading-domain-core/spec.md`
**User constraint (this stage)**: без базы данных, без HTTP, без внешних библиотек — только стандартная библиотека Go и пакет `testing`.

## Summary

Создать Go-модуль с чистым доменным ядром: описать сущности `DESIGN.md` §4 как структуры данных и добавить поведение в три из них.

- `ShelfItem` — статус-перечисление и даты; методы `ChangeStatus`, `SetStartDate`, `SetFinishDate`.
- `ReadingLog` — запись прогресса по тексту (`UpdateProgress`) и по аудио (`UpdateAudioProgress`). Методы получают необязательный `*ShelfItem` и общий объём, проверяют статус полки и автоматически завершают чтение на полке при достижении конца.
- `Review` — `SetRating` с проверкой 1–5 (`ErrInvalidRating`).

Отдельной сущности `Reading` нет (решение пользователя 2026-10-12). Методы проверяют всё до изменения состояния и возвращают именованные сигнальные ошибки. Текущий момент берётся из неэкспортируемого поля `now` в `ShelfItem`. Каждый метод покрывается табличными тестами по циклу Red → Green с отдельными коммитами `test:` → `feat:`.

## Technical Context

**Language/Version**: Go 1.22+ (директива `go 1.22` в `go.mod`). Тулчейн на машине пока не установлен: см. [quickstart.md](./quickstart.md)
**Primary Dependencies**: только стандартная библиотека (`errors`, `time`; в тестах `testing`, `reflect`)
**Storage**: N/A — хранилища на этом этапе нет
**Testing**: `go test`, табличные тесты через `t.Run` с `t.Parallel()`, `errors.Is` для проверки ошибок
**Target Platform**: любая платформа Go; разработка на Windows
**Project Type**: библиотека-пакет внутри модуля (доменный слой будущего бэкенда)
**Performance Goals**: полный прогон тестов ядра не дольше 5 с (SC-004)
**Constraints**: нет сети, БД, HTTP и сторонних модулей (`go.mod` без `require`, нет `go.sum`); ядро не зависит от инфраструктуры (FR-003)
**Scale/Scope**: 14 сущностей `DESIGN.md`, 6 методов с поведением в 3 сущностях, 14 сигнальных ошибок

Все неизвестные закрыты в [research.md](./research.md) (R1–R11). Открытых NEEDS CLARIFICATION нет.

## Constitution Check

*GATE: проверено до Phase 0 и повторно после Phase 1.*

| Принцип | Статус | Как соблюдается |
|---------|--------|-----------------|
| I. Test-First Development | PASS | Для каждого из 6 методов тесты пишутся и падают раньше реализации. Заглушки не содержат логики (R11). |
| II. Traceable Red→Green Commits | PASS | Каркас с заглушками, затем 6 пар `test:` → `feat:` в порядке из R11. Без squash, amend и rebase. |
| III. Separation of Tests | PASS | Тесты в `*_test.go` (`shelf_test.go`, `reading_log_test.go`, `review_test.go`); Go исключает их из production-сборки. |
| IV. Microservices Architecture | JUSTIFIED DEVIATION | Пока сервисов нет: один модуль в корне, ядро в `internal/domain` (выбор пользователя, R2). См. Complexity Tracking. |
| V. Contract-First HTTP API | N/A (этап) | HTTP исключён указанием пользователя. Контракт — Go API пакета ([contracts/domain-api.md](./contracts/domain-api.md)), покрывается тестами до реализации. |
| Source of Truth | PASS | Сущности `DESIGN.md` §4 перенесены. Расхождения (`ShelfItem` с датами и статусом «брошено», позиция в `ReadingLog`) взяты из указаний чата, имеющих приоритет, и записаны в spec Assumptions. Инвариант §7.5 (оценка 1–5) покрыт тестами `Review.SetRating`. |

**Post-design re-check**: после Phase 1 нарушений не добавилось. Требования FR-001 – FR-025 отражены в [data-model.md](./data-model.md) и [contracts/domain-api.md](./contracts/domain-api.md).

## Project Structure

### Documentation (this feature)

```text
specs/001-reading-domain-core/
├── spec.md
├── plan.md              # этот файл
├── research.md          # Phase 0
├── data-model.md        # Phase 1
├── quickstart.md        # Phase 1
├── contracts/
│   └── domain-api.md    # Phase 1: публичный API пакета domain
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks, ещё не создан)
```

### Source Code (repository root)

```text
go.mod                         # module github.com/PancakeNM/reading-tracker, go 1.22, без require
internal/
└── domain/
    ├── user.go                # User
    ├── book.go                # Book, BookType, TextBook, AudioBook
    ├── author.go              # Author, BookAuthor
    ├── series.go              # Series
    ├── shelf.go               # ReadingStatus, Status, ShelfItem + ChangeStatus, SetStartDate, SetFinishDate
    ├── reading_log.go         # ReadingLog + UpdateProgress, UpdateAudioProgress
    ├── review.go              # Review + SetRating
    ├── ownership.go           # AcquisitionType, Currency, BookOwnership
    ├── errors.go              # сигнальные ошибки
    ├── shelf_test.go          # тесты ShelfItem (package domain)
    ├── reading_log_test.go    # тесты ReadingLog (package domain)
    └── review_test.go         # тесты Review (package domain)
```

**Structure Decision**: один модуль в корне репозитория, доменный пакет в `internal/domain`. Сущности разложены по файлам по смысловым группам `DESIGN.md`, поведение — рядом со своей сущностью. Тесты в том же пакете, чтобы задавать неэкспортируемое поле `now`.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Принцип IV: один модуль в корне вместо отдельного модуля сервиса | Выбор пользователя. На этапе только доменное ядро без развёртываемых сервисов, границы сервисов ещё не определены. | Модуль `services/reading/` или раскладка по нескольким сервисам фиксирует границы сервисов раньше, чем они спроектированы. Ядро в `internal/` можно перенести в модуль сервиса без изменения API. |
