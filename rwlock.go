package rwlock

import (
	"github.com/hexycc/rwlock/client"
	"github.com/hexycc/rwlock/tool"
)

type RWLock struct {
	shaHashID *string
	lockKey   string
	uniqID    string
	expire    int64
}

func New(key string) *RWLock {
	return &RWLock{
		lockKey: key,
		uniqID:  tool.GetUUID(),
		expire:  10,
	}
}

func (l *RWLock) Lock() {
	client.Lock(l.lockKey, l.uniqID, l.expire)
}

func (l *RWLock) Unlock() {
	client.Unlock(l.lockKey, l.uniqID)
}

func (l *RWLock) RLock() {
	client.RLock(l.lockKey)
}

func (l *RWLock) RUnlock() {
	client.RUnlock(l.lockKey)
}
