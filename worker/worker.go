package worker

type Job struct {
	// worker ID
	ID string
	// worker Task
	Task func(map[string]string)
	// Payload is the message
	Payload map[string]string
	//TODO convert to Message
}

func (job Job) Execute() {
	job.Task(job.Payload)
}
