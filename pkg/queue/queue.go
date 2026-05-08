package queue

import "time"

type SyncRequest struct {
	MirrorID    uint
	TriggerType string
	RequestedAt time.Time
}

type UniqueQueue interface {
	Push(req SyncRequest) error
	Pop() (SyncRequest, bool)
	Len() int
	Has(mirrorID uint) bool
	Close()
}
