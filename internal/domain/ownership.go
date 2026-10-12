package domain

import "time"

// AcquisitionType — тип получения книги. 1-to-many к BookOwnership.
type AcquisitionType struct {
	ID   int64
	Name string
}

// Currency — валюта цены покупки. 1-to-many к BookOwnership.
type Currency struct {
	ID   int64
	Name string
}

// BookOwnership — владение книгой пользователем.
type BookOwnership struct {
	ID                int64
	UserID            int64 // many-to-1 к User
	BookID            int64 // many-to-1 к Book
	Acquired          time.Time
	AcquisitionTypeID int64      // many-to-1 к AcquisitionType
	PriceCents        *int64     // копейки или центы, для покупки
	CurrencyID        *int64     // many-to-1 к Currency, для покупки
	ReturnDate        *time.Time // для библиотечной аренды
}
