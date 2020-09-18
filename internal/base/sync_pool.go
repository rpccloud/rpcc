package base

import (
	"sync"
)

// SyncPoolDebug should only works on debug mode, when release it,
// please replace it with sync.Pool
// type SyncPool = SyncPoolDebug
type SyncPool = sync.Pool

var (
	safePoolDebugMutex = sync.Mutex{}
	safePoolDebugMap   = map[interface{}]bool{}
)

type SyncPoolDebug struct {
	pool sync.Pool
	New  func() interface{}
}

func (p *SyncPoolDebug) Put(x interface{}) {
	safePoolDebugMutex.Lock()
	defer safePoolDebugMutex.Unlock()

	if _, ok := safePoolDebugMap[x]; ok {
		delete(safePoolDebugMap, x)
		p.pool.Put(x)
	} else {
		panic("SyncPoolDebug Put check failed")
	}
}

func (p *SyncPoolDebug) Get() interface{} {
	safePoolDebugMutex.Lock()
	defer safePoolDebugMutex.Unlock()

	if p.pool.New == nil {
		p.pool.New = p.New
		Log("Warn: SyncPool is in debug mode, which may slow down the program")
	}

	x := p.pool.Get()

	if _, ok := safePoolDebugMap[x]; !ok {
		safePoolDebugMap[x] = true
	} else {
		panic("SyncPoolDebug Get check failed")
	}

	return x
}
