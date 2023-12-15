package todo

import (
	"sync"
	"time"
)

type WritableProperty func(item Item) Item

func Message(message string) WritableProperty {
	return func(item Item) Item {
		item.Message = message
		return item
	}
}

func DoneStatus(status bool) WritableProperty {
	return func(item Item) Item {
		item.Done = status
		return item
	}
}

type Item struct {
	Id        Id        `json:"id"`
	Message   string    `json:"message"`
	Done      bool      `json:"done"`
	Timestamp time.Time `json:"created"`
}

type List struct {
	ids idCache
	sync.RWMutex
	list map[Id]*Item
}

func NewList() *List {
	return &List{
		list: map[Id]*Item{},
	}
}

func (l *List) Add(properties ...WritableProperty) Item {
	l.Lock()
	defer l.Unlock()

	item := Item{}
	for _, property := range properties {
		item = property(item)
	}
	l.list[l.ids.next()] = &item
	return item
}

func (l *List) Remove(id Id) {
	l.Lock()
	defer l.Unlock()

	delete(l.list, id)
}

func (l *List) Get(id Id) (Item, bool) {
	item, exists := l.get(id)
	return *item, exists
}

func (l *List) get(id Id) (*Item, bool) {
	l.RLock()
	defer l.RUnlock()

	item, exists := l.list[id]
	return item, exists
}

func (l *List) Update(id Id, props ...WritableProperty) Item {
	l.Lock()
	defer l.Unlock()

	item, _ := l.get(id)
	updatedItem := *item
	for _, prop := range props {
		updatedItem = prop(*item)
	}
	l.list[id] = &updatedItem
	return updatedItem
}
