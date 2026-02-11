package models

import (
	"database/sql"

	tele "gopkg.in/telebot.v3"
)

type AppContext struct {
	DB  *sql.DB
	Bot *tele.Bot
}
