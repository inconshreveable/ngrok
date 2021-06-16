// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The implementation borrows heavily from SmallLRUCache (originally by Nathan
// Schrenk). The object maintains a doubly-linked list of elements in the
// When an element is accessed it is promoted to the head of the list, and when
// space is needed the element at the tail of the list (the least recently used
// element) is evicted.
package cache

import (
	"container/list"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type LRUCache struct {
	mu sync.Mutex

	// list & table of *entry objects
	list  *list.List
	table map[string]*list.Element

	// Our current size, in bytes. Obviously a gross simplification and low-grade
	// approximation.
	size uint64

	// How many bytes we are limiting the cache to.
	capacity uint64
}

// Values that go into LRUCache need to satisfy this interface.
type Value interface {
	Size() int
}

type Item struct {
	Key   string
	Value Value
}

type entry struct {
	key           string
	value         Value
	size          int
	time_accessed time.Time
}

func NewLRUCache(capacity uint64) *LRUCache {
	return &LRUCache{
		list:     list.New(),
		table:    make(map[string]*list.Element),
		capacity: capacity,
	}
}

func (lru *LRUCache) Get(key string) (v Value, ok bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return nil, false
	}
	lru.moveToFront(element)
	return element.Value.(*entry).value, true
}

func (lru *LRUCache) Set(key string, value Value) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element := lru.table[key]; element != nil {
		lru.updateInplace(element, value)
	} else {
		lru.addNew(key, value)
	}
}

func (lru *LRUCache) SetIfAbsent(key string, value Value) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element := lru.table[key]; element != nil {
		lru.moveToFront(element)
	} else {
		lru.addNew(key, value)
	}
}

func (lru *LRUCache) Delete(key string) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return false
	}

	lru.list.Remove(element)
	delete(lru.table, key)
	lru.size -= uint64(element.Value.(*entry).size)
	return true
}

func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.list.Init()
	lru.table = make(map[string]*list.Element)
	lru.size = 0
}

func (lru *LRUCache) SetCapacity(capacity uint64) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.capacity = capacity
	lru.checkCapacity()
}

func (lru *LRUCache) Stats() (length, size, capacity uint64, oldest time.Time) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	if lastElem := lru.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*entry).time_accessed
	}
	return uint64(lru.list.Len()), lru.size, lru.capacity, oldest
}

func (lru *LRUCache) StatsJSON() string {
	if lru == nil {
		return "{}"
	}
	l, s, c, o := lru.Stats()
	return fmt.Sprintf("{\"Length\": %v, \"Size\": %v, \"Capacity\": %v, \"OldestAccess\": \"%v\"}", l, s, c, o)
}

func (lru *LRUCache) Keys() []string {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	keys := make([]string, 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry).key)
	}
	return keys
}

func (lru *LRUCache) Items() []Item {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	items := make([]Item, 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		v := e.Value.(*entry)
		items = append(items, Item{Key: v.key, Value: v.value})
	}
	return items
}

func (lru *LRUCache) SaveItems(w io.Writer) error {
	items := lru.Items()
	encoder := gob.NewEncoder(w)
	return encoder.Encode(items)
}

func (lru *LRUCache) SaveItemsToFile(path string) error {
	if wr, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
		return err
	} else {
		defer wr.Close()
		return lru.SaveItems(wr)
	}
}

func (lru *LRUCache) LoadItems(r io.Reader) error {
	items := make([]Item, 0)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&items); err != nil {
		return err
	}

	lru.mu.Lock()
	defer lru.mu.Unlock()
	for _, item := range items {
		// XXX: copied from Set()
		if element := lru.table[item.Key]; element != nil {
			lru.updateInplace(element, item.Value)
		} else {
			lru.addNew(item.Key, item.Value)
		}
	}

	return nil
}

func (lru *LRUCache) LoadItemsFromFile(path string) error {
	if rd, err := os.Open(path); err != nil {
		return err
	} else {
		defer rd.Close()
		return lru.LoadItems(rd)
	}
}

func (lru *LRUCache) updateInplace(element *list.Element, value Value) {
	valueSize := value.Size()
	sizeDiff := valueSize - element.Value.(*entry).size
	element.Value.(*entry).value = value
	element.Value.(*entry).size = valueSize
	lru.size += uint64(sizeDiff)
	lru.moveToFront(element)
	lru.checkCapacity()
}

func (lru *LRUCache) moveToFront(element *list.Element) {
	lru.list.MoveToFront(element)
	element.Value.(*entry).time_accessed = time.Now()
}

func (lru *LRUCache) addNew(key string, value Value) {
	newEntry := &entry{key, value, value.Size(), time.Now()}
	element := lru.list.PushFront(newEntry)
	lru.table[key] = element
	lru.size += uint64(newEntry.size)
	lru.checkCapacity()
}

func (lru *LRUCache) checkCapacity() {
	// Partially duplicated from Delete
	for lru.size > lru.capacity {
		delElem := lru.list.Back()
		delValue := delElem.Value.(*entry)
		lru.list.Remove(delElem)
		delete(lru.table, delValue.key)
		lru.size -= uint64(delValue.size)
	}
}
