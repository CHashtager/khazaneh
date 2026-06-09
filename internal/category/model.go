package category

import "time"

type Category struct {
	ID        int64
	UserID    int64
	Name      string
	Kind      string
	Emoji     string
	CreatedAt time.Time
}
