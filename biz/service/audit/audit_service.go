package audit

import (
	"encoding/json"
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// auditBufferSize bounds in-flight audit entries. Audit is best-effort: when the
// buffer is full, entries are dropped (with a log) rather than blocking request
// handling or spawning unbounded goroutines.
const auditBufferSize = 1024

type AuditService struct {
	auditDAO *db.AuditLogDAO
	logCh    chan *po.AuditLog
	done     chan struct{}
}

var AuditSvc *AuditService

func InitAuditService() {
	AuditSvc = &AuditService{
		auditDAO: db.NewAuditLogDAO(),
		logCh:    make(chan *po.AuditLog, auditBufferSize),
		done:     make(chan struct{}),
	}
	go AuditSvc.writeLoop()
}

// writeLoop is the single background writer that persists audit entries,
// replacing the prior "one goroutine per Log() call" pattern.
func (s *AuditService) writeLoop() {
	for entry := range s.logCh {
		if err := s.auditDAO.Create(entry); err != nil {
			log.Printf("[audit] failed to persist entry (action=%s target=%s): %v", entry.Action, entry.Target, err)
		}
	}
	close(s.done)
}

// Stop closes the channel and waits (up to a second) for the writer to drain,
// so pending entries are flushed on graceful shutdown.
func (s *AuditService) Stop() {
	if s == nil || s.logCh == nil {
		return
	}
	close(s.logCh)
	select {
	case <-s.done:
	case <-time.After(time.Second):
		log.Println("[audit] stop timed out, some entries may be lost")
	}
}

// Log records an audit log entry asynchronously (best-effort).
func (s *AuditService) Log(c *app.RequestContext, action, target string, details interface{}) {
	ip := ""
	ua := ""
	if c != nil {
		ip = c.ClientIP()
		ua = string(c.UserAgent())
	}

	detailsJSON, _ := json.Marshal(details)
	entry := &po.AuditLog{
		Action:    action,
		Target:    target,
		Operator:  "system", // TODO: Replace with actual user when auth is implemented
		Details:   string(detailsJSON),
		IPAddress: ip,
		UserAgent: ua,
	}

	select {
	case s.logCh <- entry:
	default:
		log.Printf("[audit] buffer full, dropping entry action=%s target=%s", action, target)
	}
}
