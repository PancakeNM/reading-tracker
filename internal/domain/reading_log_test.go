package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestReadingLog_UpdateProgress(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	oldFinish := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)

	const (
		userID int64 = 10
		bookID int64 = 20
	)

	tests := []struct {
		name          string
		page          int
		totalPages    int
		itemNil       bool
		status        Status
		userID        int64
		bookID        int64
		setStart      bool
		start         time.Time
		setFinish     bool
		finish        time.Time
		wantErr       error
		wantStatus    Status
		wantPages     int
		startIsNow    bool
		finishIsNow   bool
		finishNil     bool
		samePointers  bool
		keepOldFinish bool
	}{
		{
			name:       "150 из 300, читаю, дата окончания пустая",
			page:       150,
			totalPages: 300,
			status:     StatusReading,
			userID:     userID,
			bookID:     bookID,
			setStart:   true,
			start:      start,
			wantStatus: StatusReading,
			wantPages:  150,
			finishNil:  true,
		},
		{
			name:       "страница 0",
			page:       0,
			totalPages: 300,
			status:     StatusReading,
			userID:     userID,
			bookID:     bookID,
			setStart:   true,
			start:      start,
			wantStatus: StatusReading,
			wantPages:  0,
			finishNil:  true,
		},
		{
			name:        "300 из 300, дата старта задана",
			page:        300,
			totalPages:  300,
			status:      StatusReading,
			userID:      userID,
			bookID:      bookID,
			setStart:    true,
			start:       start,
			wantStatus:  StatusCompleted,
			wantPages:   300,
			finishIsNow: true,
		},
		{
			name:       "хочу прочитать, 20 из 300",
			page:       20,
			totalPages: 300,
			status:     StatusWantToRead,
			userID:     userID,
			bookID:     bookID,
			wantStatus: StatusReading,
			wantPages:  20,
			startIsNow: true,
			finishNil:  true,
		},
		{
			name:         "читаю без даты старта, 300 из 300",
			page:         300,
			totalPages:   300,
			status:       StatusReading,
			userID:       userID,
			bookID:       bookID,
			wantStatus:   StatusCompleted,
			wantPages:    300,
			startIsNow:   true,
			finishIsNow:  true,
			samePointers: true,
		},
		{
			name:          "старая дата окончания заменяется новым указателем",
			page:          300,
			totalPages:    300,
			status:        StatusReading,
			userID:        userID,
			bookID:        bookID,
			setStart:      true,
			start:         start,
			setFinish:     true,
			finish:        oldFinish,
			wantStatus:    StatusCompleted,
			wantPages:     300,
			finishIsNow:   true,
			keepOldFinish: true,
		},
		{
			name:       "полки нет, 300 из 300",
			page:       300,
			totalPages: 300,
			itemNil:    true,
			wantPages:  300,
		},
		{
			name:       "страница −1",
			page:       -1,
			totalPages: 300,
			status:     StatusReading,
			userID:     userID,
			bookID:     bookID,
			setStart:   true,
			start:      start,
			wantErr:    ErrNegativePage,
		},
		{
			name:       "301 из 300",
			page:       301,
			totalPages: 300,
			status:     StatusReading,
			userID:     userID,
			bookID:     bookID,
			setStart:   true,
			start:      start,
			wantErr:    ErrPageExceedsTotal,
		},
		{
			name:       "статус прочитано",
			page:       50,
			totalPages: 300,
			status:     StatusCompleted,
			userID:     userID,
			bookID:     bookID,
			setStart:   true,
			start:      start,
			setFinish:  true,
			finish:     oldFinish,
			wantErr:    ErrReadingFinished,
		},
		{
			name:       "статус брошено",
			page:       50,
			totalPages: 300,
			status:     StatusDropped,
			userID:     userID,
			bookID:     bookID,
			setStart:   true,
			start:      start,
			setFinish:  true,
			finish:     oldFinish,
			wantErr:    ErrReadingFinished,
		},
		{
			name:       "чужой пользователь",
			page:       50,
			totalPages: 300,
			status:     StatusReading,
			userID:     99,
			bookID:     bookID,
			setStart:   true,
			start:      start,
			wantErr:    ErrShelfItemMismatch,
		},
		{
			name:       "чужая книга",
			page:       50,
			totalPages: 300,
			status:     StatusReading,
			userID:     userID,
			bookID:     99,
			setStart:   true,
			start:      start,
			wantErr:    ErrShelfItemMismatch,
		},
		{
			name:       "порядок: прочитано и страница −1",
			page:       -1,
			totalPages: 300,
			status:     StatusCompleted,
			userID:     userID,
			bookID:     bookID,
			setStart:   true,
			start:      start,
			setFinish:  true,
			finish:     oldFinish,
			wantErr:    ErrReadingFinished,
		},
		{
			name:       "порядок: чужая полка в статусе прочитано",
			page:       50,
			totalPages: 300,
			status:     StatusCompleted,
			userID:     99,
			bookID:     bookID,
			setStart:   true,
			start:      start,
			setFinish:  true,
			finish:     oldFinish,
			wantErr:    ErrShelfItemMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			seconds := int64(120)
			comment := "заметка"
			initialPage := 7
			log := &ReadingLog{
				ID:      1,
				UserID:  userID,
				BookID:  bookID,
				Date:    time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
				Pages:   &initialPage,
				Seconds: &seconds,
				Comment: &comment,
			}

			var item *ShelfItem
			if !tc.itemNil {
				item = &ShelfItem{
					ID:     7,
					UserID: tc.userID,
					BookID: tc.bookID,
					Status: tc.status,
					now:    func() time.Time { return now },
				}
				if tc.setStart {
					s := tc.start
					item.StartDate = &s
				}
				if tc.setFinish {
					f := tc.finish
					item.FinishDate = &f
				}
			}

			beforeLog := *log
			var beforeItem ShelfItem
			if item != nil {
				beforeItem = *item
			}
			var finishSnapshot *time.Time
			if tc.keepOldFinish {
				finishSnapshot = item.FinishDate
			}
			startPtr := (*time.Time)(nil)
			if item != nil {
				startPtr = item.StartDate
			}

			err := log.UpdateProgress(item, tc.page, tc.totalPages)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ошибка %v, ожидалась %v", err, tc.wantErr)
				}
				if !reflect.DeepEqual(beforeLog, *log) {
					t.Fatalf("запись о чтении изменилась: %+v", log)
				}
				if item != nil {
					// Функции сравнимы через reflect.DeepEqual только как nil.
					beforeItem.now = nil
					gotItem := *item
					gotItem.now = nil
					if !reflect.DeepEqual(beforeItem, gotItem) {
						t.Fatalf("полка изменилась: %+v", item)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if log.Pages == nil || *log.Pages != tc.wantPages {
				t.Fatalf("страница = %v, ожидалась %d", log.Pages, tc.wantPages)
			}
			if log.Seconds != beforeLog.Seconds || log.Comment != beforeLog.Comment || !log.Date.Equal(beforeLog.Date) {
				t.Fatal("изменились секунды, комментарий или дата записи")
			}
			if log.ID != beforeLog.ID || log.UserID != beforeLog.UserID || log.BookID != beforeLog.BookID {
				t.Fatal("изменились идентификаторы записи о чтении")
			}
			if tc.itemNil {
				return
			}
			if item.Status != tc.wantStatus {
				t.Fatalf("статус = %s, ожидался %s", item.Status, tc.wantStatus)
			}
			if item.UserID != tc.userID || item.BookID != tc.bookID || item.ID != beforeItem.ID {
				t.Fatal("изменились идентификаторы полки")
			}
			if tc.finishNil && item.FinishDate != nil {
				t.Fatalf("дата окончания = %v, ожидалась пустая", item.FinishDate)
			}
			if tc.finishIsNow && (item.FinishDate == nil || !item.FinishDate.Equal(now)) {
				t.Fatalf("дата окончания = %v, ожидалась %v", item.FinishDate, now)
			}
			if tc.startIsNow && (item.StartDate == nil || !item.StartDate.Equal(now)) {
				t.Fatalf("дата старта = %v, ожидалась %v", item.StartDate, now)
			}
			if tc.setStart && !tc.startIsNow {
				if item.StartDate != startPtr || item.StartDate == nil || !item.StartDate.Equal(tc.start) {
					t.Fatalf("дата старта изменилась: %v", item.StartDate)
				}
			}
			if tc.samePointers && item.StartDate != item.FinishDate {
				t.Fatal("дата старта и дата окончания должны указывать на одно и то же значение текущего момента")
			}
			if tc.keepOldFinish {
				if finishSnapshot == nil || !finishSnapshot.Equal(oldFinish) {
					t.Fatalf("снимок даты окончания = %v, ожидалась %v", finishSnapshot, oldFinish)
				}
				if item.FinishDate == finishSnapshot {
					t.Fatal("дата окончания записана через старый указатель")
				}
			}
		})
	}
}

