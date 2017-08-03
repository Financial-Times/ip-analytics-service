package queue

import (
	"github.com/streadway/amqp"
)

var (
	queueName = "ip.events.test"
	msgChan   = make(chan Message)
	sessions  = make(chan chan Session, 1)
	sess      = make(chan Session, 1)
)

func init() {
	conn, err := amqp.Dial("amqp://localhost")
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	_, err = Declare(queueName, ch)
	if err != nil {
		panic(err)
	}

	sessions <- sess
	sess <- Session{conn, ch}
}

//func countMessages(msgs <-chan amqp.Delivery) int {
//var cnt = 0
//for {
//select {
//case d, _ := <-msgs:
//log.Printf("Found Message: %s", string(d.Body[:]))
//d.Ack(false)
//if d.Body == nil {
//return cnt
//}
//cnt++
//case <-time.After(50 * time.Millisecond):
//return cnt
//}
//}
//}
