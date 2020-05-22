package server

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
	"github.com/ttlv/common_utils/health_checker"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/queue/amqp_queue"
	"log"
	"sort"
)

type Consumer struct {
	Providers []sms.SmsProvider
	DB        *gorm.DB
	Queue     sms.SmsQueue
}

func (c Consumer) Liveness() ([]byte, bool) {
	var (
		dbALive       = health_checker.PingDB(c.DB)
		rabbitMQALive = c.Queue.Liveness()
		err           error
		data          []byte
		alive         = false
	)
	healthCheck := struct {
		DBALive       bool
		RabbitMQALive bool
	}{
		DBALive: dbALive, RabbitMQALive: rabbitMQALive,
	}
	if dbALive && rabbitMQALive {
		alive = true
	}
	data, err = json.Marshal(healthCheck)
	if err != nil {
		return []byte(""), false
	}
	return data, alive
}

func (c Consumer) Readiness() bool {
	return true
}

func (c *Consumer) Run() {
	var (
		conf    = config.MustGetConfig()
		db, err = gorm.Open("mysql", conf.DB)
		conn    *amqp.Connection
		ch      *amqp.Channel
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
	smsServer, _ := New(db, smsQueue, c.Providers)
	msgs, err := ch.Consume(smsQueue.Queue.Name, "", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	c.DB = db
	c.Queue = smsQueue
	health_checker.StartCheck(c)
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			publishData := &sms.PublishData{}
			if err := json.Unmarshal(d.Body, publishData); err != nil {
				// TODO: Write to some place
				return
			}
			smsServer.Providers = c.Providers
			smsServer.RealSend(publishData)
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func (c *Consumer) SortProviders() {
	sorts := sms.FormattedProviderSorts(c.DB)
	log.Printf("当前的Provider排序: %v\n", sorts)
	sort.Slice(c.Providers, func(i, j int) bool {
		iIndex := indexOf(sorts, c.Providers[i].GetCode())
		jIndex := indexOf(sorts, c.Providers[j].GetCode())
		return iIndex < jIndex
	})
}

func indexOf(elements []string, element string) int {
	for i, e := range elements {
		if e == element {
			return i
		}
	}
	return 10000
}
