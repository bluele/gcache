package timewheel

import (
	"time"
)

type Job struct {
	Id   interface{}
	Time int64
}

type request struct {
	action int
	job    *Job
}

func New(reqLen, dispatchLen int) *TimeWheel {
	me := &TimeWheel{
		jobsByTime:   map[int64]map[interface{}]*Job{},
		jobsById:     map[interface{}]*Job{},
		reqs:         make(chan *request, reqLen),
		dispatchJobs: make(chan *Job, dispatchLen),
	}
	go me.mainloop()
	go me.dispatch()
	return me
}

type Callback func(*Job)

type TimeWheel struct {
	jobsByTime   map[int64]map[interface{}]*Job
	jobsById     map[interface{}]*Job
	reqs         chan *request
	dispatchJobs chan *Job
	callback     Callback
}

func (me *TimeWheel) dispatch() {
	for job := range me.dispatchJobs {
		me.callback(job)
	}
}

func (me *TimeWheel) SetCallback(c Callback) {
	me.callback = c
}

func (me *TimeWheel) Add(job *Job) {
	me.reqs <- &request{
		job: job,
	}
}

func (me *TimeWheel) Delete(id interface{}) {
	me.reqs <- &request{
		action: 1,
		job: &Job{
			Id: id,
		},
	}
}

func (me *TimeWheel) Purge() {
	me.reqs <- &request{
		action: 2,
	}
}

func (me *TimeWheel) mainloop() {
	tk := time.NewTicker(time.Millisecond * 600)
	lastTime := time.Now().Unix()
	expiredJobTimes := map[int64]struct{}{}
	var pendingJobs []*Job
	for {
		tick := tk.C
		var dispatchJobs chan *Job
		var nextJob *Job
		if len(pendingJobs) > 0 {
			// Keep receiving requests even when a callback blocks dispatch.
			// Finish this batch before collecting another tick's expirations.
			tick = nil
			dispatchJobs = me.dispatchJobs
			nextJob = pendingJobs[0]
		}
		select {
		case now := <-tick:
			for jobTime := range expiredJobTimes {
				for _, job := range me.jobsByTime[jobTime] {
					delete(me.jobsById, job.Id)
					pendingJobs = append(pendingJobs, job)
				}
				delete(me.jobsByTime, jobTime)
			}
			expiredJobTimes = map[int64]struct{}{}
			nowTs := now.Unix()
			for jobTime := lastTime + 1; jobTime <= nowTs; jobTime++ {
				for _, job := range me.jobsByTime[jobTime] {
					delete(me.jobsById, job.Id)
					pendingJobs = append(pendingJobs, job)
				}
				delete(me.jobsByTime, jobTime)
			}
			if now.Unix() > lastTime {
				lastTime = nowTs
			}
		case dispatchJobs <- nextJob:
			pendingJobs[0] = nil
			pendingJobs = pendingJobs[1:]
		case req := <-me.reqs:
			if req.action == 0 {
				if job, ok := me.jobsById[req.job.Id]; ok {
					delete(me.jobsByTime[job.Time], req.job.Id)
				}
				me.jobsById[req.job.Id] = req.job
				if _, ok := me.jobsByTime[req.job.Time]; !ok {
					me.jobsByTime[req.job.Time] = map[interface{}]*Job{}
				}
				me.jobsByTime[req.job.Time][req.job.Id] = req.job
				if req.job.Time <= lastTime {
					expiredJobTimes[req.job.Time] = struct{}{}
				}
			} else if req.action == 1 {
				if job, ok := me.jobsById[req.job.Id]; ok {
					delete(me.jobsById, req.job.Id)
					delete(me.jobsByTime[job.Time], req.job.Id)
				}
			} else if req.action == 2 {
				me.jobsByTime = map[int64]map[interface{}]*Job{}
				me.jobsById = map[interface{}]*Job{}
				expiredJobTimes = map[int64]struct{}{}
				pendingJobs = nil

			}
		}
	}
}
