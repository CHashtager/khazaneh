package expense

import "time"

type Transaction struct {
	ID          int64
	UserID      int64
	AccountID   *int64
	CategoryID  *int64
	Category    string
	Kind        string
	Amount      string
	Currency    string
	Merchant    string
	Note        string
	RawText     string
	OccurredAt  time.Time
	JalaliYear  int
	JalaliMonth int
	JalaliDay   int
	JalaliYM    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateTransaction struct {
	UserID     int64
	CategoryID *int64
	Kind       string
	Amount     string
	Currency   string
	Merchant   string
	Note       string
	RawText    string
	OccurredAt time.Time
}

type MonthlySummary struct {
	JalaliYM     string
	ExpenseTotal string
	IncomeTotal  string
}
