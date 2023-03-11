package problem9

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"sync"
)

type QueueMap struct {
	sync.RWMutex
	maxId      int
	ready      map[string]Queue
	inProgress map[int]*Job
}

type Queue struct {
	Jobs []*Job `json:"job"`
}

func (qu *QueueMap) Dump() {
	qu.Lock()
	defer qu.Unlock()

	data, _ := json.MarshalIndent(qu.ready, "", " ")
	ioutil.WriteFile("debug.json", data, 0644)
}

func NewQueueMap() *QueueMap {
	return &QueueMap{
		ready:      make(map[string]Queue),
		inProgress: make(map[int]*Job),
		maxId:      1,
	}
}

func (qu *QueueMap) Retrieve(keys []string) (value *Job, ok bool) {
	qu.Lock()
	defer qu.Unlock()

	for _, key := range keys {
		q := qu.ready[key]

		if len(q.Jobs) == 0 {
			continue
		}
		qJob := q.Jobs[0]
		if value == nil || *qJob.Priority > *value.Priority {
			value = qJob
		}
	}
	if value != nil {
		qu.inProgress[value.Id] = value
		q := qu.ready[value.Queue]
		q.Jobs = q.Jobs[1:]
		qu.ready[value.Queue] = q
	}
	return value, value != nil
}

func (qu *QueueMap) Delete(jobId int) bool {
	qu.Lock()
	defer qu.Unlock()

	for _, q := range qu.ready {
		for i, j := range q.Jobs {
			if j.Id == jobId {
				q.Jobs = append(q.Jobs[:i], q.Jobs[i+1:]...)
				qu.ready[j.Queue] = q
				return true
			}
		}
	}
	if _, ok := qu.inProgress[jobId]; ok {
		delete(qu.inProgress, jobId)
		return true
	}

	return false
}

func (qu *QueueMap) RetrieveChanneled(keys []string, clientId string) (j *Job) {
	for {
		j, ok := qu.Retrieve(keys)
		if ok {
			return j
		}
	}
}

func (qu *QueueMap) Store(key string, value *Job) int {
	qu.Lock()
	defer qu.Unlock()

	if value.Id == 0 {
		value.Id = qu.maxId
		qu.maxId++
	}

	q := qu.ready[key]
	q.Jobs = append(q.Jobs, value)

	sort.Slice(q.Jobs, func(i, j int) bool {
		return *q.Jobs[i].Priority > *q.Jobs[j].Priority
	})

	qu.ready[key] = q
	return value.Id
}

func (qu *QueueMap) AbortJobs(jobs []ClientJob) (t bool) {
	qu.Lock()
	defer qu.Unlock()

	for _, job := range jobs {
		j, ok := qu.inProgress[job.jobId]
		if ok {
			go qu.Store(j.Queue, j)
			delete(qu.inProgress, j.Id)
			t = true
		}
	}
	return t
}
