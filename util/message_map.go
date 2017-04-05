package util

import (
	"sync"
	"time"
)

// MessageMap will syncronize multireads and writes of message times
type MessageMap struct {
	Items map[string]time.Time
	sync.RWMutex
}

// NewMessageMap returns a fresh message map
func NewMessageMap() *MessageMap {
	return &MessageMap{map[string]time.Time{}, sync.RWMutex{}}
}

// Add will insert an item into the message map
func (mm *MessageMap) Add(guid string, t time.Time) {
	mm.Lock()
	defer mm.Unlock()
	mm.Items[guid] = t
}

// Get will get the key and pass a bool back if it was found
func (mm *MessageMap) Get(key string) (time.Time, bool) {
	mm.RLock()
	defer mm.RUnlock()
	val, ok := mm.Items[key]
	return val, ok
}

// Delete reoves key if can be found, otherwise returns false
func (mm *MessageMap) Delete(key string) (time.Time, bool) {
	mm.Lock()
	defer mm.Unlock()
	start, ok := mm.Items[key]
	if ok {
		delete(mm.Items, key)
		return start, true
	}
	return time.Time{}, false
}

// Overdue will return any messages with durations longer than cutoff in seconds
func (mm *MessageMap) Overdue(cutoffSeconds float64) map[string]float64 {
	overdue := map[string]float64{}
	mm.RLock()
	defer mm.RUnlock()
	now := time.Now()
	for id, start := range mm.Items {
		duration := now.Sub(start).Seconds()
		if duration >= cutoffSeconds {
			overdue[id] = duration
		}
	}
	return overdue
}

// Empty will reset the contents of the message map
func (mm *MessageMap) Empty() {
	mm.Lock()
	defer mm.Unlock()
	mm.Items = map[string]time.Time{}
}
