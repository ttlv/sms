package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/rs/cors"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/service/producer_http/http_server"
	"log"
	"net/http"
)

func main() {
	cfg := config.MustGetConfig()
	db, err := gorm.Open("mysql", cfg.DB)
	if err != nil {
		panic("无法连接mysql")
		return
	}
	cs := cors.New(cors.Options{
		//AllowedOrigins:   []string{"http://localhost:3002"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
		Debug:            true,
	})

	sessionStore := sessions.NewCookieStore([]byte("GbeVMHok6yjFXTgDkwUzVMj"))

	router := http_server.New(db, sessionStore)
	handler := cs.Handler(router)

	log.Printf("========== Visit http://localhost%v ==========\n", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, handler))
}
