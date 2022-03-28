package msg_broker

type Options struct {
	QueueSize int
}

type Option func(*Options)

// QueueSize is the maximum number of messages in the queue.
func QueueSize(qs int) Option {
	return func(o *Options) {
		o.QueueSize = qs
	}
}
