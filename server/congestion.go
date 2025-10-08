package server

import (
	"context"
	"math"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/itsabgr/ge"
)

type congestionMap struct {
	mutex sync.Mutex
	total int64
	limit int
	seq   int64
	mapp  map[common.Address]map[int64]func(bool)
	count atomic.Int64
}

func (cm *congestionMap) Count() int {
	return int(cm.count.Load())
}

func (cm *congestionMap) Init(total int, limit int) {
	cm.total = int64(total)
	cm.limit = limit
	cm.seq = 0
	cm.mapp = make(map[common.Address]map[int64]func(bool))
}

func (cm *congestionMap) Push(key common.Address) (context.Context, context.CancelFunc) {
	if cm.count.Add(1) > cm.total {
		cm.count.Add(-1)
		return nil, nil
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.seq++
	id := cm.seq
	ge.Assert(id > 0)
	fm := cm.mapp[key]
	if fm == nil {
		mapp := make(map[int64]func(bool), cm.limit)
		cm.mapp[key] = mapp
		fm = mapp
	}

	ctx, cancelCtx := context.WithCancel(context.Background())

	cancelInternal := func(lock bool) {
		cancelCtx()
		cm.delete(lock, key, id)
	}

	if lastCancelInternal := cm.push(fm, cm.limit, id, cancelInternal); lastCancelInternal != nil {
		lastCancelInternal(false)
	}

	return ctx, func() { cancelInternal(true) }

}

func (cm *congestionMap) delete(lock bool, key common.Address, id int64) {
	if lock {
		cm.mutex.Lock()
		defer cm.mutex.Unlock()
	}
	if fm := cm.mapp[key]; fm != nil {
		if fm[id] != nil {
			cm.count.Add(-1)
			delete(fm, id)
			if len(fm) == 0 {
				delete(cm.mapp, key)
			}
		}

	}
}

func (cm *congestionMap) push(fm map[int64]func(bool), limit int, id int64, val func(bool)) func(bool) {
	if len(fm) < limit {
		fm[id] = val
		return nil
	}
	var last int64 = math.MaxInt64
	for key := range fm {
		last = min(last, key)
	}
	item := fm[last]
	ge.Assert(item != nil)
	delete(fm, last)
	fm[id] = val
	return item
}
