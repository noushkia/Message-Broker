package msg_broker

import (
	"github.com/google/uuid"
	"math/rand"
	"sync"
	"time"
)

// Event is given to a subscription handler for processing
type Event interface {
	Topic() string
	Message() []byte
	Ack() error
	Error() error
}

// Event is the object returned by the broker when you subscribe to a topic
type event struct {
	// ID to uniquely identify the event
	ID string
	// topic of event, e.g. "registry.service.created"
	topic string
	// Timestamp of the event
	Timestamp time.Time
	// Payload contains the encoded message
	Payload []byte
	// Err contains possible errors of the event
	Err error
}

func (e *event) Topic() string {
	return e.topic
}

func (e *event) Message() []byte {
	return e.Payload
}

func (e *event) Error() error {
	return e.Err
}

func (e *event) Ack() error {
	return nil
}

// Broker is an interface used for asynchronous messaging.
type Broker interface {
	Init(...Option) error
	Options() Options
	Publish(topic string, m []byte) error
	Subscribe(topic string, h Handler) (Subscriber, error)
	String() string
}

// Handler is used to process messages via a subscription of a topic.
// The handler is passed a publication interface which contains the
// message and optional Ack method to acknowledge receipt of the message.
type Handler func(Event) error

// Subscriber is a convenience return type for the Subscribe method
type Subscriber struct {
	ID      string
	Topic   string
	Handler Handler
}

type MQBroker struct {
	// options to store broker options
	options Options
	// subscribers to store broker subscribers
	subscribers map[string][]*Subscriber
	// messageQueue to store messages
	messageQueue map[string][][]byte
	// mu to utilize a mutex lock for race conditions
	mu sync.Mutex
}

func NewMQBroker(opts ...Option) *MQBroker {
	rand.Seed(time.Now().UnixNano())

	mq := &MQBroker{
		messageQueue: make(map[string][][]byte),
		subscribers:  make(map[string][]*Subscriber),
	}

	for _, o := range opts {
		o(&mq.options)
	}

	return mq
}

func (M *MQBroker) Init(opts ...Option) error {
	for _, o := range opts {
		o(&M.options)
	}
	return nil
}

func (M *MQBroker) Options() Options {
	return M.options
}

func (M *MQBroker) Publish(topic string, m []byte) error {
	//TODO Check topic
	M.mu.Lock()

	//TODO Check overflow

	M.messageQueue[topic] = append(M.messageQueue[topic], m)

	subscribers, ok := M.subscribers[topic]

	M.mu.Unlock()

	if !ok {
		return nil
	}

	ev := &event{
		topic:     topic,
		Payload:   m,
		Timestamp: time.Now(),
	}

	M.mu.Lock()
	defer M.mu.Unlock()

	// Check all subscribers for error
	for _, subscriber := range subscribers {
		if err := subscriber.Handler(ev); err != nil {
			//TODO add error handler
			return err
		}
	}

	delete(M.messageQueue, topic)

	return nil
}

func (M *MQBroker) Subscribe(topic string, h Handler) (*Subscriber, error) {
	//TODO Check topic

	newSub := &Subscriber{
		ID:      uuid.New().String(),
		Topic:   topic,
		Handler: h,
	}

	M.mu.Lock()

	M.subscribers[topic] = append(M.subscribers[topic], newSub)

	M.mu.Unlock()

	go func() {
		for _, msg := range M.messageQueue[topic] {
			_ = h(&event{
				topic:   topic,
				Payload: msg,
			})
		}

		delete(M.messageQueue, topic)
	}()

	return newSub, nil
}

func (M *MQBroker) String() string {
	return "Message Queue Broker"
}
