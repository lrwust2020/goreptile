package tool

import (
	"errors"
	"fmt"
	sched "github.com/fmyxyz/goreptile/scheduler"
	"runtime"
	"time"
)

func Monitoring(
	scheduler sched.Scheduler,
	intervalNs time.Duration,
	maxIdleCount uint,
	autoStop bool,
	detailSummary bool,
	record Record) <-chan uint64 {
	if scheduler == nil {
		panic(errors.New("The Scheduler is invalid!"))
	}
	if intervalNs < time.Millisecond {
		intervalNs = time.Millisecond
	}
	if maxIdleCount < 1000 {
		maxIdleCount = 1000
	}
	stopNotifier := make(chan byte, 1)

	reportError(scheduler, record, stopNotifier)

	recordSummary(scheduler, detailSummary, record, stopNotifier)

	checkCountChan := make(chan uint64, 2)

	checkStatus(
		scheduler,
		intervalNs,
		maxIdleCount,
		autoStop,
		checkCountChan,
		record,
		stopNotifier)

	return checkCountChan
}
func checkStatus(
	scheduler sched.Scheduler,
	intervalNs time.Duration,
	maxIdleCount uint,
	autoStop bool,
	checkCountChan chan<- uint64,
	record Record,
	stopNotifier chan<- byte,
) {
	var checkCount uint64
	go func() {
		defer func() {
			stopNotifier <- 1
			stopNotifier <- 2
			checkCountChan <- checkCount
		}()
		waitForSchedulerStart(scheduler)
		var idleCount uint
		var firstIdleTime time.Time
		for {
			if scheduler.Idle() {
				idleCount++
				if idleCount == 1 {
					firstIdleTime = time.Now()
				}
				if idleCount > maxIdleCount {
					msg := fmt.Sprintf("Idle count reached the max idle count since %s\n", time.Since(firstIdleTime).String())
					record(0, msg)
					if scheduler.Idle() {
						if autoStop {
							var result string
							if scheduler.Stop() {
								result = "success"
							} else {
								result = "failing"
							}
							msg = fmt.Sprintf("Stop Scheduler:%s\n", result)
							record(0, msg)
						}
						break
					} else {
						if idleCount > 0 {
							idleCount = 0
						}
					}
				}
			} else {
				if idleCount > 0 {
					idleCount = 0
				}
			}
			checkCount++
			time.Sleep(intervalNs)
		}
	}()
}
func reportError(
	scheduler sched.Scheduler,
	record Record,
	stopNotifier <-chan byte) {
	go func() {

		waitForSchedulerStart(scheduler)
		for {
			select {
			case <-stopNotifier:
				return
			default:
			}
			errorChan := scheduler.ErrorChan()
			if errorChan == nil {
				return
			}
			err := <-errorChan

			if err != nil {
				errMsg := fmt.Sprintf("Error (received from error channel):%s", err)

				record(2, errMsg)
			}
			time.Sleep(time.Millisecond)
		}
	}()
}
func waitForSchedulerStart(scheduler sched.Scheduler) {
	for !scheduler.Running() {
		time.Sleep(time.Millisecond)
	}
}
func recordSummary(
	scheduler sched.Scheduler,
	detailSummary bool,
	record Record,
	stopNotifier <-chan byte) {
	go func() {
		waitForSchedulerStart(scheduler)

		var recordCount uint64 = 1
		var startTime = time.Now()
		var prevSchedSummary sched.SchedSummary
		var prevNumGoroutine int
		for {
			select {
			case <-stopNotifier:
				return
			default:
			}
			var currentNumGoroutine = runtime.NumGoroutine()
			var currentSchedSummary = scheduler.Summary("	")

			if currentNumGoroutine != prevNumGoroutine ||
				!currentSchedSummary.Same(prevSchedSummary) {
				var schedSummaryStr = func() string {
					if detailSummary {
						return currentSchedSummary.Detail()
					} else {
						return currentSchedSummary.String()
					}
				}()

				var info = fmt.Sprintf(summaryForMonitoring, recordCount, currentNumGoroutine, schedSummaryStr, time.Since(startTime).String())

				record(0, info)
				prevSchedSummary = currentSchedSummary
				prevNumGoroutine = currentNumGoroutine
				recordCount++

			}
			time.Sleep(time.Millisecond)
		}
	}()
}

var summaryForMonitoring = `Monitor - Collected information[%d]:\n
	Goroutine Number:%d
	Scheduler:%s
	Escaped time:%s
`

type Record func(level byte, content string)
