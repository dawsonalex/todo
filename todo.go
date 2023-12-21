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

func ToggleDone() WritableProperty {
	return func(item Item) Item {
		item.Done = !item.Done
		return item
	}
}

func Done() WritableProperty {
	return func(item Item) Item {
		item.Done = true
		return item
	}
}

func NotDone() WritableProperty {
	return func(item Item) Item {
		item.Done = false
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
	// todo: this could do with indexing based on message and date at some point
	list itemList
}

func NewListFromDeserialised(serialisedList string) *List {
	if len(serialisedList) == 0 {
		return &List{
			list: make(itemList, 0),
		}
	}

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
	l.list = append(l.list, &item)
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

func (l *List) GetAll() []Item {
	items := make([]Item, len(l.list))
	for i, item := range l.list {
		items[i] = *item
	}
	return items
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
