package generator

import "time"

type metrics struct {
	total     int
	succeeded int
	failed    int
	genDur    time.Duration // cumulative time spent in buildAuditEvent()
	sendDur   time.Duration // cumulative time spent in sender.send()
	totalDur  time.Duration // wall clock time for the whole Run

}
