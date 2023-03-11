package problem9

import "protohackers/util"

type ServerWithReader struct {
	util.ServerTCP
}

func (s *ServerWithReader) handleLine(request *Request, client *Client) Response {

	switch request.Request {
	case "put":
		if request.Priority == nil || request.Queue == "" || request.Job == nil {
			return Response{
				Status: "error",
				Error:  "Missing required parameters",
			}
		} else {
			jobId := queues.Store(request.Queue, &Job{
				Details:  request.Job,
				Priority: request.Priority,
				Queue:    request.Queue,
			})
			return Response{
				Status: "ok",
				Id:     jobId,
			}
		}
	case "get":
		if len(request.Queues) == 0 {
			return Response{
				Status: "error",
				Error:  "Missing required parameters",
			}
		}
		job, ok := queues.Retrieve(request.Queues)
		if !ok {
			if request.Wait {
				job := queues.RetrieveChanneled(request.Queues, client.id)
				client.currentJobs = append(client.currentJobs, ClientJob{
					queue: job.Queue,
					jobId: job.Id,
				})
				return Response{
					Status:   "ok",
					Id:       job.Id,
					Job:      job.Details,
					Priority: job.Priority,
					Queue:    job.Queue,
				}
			} else {
				return Response{
					Status: "no-job",
				}
			}
		}
		client.currentJobs = append(client.currentJobs, ClientJob{
			queue: job.Queue,
			jobId: job.Id,
		})
		return Response{
			Status:   "ok",
			Id:       job.Id,
			Job:      job.Details,
			Priority: job.Priority,
			Queue:    job.Queue,
		}
	case "abort":
		if request.Id == 0 || !client.hasJob(request.Id) {
			return Response{
				Status: "no-job",
			}
		} else {
			job := client.getJob(request.Id)
			ok := queues.AbortJobs([]ClientJob{job})

			if !ok {
				return Response{
					Status: "no-job",
				}
			}
			return Response{
				Status: "ok",
			}
		}

	case "delete":
		if request.Id == 0 {
			return Response{
				Status: "no-job",
			}
		} else {
			ok := queues.Delete(request.Id)
			if !ok {
				return Response{
					Status: "no-job",
				}
			}
			return Response{
				Status: "ok",
			}
		}
	case "debug":
		queues.Dump()
		return Response{
			Status: "ok",
		}
	default:
		return Response{
			Status: "error",
			Error:  "Uknown",
		}
	}
}
