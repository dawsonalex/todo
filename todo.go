package todo

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Priority rune

func (p Priority) Valid() bool {
	return p >= 'A' && p <= 'Z'
}

type Id int

type Item struct {
	Message       string            `json:"description"`
	Done          bool              `json:"done"`
	Priority      Priority          `json:"priority"`
	CreatedDate   time.Time         `json:"created-date"`
	CompletedDate time.Time         `json:"completed-date"`
	Projects      []string          `json:"projects"`
	Contexts      []string          `json:"contexts"`
	SpecialKeys   map[string]string `json:"special-keys"` // Not used, but should be preserved for other tools.
}

func (i *Item) MarshalText() (text []byte, err error) {
	var parts []string

	if i.Done {
		parts = append(parts, "x")
	}

	if i.Priority.Valid() {
		parts = append(parts, fmt.Sprintf("(%c)", rune(i.Priority)))
	}

	if !i.CreatedDate.IsZero() {
		if !i.CompletedDate.IsZero() {
			parts = append(parts, i.CompletedDate.Format("2006-01-02"))
		}
		parts = append(parts, i.CreatedDate.Format("2006-01-02"))
	}

	parts = append(parts, i.Message)

	return []byte(strings.Join(parts, " ")), nil
}

func (i *Item) UnmarshalText(text []byte) error {
	textString := string(text)
	if len(textString) == 0 {
		return errors.New("no item found in text")
	}

	nextPosition := 0

	// lowercase x indicates item is done
	if textString[nextPosition] == 'x' {
		i.Done = true
		nextPosition += 2
	}

	// optional priority: "(A) "
	if nextPosition < len(textString) && textString[nextPosition] == '(' {
		i.Priority = Priority(text[nextPosition+1])
		nextPosition += 4 // skip "(A) "
	}

	// optional completion and/or creation date
	if nextPosition+10 <= len(textString) {
		if firstDate, err := time.ParseInLocation("2006-01-02", textString[nextPosition:nextPosition+10], time.Local); err == nil {
			nextPosition += 11

			if nextPosition+10 <= len(textString) {
				if createdDate, err := time.ParseInLocation("2006-01-02", textString[nextPosition:nextPosition+10], time.Local); err == nil {
					i.CreatedDate = createdDate
					i.CompletedDate = firstDate
					nextPosition += 11
				} else {
					i.CreatedDate = firstDate
				}
			} else {
				i.CreatedDate = firstDate
			}
		}
	}

	i.Message = textString[nextPosition:]
	i.Projects, i.Contexts, i.SpecialKeys = parseMessage(textString[nextPosition:])

	return nil
}

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

// List is a list of todo items. List is safe for concurrent use.
type List struct {
	sync.RWMutex
	list itemList
}

func (l *List) Add(item Item) Item {
	l.Lock()
	defer l.Unlock()

	l.list = append(l.list, &item)
	return item
}

func (l *List) Remove(id Id) {
	l.Lock()
	defer l.Unlock()

	l.list = append(l.list[:id], l.list[id+1:]...)
}

func (l *List) Get(id Id) (Item, bool) {
	l.RLock()
	defer l.RUnlock()

	idx := int(id)
	if idx < 0 || idx >= len(l.list) {
		return Item{}, false
	}
	return *l.list[idx], true
}

func (l *List) GetAll() []Item {
	l.RLock()
	defer l.RUnlock()

	items := make([]Item, len(l.list))
	for i, item := range l.list {
		items[i] = *item
	}
	return items
}

// ReadFile reads a todo.txt file and returns a List.
// Returns an empty list if the file does not exist.
func ReadFile(path string) (*List, error) {
	path = filepath.Clean(path)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return &List{list: make(itemList, 0)}, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	list := &List{list: make(itemList, 0)}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item Item
		if err := item.UnmarshalText([]byte(line)); err != nil {
			continue
		}
		list.list = append(list.list, &item)
	}
	return list, scanner.Err()
}

// WriteFile writes all items in the list to path in todo.txt format.
// The write is atomic: a temp file is written then renamed into place.
func WriteFile(path string, list *List) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}

	tmpPath := filepath.Clean(path + ".tmp")
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	list.RLock()
	w := bufio.NewWriter(f)
	writeErr := func() error {
		for _, item := range list.list {
			text, err := item.MarshalText()
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "%s\n", text); err != nil {
				return err
			}
		}
		return w.Flush()
	}()
	list.RUnlock()

	if closeErr := f.Close(); closeErr != nil && writeErr == nil {
		writeErr = closeErr
	}
	if writeErr != nil {
		_ = os.Remove(tmpPath)
		return writeErr
	}
	return os.Rename(tmpPath, path)
}
