package util

import (
	"container/list"
	"sync"
)

type Ring struct {
	sync.Mutex
	*list.List
	capacity int
}

func NewRing(capacity int) *Ring {
	return &Ring{capacity: capacity, List: list.New()}
}

func (r *Ring) Add(item interface{}) interface{} {
	r.Lock()
	defer r.Unlock()

	// add new item
	r.PushFront(item)

	// remove old item if at capacity
	var old interface{}
	if r.Len() >= r.capacity {
		old = r.Remove(r.Back())
	}

	return old
}

func (r *Ring) Slice() []interface{} {
	r.Lock()
	defer r.Unlock()

	i := 0
	items := make([]interface{}, r.Len())
	for e := r.Front(); e != nil; e = e.Next() {
		items[i] = e.Value
		i++
	}

	return items
}
