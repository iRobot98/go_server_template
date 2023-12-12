package base

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type CounterStructPointer = *CounterStruct

type CounterStruct struct {
	count atomic.Int32
}

const TwentyMicroSecond time.Duration = 20 * time.Microsecond

const TwoSeconds = 2 * time.Second

func (c CounterStructPointer) Increment() int {
	return int(c.count.Add(1))

}
func (c CounterStructPointer) Decrement() int {
	return int(c.count.Add(-1))
}

func (c CounterStructPointer) CurrentCount() int {
	return int(c.count.Load())
}

// potential race condition
// thread b reads 0 adds 1
// thread a reads 3 sets 0
// thread c reads 2 sets 3
// thread d reads 1 sets 2
func (c CounterStructPointer) Reset() int {
	c.count.Store(0)
	return int(c.count.Load())
}

type CustomDuration struct {
	Difference time.Duration `json:"difference"`
	started    bool
	stopped    bool
	Start_     time.Time `json:"start"`
	End_       time.Time `json:"end"`
}

func NewCustomDuration() CustomDuration {
	return CustomDuration{}
}

func (c *CustomDuration) Start() time.Time {
	if !c.started {
		c.Start_ = time.Now()
		c.started = true
	}
	return c.Start_
}
func (c *CustomDuration) Stop() time.Time {
	if !c.stopped {
		c.End_ = time.Now()
		c.Difference = c.End_.Sub(c.Start_)
	}
	return c.End_
}

func (c *CustomDuration) Ping() time.Duration {
	return time.Since(c.Start_)
}

func FormatTime(t *time.Time) string {
	return t.Format("2-Jan-2006 15:04:05")
}
func (c *CustomDuration) String() string {

	//                 	duration:        duration:
	return fmt.Sprintf("start: 	%s\nend: 	   %ss\nduration: %dms\n",
		FormatTime(&c.Start_), FormatTime(&c.End_), c.Difference.Milliseconds())
}

type MutexedMap[T any] struct {
	sync.RWMutex
	Map map[string]T `json:"Map"`
	// wg  sync.WaitGroup
}

func (m *MutexedMap[T]) Init() {
	m.Lock()
	if m.Map == nil {
		m.Map = make(map[string]T)
	}
	m.Unlock()
}

func (m *MutexedMap[T]) Set(key string, val T) error {
	m.Lock()
	m.Map[key] = val
	m.Unlock()

	return nil
}

func (m *MutexedMap[T]) Get(key string) (val T, ok bool) {
	m.RLock()
	val, ok = m.Map[key]
	m.RUnlock()
	return
}

func (m *MutexedMap[T]) Len() (l int) {
	m.RLock()
	l = len(m.Map)
	m.RUnlock()
	return
}

func (m *MutexedMap[T]) Keys() (keys []string) {
	m.RLock()
	for key := range m.Map {
		keys = append(keys, key)
	}
	m.RUnlock()

	return
}

// Removes a key from the map
// returns a copy of the deleted value
// (in case of clean operations)
func (m *MutexedMap[T]) Remove(key string) T {
	m.Lock()
	val, ok := m.Map[key]
	if ok {
		delete(m.Map, key)
	}
	m.Unlock()

	return val
}
