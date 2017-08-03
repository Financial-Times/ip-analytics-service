package queue

import (
	"context"
	"testing"
)

var msgChan = make(chan Message)

func init() {
	ctx, done := context.WithCancel(context.Background())
	queueName := "ip.events.test"
	go func() {
		Publish(Redial(ctx, "amqp://localhost", queueName), msgChan, queueName)
		done()
	}()
}

func TestPublisherPublish(t *testing.T) {
	confirm := make(chan bool, 1)
	msg := []byte("hello world")
	msgChan <- Message{msg, confirm}
	close(msgChan)
}

func TestPublisherConfirm(t *testing.T) {
	confirm := make(chan bool, 1)
	msg := []byte("hello world")
	msgChan <- Message{msg, confirm}
	ok := <-confirm
	if !ok {
		t.Errorf("Expected true from published confirmation, got: %v", ok)
	}
}

func TestPublisherReturnOnChanClose(t *testing.T) {
	confirm := make(chan bool, 1)
	msg := []byte("hello world")
	msgChan <- Message{msg, confirm}
	close(msgChan)
}
