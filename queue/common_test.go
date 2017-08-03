package queue

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
