package home

import (
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"net/http"
)

type Handlers struct {
	DB           *gorm.DB
	SessionStore *sessions.CookieStore
}

func NewHandlers(db *gorm.DB, sessionStore *sessions.CookieStore) Handlers {
	return Handlers{DB: db, SessionStore: sessionStore}
}

func (handlers Handlers) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("SMS Service Is Working Now...."))
}
