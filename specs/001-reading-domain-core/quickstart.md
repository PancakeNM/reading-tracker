# Quickstart: проверка доменного ядра

**Feature**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

## Предусловия

- Go 1.22 или новее. Сейчас на машине Go не установлен. Вариант установки на Windows:

  ```powershell
  winget install GoLang.Go
  ```

  После установки откройте новый терминал и проверьте: `go version`.
- Сеть, база данных и внешние сервисы не нужны.

## Сборка и статические проверки

Из корня репозитория:

```powershell
go build ./...
go vet ./...
gofmt -l .          # пустой вывод — код отформатирован
```

## Тесты

```powershell
go test ./internal/domain/ -v
go test ./internal/domain/ -run 'TestShelfItem_' -v
go test ./internal/domain/ -run 'TestReadingLog_' -v
go test ./internal/domain/ -run 'TestReview_' -v
go test ./internal/domain/ -count=100      # детерминированность (SC-003)
go test ./internal/domain/ -race           # гонки при t.Parallel() (нужен cgo/gcc)
```

**Ожидаемо**: все случаи из таблицы обязательных тестов в [contracts/domain-api.md](./contracts/domain-api.md) проходят, полный прогон занимает меньше 5 секунд (SC-004).

## Проверка Red → Green по истории (SC-005)

```powershell
git log --oneline -- internal/domain/
```

Для каждого из шести методов коммит `test: add failing tests for <Сущность>.<Метод>` идёт раньше `feat: implement <Сущность>.<Метод> logic`. Воспроизвести фазу Red:

```powershell
git checkout <hash коммита test:>
go test ./internal/domain/ -run 'Test<Сущность>_<Метод>'   # ожидается FAIL на проверках, не ошибка компиляции
git checkout 001-reading-domain-core
```

## Проверка модели (SC-006)

Сверить таблицы в [data-model.md](./data-model.md) с `DESIGN.md` §4 и файлами в `internal/domain/`. Отличаться может только `ShelfItem`: поле `Status` вместо `StatusID`, добавлены `StartDate` и `FinishDate`.

## Проверка ограничения «только стандартная библиотека»

```powershell
go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/domain/
```

**Ожидаемо**: единственная строка — `github.com/PancakeNM/reading-tracker/internal/domain`. В `go.mod` нет `require`, файла `go.sum` нет.
