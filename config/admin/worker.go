package admin

import (
	"errors"
	"fmt"
	"time"

	"github.com/qor/media/oss"
	"github.com/qor/worker"
)

type SimpleQueue struct {
	Worker *worker.Worker
}

func (q *SimpleQueue) Add(j worker.QorJobInterface) error {
	return q.Worker.RunJob(j.GetJobID())
}

func (q *SimpleQueue) Run(j worker.QorJobInterface) error {
	job := j.GetJob()

	if job.Handler != nil {
		return job.Handler(j.GetSerializableArgument(j), j)
	}

	return errors.New("SimpleQue:no handler found for job " + job.Name)
}

func (q *SimpleQueue) Kill(j worker.QorJobInterface) error {
	return errors.New("SimpleQue:kill not implemented")
}

func (q *SimpleQueue) Remove(j worker.QorJobInterface) error {
	return errors.New("SimpleQue:remove not implemented")
}

func getWorker() *worker.Worker {
	sq := SimpleQueue{}
	Worker := worker.New(&worker.Config{
		Queue: &sq,
	})
	sq.Worker = Worker

	type sendNewsletterArgument struct {
		Subject      string
		Content      string `sql:"size:65532"`
		SendPassword string
		worker.Schedule
	}

	Worker.RegisterJob(&worker.Job{
		Name: "Send Newsletter",
		Handler: func(argument interface{}, qorJob worker.QorJobInterface) error {
			qorJob.AddLog("Started sending newsletters...")
			qorJob.AddLog(fmt.Sprintf("Argument: %+v", argument.(*sendNewsletterArgument)))
			for i := 1; i <= 100; i++ {
				time.Sleep(100 * time.Millisecond)
				qorJob.AddLog(fmt.Sprintf("Sending newsletter %v...", i))
				qorJob.SetProgress(uint(i))
			}
			qorJob.AddLog("Finished send newsletters")
			return nil
		},
		Resource: Admin.NewResource(&sendNewsletterArgument{}),
	})

	type importProductArgument struct {
		File oss.OSS
	}

	return Worker
}
