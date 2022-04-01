package broker

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"sync"
)

//TODO Add buffer overflow handler

type handler func(conn net.Conn) error

type Subscriber struct {
	// ID is subscribers id
	ID string
	// Conn is the connection of current subscriber
	Conn net.Conn
	// Topic is the topic subscribed
	Topic string
	// Handler is the handler for subscriber
	Handler handler
}

type Broker struct {
	// MaxLen is the maximum length of the queue
	MaxLen int
	// Queue is a map of topic: payload which is the server messages
	Queue map[string][]map[string]string
	// Subscribers is the list of subscribers with subscribed topics
	Subscribers map[string][]*Subscriber
	// Mutex is used for security during concurrent execution
	Mutex sync.Mutex
}

func NewBroker(maxLen int) *Broker {
	queue := make(map[string][]map[string]string)
	subscribers := make(map[string][]*Subscriber)

	return &Broker{
		MaxLen:      maxLen,
		Queue:       queue,
		Subscribers: subscribers,
	}
}

func (broker *Broker) Publish(topic string, payload map[string]string) error {

	broker.Mutex.Lock()

	if len(broker.Queue[topic]) > broker.MaxLen {
		broker.Mutex.Unlock()
		return nil
		//return Overflow
	}

	broker.Queue[topic] = append(broker.Queue[topic], payload)

	subs, ok := broker.Subscribers[topic]

	broker.Mutex.Unlock()
	if !ok {
		return nil
	}

	//generate new event to send to subscribers?

	broker.Mutex.Lock()
	defer broker.Mutex.Unlock()

	for _, sub := range subs {
		if err := sub.Handler(sub.Conn); err != nil {
			return err
		}
	}

	delete(broker.Queue, topic)

	return nil
}

func (broker *Broker) Subscribe(topic string, conn net.Conn, Handler handler) (*Subscriber, error) {

	newSub := &Subscriber{
		ID:      uuid.New().String(),
		Conn:    conn,
		Topic:   topic,
		Handler: Handler,
	}

	broker.Mutex.Lock()
	broker.Subscribers[topic] = append(broker.Subscribers[topic], newSub)
	broker.Mutex.Unlock()

	return newSub, nil
}

func (broker *Broker) Stop() {
	fmt.Println("Broker Exit")
}
