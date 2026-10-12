package domain

import (
	"errors"
	"reflect"
	"testing"
)

func TestReview_SetRating(t *testing.T) {
	tests := []struct {
		name    string
		rating  int
		wantErr error
	}{
		{name: "оценка 1", rating: 1},
		{name: "оценка 3", rating: 3},
		{name: "оценка 5", rating: 5},
		{name: "оценка 0", rating: 0, wantErr: ErrInvalidRating},
		{name: "оценка 6", rating: 6, wantErr: ErrInvalidRating},
		{name: "оценка −1", rating: -1, wantErr: ErrInvalidRating},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			review := &Review{
				ID:     1,
				BookID: 2,
				UserID: 3,
				Rating: 4,
				Text:   "текст",
			}
			before := *review

			err := review.SetRating(tc.rating)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ошибка %v, ожидалась %v", err, tc.wantErr)
				}
				if !reflect.DeepEqual(before, *review) {
					t.Fatalf("отзыв изменился: %+v", review)
				}
				return
			}
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			before.Rating = tc.rating
			if !reflect.DeepEqual(before, *review) {
				t.Fatalf("отзыв = %+v, ожидалось %+v", review, before)
			}
		})
	}
}
