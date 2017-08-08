package queue

import (
	"testing"
)

func TestPublisherConfirm(t *testing.T) {
	go func() {
		Publish(sessions, msgChan, queueName)
	}()
	confirm := make(chan bool, 1)
	msg := []byte("hello world")
	msgChan <- Message{msg, confirm}
	ok := <-confirm
	if !ok {
		t.Errorf("Expected true from published confirmation, got: %v", ok)
	}
}
