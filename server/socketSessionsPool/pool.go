package socketSessionsPool

import (
	"sync"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

type SocketSessionsPool struct {
	sync.Mutex

	sessions []sockjs.Session
}

func New() *SocketSessionsPool {
	return &SocketSessionsPool{
		sessions: make([]sockjs.Session, 0),
	}
}

func (p *SocketSessionsPool) Add(s sockjs.Session) {
	p.Lock()
	defer p.Unlock()

	p.sessions = append(p.sessions, s)
}

func (p *SocketSessionsPool) Send(payload string) {
	p.Lock()
	defer p.Unlock()

	sessionsToRemove := make([]sockjs.Session, 0)

	for _, s := range p.sessions {
		if err := s.Send(payload); err != nil {
			s.Close(1, "failed to send")
			sessionsToRemove = append(sessionsToRemove, s)
		}
	}

	for _, s := range sessionsToRemove {
		for i, ss := range p.sessions {
			if ss == s {
				p.sessions = append(p.sessions[:i], p.sessions[i+1:]...)
				break
			}
		}
	}
}
