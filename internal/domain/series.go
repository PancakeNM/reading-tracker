package domain

// Series — серия книг. 1-to-many к Book.
type Series struct {
	ID   int64
	Name string // уникально без учёта регистра
}
