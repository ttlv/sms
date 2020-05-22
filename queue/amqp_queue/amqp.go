package amqp_queue

import (
	"encoding/json"

	"github.com/streadway/amqp"
	"github.com/ttlv/sms"
)

const QUEUE_KEY = "sms"

type AMQPQueue struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      amqp.Queue
	isAlive    bool
}

func New(conn *amqp.Connection, ch *amqp.Channel) *AMQPQueue {
	q := &AMQPQueue{Connection: conn, Channel: ch, isAlive: true}
	queue, err := ch.QueueDeclare(QUEUE_KEY, false, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	q.Queue = queue

	closeNotify := conn.NotifyClose(make(chan *amqp.Error))
	go func(q *AMQPQueue) {
		_ = <-closeNotify
		q.isAlive = false
	}(q)
	return q
}

func (q AMQPQueue) Liveness() bool {
	return q.isAlive
}

func (q AMQPQueue) Publish(data *sms.PublishData) {
	var (
		publishData []byte
		err         error
	)
	if publishData, err = json.Marshal(data); err != nil {
		panic(err)
	}
	err = q.Channel.Publish(
		"",           // exchange
		q.Queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        publishData,
		},
	)
	if err != nil {
		panic(err)
	}
}
