package user

import "time"

type User struct {
	ID             int64
	TelegramUserID int64
	DisplayName    string
	Username       string
	Currency       string
	Timezone       string
	CreatedAt      time.Time
}
