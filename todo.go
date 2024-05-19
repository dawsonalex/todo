package todo

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type WritableProperty func(item Item) Item

func Message(message string) WritableProperty {
	return func(item Item) Item {
		item.Description = message
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
		return item // comment thingkjaflkjdsflkjlaksdjflkjadsf
	}
}

func NotDone() WritableProperty {
	return func(item Item) Item {
		item.Done = false
		return item
	}
}

type Id int

type Priority rune

func (p Priority) Valid() bool {
	return p > 'A' && p < 'Z'
}

type Item struct {
	Description   string            `json:"description"`
	Done          bool              `json:"done"`
	Priority      Priority          `json:"priority"`
	CreatedDate   time.Time         `json:"created-date"`
	CompletedDate time.Time         `json:"completed-date"`
	Projects      []string          `json:"projects"`
	Contexts      []string          `json:"contexts"`
	SpecialKeys   map[string]string `json:"special-keys"` // Not uesd, but should be preserved for other tools.
}

func (i *Item) MarshalText() (text []byte, err error) {
	var textStr string
	if i.Done {
		textStr = "x"
	}

	// only output completed date if the created date exists
	if !i.CreatedDate.IsZero() {
		textStr = textStr + i.CompletedDate.Format("2006-01-02") + " " + i.CreatedDate.Format("2006-01-02")
	}

	textStr = textStr + " " + i.Description

	for _, project := range i.Projects {
		textStr = textStr + " +" + project
	}

	for _, context := range i.Contexts {
		textStr = textStr + " @" + context
	}

	for key, value := range i.SpecialKeys {
		textStr = textStr + fmt.Sprintf(" %s:%s", key, value)
	}

	return []byte(textStr), nil
}

func (i *Item) UnmarshalText(text []byte) error {

	// TODO: Parse a line into an item
	textString := string(text)
	if len(textString) == 0 {
		return errors.New("no item found in text")
	}

	nextPosition := 0

	// lowercase x indicates item is done
	if textString[nextPosition] == 'x' {
		i.Done = true

		// next position is characters ahead
		nextPosition += 2
	}

	// optional priority
	if textString[nextPosition] == '(' {
		i.Priority = Priority(text[nextPosition+1])

		// move cursor past (A)
		nextPosition += 3
	}

	// optional completion date
	if firstDate, err := time.ParseInLocation("2006-01-02", textString[nextPosition:nextPosition+10], time.Local); err == nil {
		// move cursor to after first valid date
		nextPosition += 11

		// If the next set of chars is a date, this is the creation date
		if createdDate, err := time.ParseInLocation("2006-01-02", textString[nextPosition:nextPosition+10], time.Local); err == nil {
			i.CreatedDate = createdDate
			i.CompletedDate = firstDate

			nextPosition += 11
		} else {
			// If we only have one date, it's the created date
			i.CreatedDate = firstDate
		}
	}

	// everything else is description with optional tags and context
	i.Description = textString[nextPosition:]

	projects, contexts, specialKeys := parseMessage(textString[nextPosition:])
	i.Projects = projects
	i.Contexts = contexts
	i.SpecialKeys = specialKeys

	return nil
}

// parseMessage parses a
func parseMessage(message string) (projects []string, contexts []string, specialKeys map[string]string) {
	for _, word := range strings.Fields(message) {
		if word[0] == '+' {
			projects = append(projects, word[1:])
			continue
		}

		if word[0] == '@' {
			contexts = append(contexts, word[1:])
			continue
		}

		if pos := strings.Index(word, ":"); pos != -1 {
			if specialKeys == nil {
				specialKeys = make(map[string]string)
			}
			specialKeys[word[:pos]] = word[pos+1:]
		}
	}
	return projects, contexts, specialKeys
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
			//Id:          Id(i),
			Description: itemParts[0],
			Done:        done,
			CreatedDate: timestamp,
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
		itemString := fmt.Sprintf("%s|%t|%s\n", item.Description, item.Done, item.CreatedDate.Format(time.RFC3339))
		list = list + itemString
	}
	return list
}

func (l *List) Add(properties ...WritableProperty) Item {
	l.Lock()
	defer l.Unlock()

	item := Item{
		//Id: Id(len(l.list)),
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
