package http_server

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/service/producer_http/app/action"
	"github.com/ttlv/sms/service/producer_http/home"
	"net/http"
	"strings"
)

func New(db *gorm.DB, sessionStore *sessions.CookieStore) *chi.Mux {
	var (
		c    = config.MustGetConfig()
		conn *amqp.Connection
		err  error
		ch   *amqp.Channel
	)
	SetDBCallback(db)
	if conn, err = amqp.Dial(c.AMQPDial); err != nil {
		panic("Failed to connect to RabbitMQ")
	}
	if ch, err = conn.Channel(); err != nil {
		panic("Failed to open a channel")
	}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(Authentication(sessionStore))
	homeHandlers := home.NewHandlers(db, sessionStore)
	router.Get("/", homeHandlers.Home)
	actionHandlers := action.NewHandlers(db, sessionStore, conn, ch)
	router.Post("/send", actionHandlers.Send)
	return router
}

func SetDBCallback(db *gorm.DB) {
	db.AutoMigrate(&sms.SmsRecord{}, &sms.SmsFailureRecord{}, &sms.SmsAvailable{})
}

func Authentication(sessionStore *sessions.CookieStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			session, err := sessionStore.Get(r, "sms_service")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			authStr := r.Header.Get("Authorization")
			accessToken := strings.TrimPrefix(authStr, "Token token=")
			if accessToken != "" {
				session.Values["token"] = accessToken
				session.Save(r, w)
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
