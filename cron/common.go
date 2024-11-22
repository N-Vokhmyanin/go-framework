package cron

import "time"

type nextRun struct {
	nextFunc func(t time.Time) time.Time
}

var _ Schedule = (*nextRun)(nil)

//goland:noinspection GoUnusedExportedFunction
func Next(f func(t time.Time) time.Time) Schedule {
	return &nextRun{f}
}

func (n *nextRun) Next(t time.Time) time.Time {
	return n.nextFunc(t)
}
