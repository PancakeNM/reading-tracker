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
