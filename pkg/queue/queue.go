package queue

type Job struct {
	JobID        int64
	Source       string
	IndustryType string
	Location     string
}

type JobQueue struct {
	Jobs chan Job
}

func NewJobQueue(size int64) *JobQueue {
	return &JobQueue{
		Jobs: make(chan Job, size),
	}
}

func (j JobQueue) Enqueue(id int64, source string, industryType string, location string) {
	j.Jobs <- Job{
		JobID:        id,
		Source:       source,
		IndustryType: industryType,
		Location:     location,
	}
}