func TestReadingLog_UpdateAudioProgress(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	const (
		userID int64 = 10
		bookID int64 = 20
	)

	tests := []struct {
		name         string
		seconds      int64
		totalSeconds int64
		itemNil      bool
		status       Status
		userID       int64
		bookID       int64
		setStart     bool
		start        time.Time
		wantErr      error
		wantStatus   Status
		wantSeconds  int64
		startIsNow   bool
		finishIsNow  bool
		finishNil    bool
		samePointers bool
	}{
		{
			name:         "12 000 из 36 000",
			seconds:      12000,
			totalSeconds: 36000,
			status:       StatusReading,
			userID:       userID,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantStatus:   StatusReading,
			wantSeconds:  12000,
			finishNil:    true,
		},
		{
			name:         "0 секунд",
			seconds:      0,
			totalSeconds: 36000,
			status:       StatusReading,
			userID:       userID,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantStatus:   StatusReading,
			wantSeconds:  0,
			finishNil:    true,
		},
		{
			name:         "36 000 из 36 000, дата старта задана",
			seconds:      36000,
			totalSeconds: 36000,
			status:       StatusReading,
			userID:       userID,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantStatus:   StatusCompleted,
			wantSeconds:  36000,
			finishIsNow:  true,
		},
		{
			name:         "хочу прочитать, 600 с",
			seconds:      600,
			totalSeconds: 36000,
			status:       StatusWantToRead,
			userID:       userID,
			bookID:       bookID,
			wantStatus:   StatusReading,
			wantSeconds:  600,
			startIsNow:   true,
			finishNil:    true,
		},
		{
			name:         "читаю без даты старта, 36 000 из 36 000",
			seconds:      36000,
			totalSeconds: 36000,
			status:       StatusReading,
			userID:       userID,
			bookID:       bookID,
			wantStatus:   StatusCompleted,
			wantSeconds:  36000,
			startIsNow:   true,
			finishIsNow:  true,
			samePointers: true,
		},
		{
			name:         "полки нет",
			seconds:      36000,
			totalSeconds: 36000,
			itemNil:      true,
			wantSeconds:  36000,
		},
		{
			name:         "−1 с",
			seconds:      -1,
			totalSeconds: 36000,
			status:       StatusReading,
			userID:       userID,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantErr:      ErrNegativeSeconds,
		},
		{
			name:         "36 001 из 36 000",
			seconds:      36001,
			totalSeconds: 36000,
			status:       StatusReading,
			userID:       userID,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantErr:      ErrSecondsExceedTotal,
		},
		{
			name:         "статус прочитано",
			seconds:      600,
			totalSeconds: 36000,
			status:       StatusCompleted,
			userID:       userID,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantErr:      ErrReadingFinished,
		},
		{
			name:         "статус брошено",
			seconds:      600,
			totalSeconds: 36000,
			status:       StatusDropped,
			userID:       userID,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantErr:      ErrReadingFinished,
		},
		{
			name:         "чужой пользователь",
			seconds:      600,
			totalSeconds: 36000,
			status:       StatusReading,
			userID:       99,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantErr:      ErrShelfItemMismatch,
		},
		{
			name:         "чужая книга",
			seconds:      600,
			totalSeconds: 36000,
			status:       StatusReading,
			userID:       userID,
			bookID:       99,
			setStart:     true,
			start:        start,
			wantErr:      ErrShelfItemMismatch,
		},
		{
			name:         "порядок: брошено и −1 с",
			seconds:      -1,
			totalSeconds: 36000,
			status:       StatusDropped,
			userID:       userID,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantErr:      ErrReadingFinished,
		},
		{
			name:         "порядок: чужая полка в статусе брошено",
			seconds:      600,
			totalSeconds: 36000,
			status:       StatusDropped,
			userID:       99,
			bookID:       bookID,
			setStart:     true,
			start:        start,
			wantErr:      ErrShelfItemMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pages := 40
			seconds := int64(15)
			comment := "заметка"
			log := &ReadingLog{
				ID:      1,
				UserID:  userID,
				BookID:  bookID,
				Date:    time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
				Pages:   &pages,
				Seconds: &seconds,
				Comment: &comment,
			}

			var item *ShelfItem
			if !tc.itemNil {
				item = &ShelfItem{
					ID:     7,
					UserID: tc.userID,
					BookID: tc.bookID,
					Status: tc.status,
					now:    func() time.Time { return now },
				}
				if tc.setStart {
					s := tc.start
					item.StartDate = &s
				}
			}

			beforeLog := *log
			var beforeItem ShelfItem
			if item != nil {
				beforeItem = *item
			}
			startPtr := (*time.Time)(nil)
			if item != nil {
				startPtr = item.StartDate
			}

			err := log.UpdateAudioProgress(item, tc.seconds, tc.totalSeconds)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ошибка %v, ожидалась %v", err, tc.wantErr)
				}
				if !reflect.DeepEqual(beforeLog, *log) {
					t.Fatalf("запись о чтении изменилась: %+v", log)
				}
				if item != nil {
					beforeItem.now = nil
					gotItem := *item
					gotItem.now = nil
					if !reflect.DeepEqual(beforeItem, gotItem) {
						t.Fatalf("полка изменилась: %+v", item)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if log.Pages != beforeLog.Pages {
				t.Fatal("изменились страницы")
			}
			if log.Seconds == beforeLog.Seconds || log.Seconds == nil || *log.Seconds != tc.wantSeconds {
				t.Fatalf("секунды = %v, ожидались %d", log.Seconds, tc.wantSeconds)
			}
			if log.Comment != beforeLog.Comment || !log.Date.Equal(beforeLog.Date) {
				t.Fatal("изменились комментарий или дата записи")
			}
			if tc.itemNil {
				return
			}
			if item.Status != tc.wantStatus {
				t.Fatalf("статус = %s, ожидался %s", item.Status, tc.wantStatus)
			}
			if tc.finishNil && item.FinishDate != nil {
				t.Fatalf("дата окончания = %v, ожидалась пустая", item.FinishDate)
			}
			if tc.finishIsNow && (item.FinishDate == nil || !item.FinishDate.Equal(now)) {
				t.Fatalf("дата окончания = %v, ожидалась %v", item.FinishDate, now)
			}
			if tc.startIsNow && (item.StartDate == nil || !item.StartDate.Equal(now)) {
				t.Fatalf("дата старта = %v, ожидалась %v", item.StartDate, now)
			}
			if tc.setStart && !tc.startIsNow {
				if item.StartDate != startPtr || item.StartDate == nil || !item.StartDate.Equal(tc.start) {
					t.Fatalf("дата старта изменилась: %v", item.StartDate)
				}
			}
			if tc.samePointers && item.StartDate != item.FinishDate {
				t.Fatal("дата старта и дата окончания должны указывать на одно и то же значение текущего момента")
			}
		})
	}
}
