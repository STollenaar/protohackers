package problem9

import "net"

type Request struct {
	Request  string      `json:"request"`
	Id       int         `json:"id,omitempty"`
	Queue    string      `json:"queue,omitempty"`
	Queues   []string    `json:"queues,omitempty"`
	Job      interface{} `json:"job,omitempty"`
	Priority *int        `json:"pri,omitempty"`
	Wait     bool        `json:"wait,omitempty"`
}

type Response struct {
	Status   string      `json:"status"`
	Error    string      `json:"error,omitempty"`
	Id       int         `json:"id,omitempty"`
	Queue    string      `json:"queue,omitempty"`
	Job      interface{} `json:"job,omitempty"`
	Priority *int        `json:"pri,omitempty"`
}

type Client struct {
	conn        net.Conn
	id          string
	currentJobs []ClientJob
}

type ClientJob struct {
	queue string
	jobId int
}

type Job struct {
	Id       int         `json:"id"`
	Details  interface{} `json:"details"`
	Priority *int        `json:"priority"`
	Queue    string      `json:"queue"`
}

func (c *Client) closeConnection() {
	defer c.conn.Close()
	queues.AbortJobs(c.currentJobs)
}

func (c *Client) hasJob(id int) bool {
	for _, job := range c.currentJobs {
		if job.jobId == id {
			return true
		}
	}
	return false
}

func (c *Client) getJob(id int) ClientJob {
	for _, job := range c.currentJobs {
		if job.jobId == id {
			return job
		}
	}
	return ClientJob{}
}
