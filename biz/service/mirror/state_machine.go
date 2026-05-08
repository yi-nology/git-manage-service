package mirror

import (
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type MirrorStatus struct {
	Current string
}

func NewMirrorStatus(current string) *MirrorStatus {
	return &MirrorStatus{Current: current}
}

var validTransitions = map[string][]string{
	po.MirrorStatusActive:  {po.MirrorStatusSyncing, po.MirrorStatusPaused},
	po.MirrorStatusSyncing: {po.MirrorStatusActive, po.MirrorStatusFailed},
	po.MirrorStatusFailed:  {po.MirrorStatusSyncing, po.MirrorStatusPaused, po.MirrorStatusActive},
	po.MirrorStatusPaused:  {po.MirrorStatusActive},
}

func (s *MirrorStatus) CanTransitionTo(target string) bool {
	allowed, ok := validTransitions[s.Current]
	if !ok {
		return false
	}
	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}

func (s *MirrorStatus) TransitionTo(target string) error {
	if !s.CanTransitionTo(target) {
		return fmt.Errorf("invalid state transition: %s -> %s", s.Current, target)
	}
	s.Current = target
	return nil
}

func CanStartSync(status string) bool {
	return status == po.MirrorStatusActive || status == po.MirrorStatusFailed
}

func IsTerminalStatus(status string) bool {
	return status == po.MirrorStatusPaused
}
