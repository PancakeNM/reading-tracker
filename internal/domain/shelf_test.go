package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestShelfItem_ChangeStatus(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	finish := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		from    Status
		to      Status
		wantErr error
	}{
		{name: "хочу прочитать → читаю", from: StatusWantToRead, to: StatusReading},
		{name: "прочитано → читаю", from: StatusCompleted, to: StatusReading},
		{name: "брошено → хочу прочитать", from: StatusDropped, to: StatusWantToRead},
		{name: "тот же статус", from: StatusReading, to: StatusReading},
		{name: "статус 99", from: StatusReading, to: Status(99), wantErr: ErrInvalidStatus},
		{name: "статус −1", from: StatusReading, to: Status(-1), wantErr: ErrInvalidStatus},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := start
			f := finish
			item := &ShelfItem{
				ID:         1,
				UserID:     10,
				BookID:     20,
				Status:     tc.from,
				StartDate:  &s,
				FinishDate: &f,
				now:        func() time.Time { return now },
			}
			before := *item

			err := item.ChangeStatus(tc.to)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ошибка %v, ожидалась %v", err, tc.wantErr)
				}
				assertShelfUnchanged(t, before, *item)
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			before.Status = tc.to
			assertShelfEqual(t, before, *item)
		})
	}
}

func TestShelfItem_SetStartDate(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	past := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	sameAsFinish := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	finishEarly := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	afterFinish := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	future := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		status    Status
		date      time.Time
		setFinish bool
		finish    time.Time
		wantErr   error
	}{
		{name: "дата в прошлом", status: StatusWantToRead, date: past},
		{name: "дата равна текущему моменту", status: StatusReading, date: now},
		{name: "дата равна дате окончания", status: StatusCompleted, date: sameAsFinish, setFinish: true, finish: sameAsFinish},
		{name: "дата в 2030 году", status: StatusReading, date: future, wantErr: ErrStartDateInFuture},
		{name: "дата позже даты окончания", status: StatusReading, date: afterFinish, setFinish: true, finish: finishEarly, wantErr: ErrStartDateAfterFinish},
		{name: "порядок: дата в 2030 году и дата окончания раньше", status: StatusCompleted, date: future, setFinish: true, finish: finishEarly, wantErr: ErrStartDateInFuture},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			item := &ShelfItem{
				ID:     1,
				UserID: 10,
				BookID: 20,
				Status: tc.status,
				now:    func() time.Time { return now },
			}
			if tc.setFinish {
				f := tc.finish
				item.FinishDate = &f
			}
			before := *item
			oldStart := item.StartDate

			err := item.SetStartDate(tc.date)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ошибка %v, ожидалась %v", err, tc.wantErr)
				}
				assertShelfUnchanged(t, before, *item)
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if item.StartDate == nil || !item.StartDate.Equal(tc.date) {
				t.Fatalf("дата старта = %v, ожидалась %v", item.StartDate, tc.date)
			}
			if item.StartDate == oldStart {
				t.Fatal("дата старта записана через старый указатель")
			}
			before.StartDate = item.StartDate
			assertShelfEqual(t, before, *item)
		})
	}
}

func TestShelfItem_SetFinishDate(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	between := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	earlier := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	future := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		status   Status
		setStart bool
		date     time.Time
		wantErr  error
	}{
		{name: "прочитано, дата между стартом и сейчас", status: StatusCompleted, setStart: true, date: between},
		{name: "брошено, дата между стартом и сейчас", status: StatusDropped, setStart: true, date: between},
		{name: "дата равна дате старта", status: StatusDropped, setStart: true, date: start},
		{name: "статус читаю", status: StatusReading, setStart: true, date: between, wantErr: ErrFinishDateNotAllowed},
		{name: "статус хочу прочитать", status: StatusWantToRead, setStart: true, date: between, wantErr: ErrFinishDateNotAllowed},
		{name: "дата в 2030 году", status: StatusCompleted, setStart: true, date: future, wantErr: ErrFinishDateInFuture},
		{name: "нет даты старта", status: StatusCompleted, date: between, wantErr: ErrStartDateMissing},
		{name: "дата раньше даты старта", status: StatusCompleted, setStart: true, date: earlier, wantErr: ErrFinishDateBeforeStart},
		{name: "порядок: читаю и дата в 2030 году", status: StatusReading, setStart: true, date: future, wantErr: ErrFinishDateNotAllowed},
		{name: "порядок: прочитано без даты старта и дата в 2030 году", status: StatusCompleted, date: future, wantErr: ErrFinishDateInFuture},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			item := &ShelfItem{
				ID:     1,
				UserID: 10,
				BookID: 20,
				Status: tc.status,
				now:    func() time.Time { return now },
			}
			if tc.setStart {
				s := start
				item.StartDate = &s
			}
			before := *item
			oldFinish := item.FinishDate

			err := item.SetFinishDate(tc.date)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ошибка %v, ожидалась %v", err, tc.wantErr)
				}
				assertShelfUnchanged(t, before, *item)
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if item.FinishDate == nil || !item.FinishDate.Equal(tc.date) {
				t.Fatalf("дата окончания = %v, ожидалась %v", item.FinishDate, tc.date)
			}
			if item.FinishDate == oldFinish || item.FinishDate == item.StartDate {
				t.Fatal("дата окончания записана через старый указатель")
			}
			before.FinishDate = item.FinishDate
			assertShelfEqual(t, before, *item)
		})
	}
}

func assertShelfUnchanged(t *testing.T, before, got ShelfItem) {
	t.Helper()
	assertShelfEqual(t, before, got)
}

func assertShelfEqual(t *testing.T, before, got ShelfItem) {
	t.Helper()
	before.now = nil
	got.now = nil
	if !reflect.DeepEqual(before, got) {
		t.Fatalf("полка = %+v, ожидалось %+v", got, before)
	}
}
