package broker

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net"
	"sync"
)

const workerSize = 10

//TODO Add buffer overflow handler

var OverflowError = errors.New("broker overflow")

type handler func(conn net.Conn, payLoad map[string]string) error

type PayLoad map[string]string

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

type Event struct {
	Topic   string
	payLoad PayLoad
}

type Broker struct {
	// MaxLen is the maximum length of the queue
	MaxLen int
	// Queue is a map of topic: payload which is the server messages
	Queue map[string][]PayLoad
	// Subscribers is the list of subscribers with subscribed topics
	Subscribers map[string][]*Subscriber
	// Mutex is used for security during concurrent execution
	Mutex sync.Mutex
	// Chan is used for asynchronous publishing
	// It's a channel that stores the topics to be published
	// with their respective payLoad
	Chan chan *Event
}

func NewBroker(maxLen int, workerSize int) *Broker {
	queue := make(map[string][]PayLoad)
	subscribers := make(map[string][]*Subscriber)

	newBroker := &Broker{
		MaxLen:      maxLen,
		Queue:       queue,
		Subscribers: subscribers,
		Chan:        make(chan *Event),
	}

	newBroker.startPublishChan(workerSize)

	return newBroker
}

func (broker *Broker) startPublishChan(size int) {
	for i := 0; i < size; i++ {
		go broker.publishAsync()
	}
}

func (broker *Broker) Publish(topic string, payload PayLoad) error {

	broker.Mutex.Lock()

	if len(broker.Queue[topic]) > broker.MaxLen {
		broker.Mutex.Unlock()
		return OverflowError
	}

	broker.Queue[topic] = append(broker.Queue[topic], payload)

	subs, ok := broker.Subscribers[topic]

	broker.Mutex.Unlock()
	if !ok {
		return nil
	}

	broker.Mutex.Lock()
	defer broker.Mutex.Unlock()

	// This can be fixed?
	for _, sub := range subs {
		if err := sub.Handler(sub.Conn, payload); err != nil {
			// Connection error; The client quit
			if _, t := err.(*net.OpError); t {
				//TODO Remove this sub
				continue
			} else {
				return err
			}
		}
	}

	delete(broker.Queue, topic)

	return nil
}

func (broker *Broker) publishAsync() {
	for event := range broker.Chan {
		if err := broker.Publish(event.Topic, event.payLoad); err != nil {
			continue
		}
	}
}

func (broker *Broker) PublishAsync(topic string, payload PayLoad) error {
	event := &Event{
		Topic:   topic,
		payLoad: payload,
	}

	broker.Chan <- event

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
