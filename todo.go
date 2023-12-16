package todo

import (
	"fmt"
	"strconv"
	"strings"
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

type Id int

type Item struct {
	Id        Id        `json:"-"`
	Message   string    `json:"message"`
	Done      bool      `json:"done"`
	Timestamp time.Time `json:"created"`
}

type itemList []*Item

type List struct {
	sync.RWMutex
	list itemList
}

func NewListFromDeserialised(serialisedList string) *List {
	serialisedItems := strings.Split(serialisedList, "\n")
	list := List{
		list: make(itemList, len(serialisedItems)),
	}
	for i, serialisedItem := range serialisedItems {
		itemParts := strings.Split(serialisedItem, "|")
		done, _ := strconv.ParseBool(itemParts[1])
		timestamp, _ := time.Parse(time.RFC3339, itemParts[2])
		list.list[i] = &Item{
			Id:        Id(i),
			Message:   itemParts[0],
			Done:      done,
			Timestamp: timestamp,
		}
	}
	return &list
}

// SerialiseToString returns the string serialisation of a list.
// This contains each item in the list in the format:
// message|done|timestamp
// each item is separated by a newline.
func (l *List) SerialiseToString() string {
	var list string
	for _, item := range l.list {
		itemString := fmt.Sprintf("%s|%t|%s\n", item.Message, item.Done, item.Timestamp.Format(time.RFC3339))
		list = list + itemString
	}
	return list
}

func (l *List) Add(properties ...WritableProperty) Item {
	l.Lock()
	defer l.Unlock()

	item := Item{
		Id: Id(len(l.list)),
	}
	for _, property := range properties {
		item = property(item)
	}
	l.list[len(l.list)] = &item
	return item
}

func (l *List) Remove(id Id) {
	l.Lock()
	defer l.Unlock()

	l.list = append(l.list[:id], l.list[id+1:]...)
}

func (l *List) Get(id Id) (Item, bool) {
	item, exists := l.get(id)
	return *item, exists
}

func (l *List) get(itemId Id) (*Item, bool) {
	l.RLock()
	defer l.RUnlock()

	id := int(itemId)
	if id < 0 || id > len(l.list) {
		return nil, false
	}

	item := l.list[id]
	return item, true
}

func (l *List) Update(id Id, props ...WritableProperty) (Item, bool) {
	l.Lock()
	defer l.Unlock()

	item, exists := l.get(id)
	if !exists {
		return Item{}, false
	}
	updatedItem := *item
	for _, prop := range props {
		updatedItem = prop(*item)
	}
	l.list[id] = &updatedItem
	return updatedItem, true
}
