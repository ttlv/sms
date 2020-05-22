package main

import (
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/queue/amqp_queue"
	"github.com/ttlv/sms/webhook"
)

func main() {
	var (
		conf    = config.MustGetConfig()
		db, err = gorm.Open("mysql", conf.DB)
		conn    *amqp.Connection
		ch      *amqp.Channel
		router  = mux.NewRouter()
	)
	if err != nil {
		panic(err)
	}
	if conn, err = amqp.Dial(conf.AMQPDial); err != nil {
		panic(err)
	}
	defer conn.Close()

	if ch, err = conn.Channel(); err != nil {
		panic(err)
	}
	defer ch.Close()

	smsQueue := amqp_queue.New(conn, ch)
	defer db.Close()
	if err != nil {
		panic(err)
	}
	server := webhook.NewWebHookServer(db, smsQueue)
	server.Run()
	log.Fatal(http.ListenAndServe(conf.Port, router))
}
