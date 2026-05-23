package todo

// TODO: This would be a good candidate for fuzz testing

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestItem_UnmarshalText(t *testing.T) {

	type Test struct {
		Name     string
		Item     string
		Valid    bool
		Expected Item
	}

	tests := []Test{
		{
			Name:  "Uncomplete description only",
			Item:  "Complete this test",
			Valid: true,
			Expected: Item{
				Message:       "Complete this test",
				Done:          false,
				Priority:      0,
				CreatedDate:   time.Time{},
				CompletedDate: time.Time{},
				Projects:      nil,
				Contexts:      nil,
				SpecialKeys:   nil,
			},
		},
		{
			Name:  "Complete description only",
			Item:  "x Complete this test",
			Valid: true,
			Expected: Item{
				Message:       "Complete this test",
				Done:          true,
				Priority:      0,
				CreatedDate:   time.Time{},
				CompletedDate: time.Time{},
				Projects:      nil,
				Contexts:      nil,
				SpecialKeys:   nil,
			},
		},
		{
			Name:  "Complete description creation date",
			Item:  "x 2024-05-18 Complete this test",
			Valid: true,
			Expected: Item{
				Message:       "Complete this test",
				Done:          true,
				Priority:      0,
				CreatedDate:   time.Date(2024, 05, 18, 0, 0, 0, 0, time.Local),
				CompletedDate: time.Time{},
				Projects:      nil,
				Contexts:      nil,
				SpecialKeys:   nil,
			},
		},
		{
			Name:  "Complete description completed date, creation date",
			Item:  "x 2024-05-18 2024-05-17 Complete this test",
			Valid: true,
			Expected: Item{
				Message:       "Complete this test",
				Done:          true,
				Priority:      0,
				CreatedDate:   time.Date(2024, 05, 17, 0, 0, 0, 0, time.Local),
				CompletedDate: time.Date(2024, 05, 18, 0, 0, 0, 0, time.Local),
				Projects:      nil,
				Contexts:      nil,
				SpecialKeys:   nil,
			},
		},
		{
			Name:  "Complete description completed date, creation date, project",
			Item:  "x 2024-05-18 2024-05-17 Complete this test +todoProject",
			Valid: true,
			Expected: Item{
				Message:       "Complete this test +todoProject",
				Done:          true,
				Priority:      0,
				CreatedDate:   time.Date(2024, 05, 17, 0, 0, 0, 0, time.Local),
				CompletedDate: time.Date(2024, 05, 18, 0, 0, 0, 0, time.Local),
				Projects:      []string{"todoProject"},
				Contexts:      nil,
				SpecialKeys:   nil,
			},
		},
		{
			Name:  "Complete description completed date, creation date, context",
			Item:  "x 2024-05-18 2024-05-17 Complete this test @todoContext",
			Valid: true,
			Expected: Item{
				Message:       "Complete this test @todoContext",
				Done:          true,
				Priority:      0,
				CreatedDate:   time.Date(2024, 05, 17, 0, 0, 0, 0, time.Local),
				CompletedDate: time.Date(2024, 05, 18, 0, 0, 0, 0, time.Local),
				Projects:      nil,
				Contexts:      []string{"todoContext"},
				SpecialKeys:   nil,
			},
		},
		{
			Name:  "Complete description completed date, creation date, special keys",
			Item:  "x 2024-05-18 2024-05-17 Complete this test key:value",
			Valid: true,
			Expected: Item{
				Message:       "Complete this test key:value",
				Done:          true,
				Priority:      0,
				CreatedDate:   time.Date(2024, 05, 17, 0, 0, 0, 0, time.Local),
				CompletedDate: time.Date(2024, 05, 18, 0, 0, 0, 0, time.Local),
				Projects:      nil,
				Contexts:      nil,
				SpecialKeys: map[string]string{
					"key": "value",
				},
			},
		},
		{
			Name:  "Empty string",
			Item:  "",
			Valid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(tt *testing.T) {
			var item Item
			err := item.UnmarshalText([]byte(test.Item))
			if err != nil && test.Valid {
				tt.Errorf("UnmarshalText failed on valid string: %v", err)
			}

			if err == nil && !test.Valid {
				tt.Errorf("UnmarshalText succeeded on invalid string: %s - item: %v", test.Item, item)
			}

			if test.Valid && !reflect.DeepEqual(item, test.Expected) {
				tt.Errorf(
					"got: %+v\nbut expected: %+v",
					item,
					test.Expected,
				)
			}
		})
	}
}

func TestItem_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected string
	}{
		{
			name:     "simple incomplete",
			item:     Item{Message: "buy milk"},
			expected: "buy milk",
		},
		{
			name:     "done no dates",
			item:     Item{Message: "buy milk", Done: true},
			expected: "x buy milk",
		},
		{
			name: "with creation date only",
			item: Item{
				Message:     "buy milk",
				CreatedDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local),
			},
			expected: "2024-01-15 buy milk",
		},
		{
			name: "done with both dates",
			item: Item{
				Message:       "buy milk",
				Done:          true,
				CreatedDate:   time.Date(2024, 1, 14, 0, 0, 0, 0, time.Local),
				CompletedDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local),
			},
			expected: "x 2024-01-15 2024-01-14 buy milk",
		},
		{
			name: "done with creation date but no completed date",
			item: Item{
				Message:     "buy milk",
				Done:        true,
				CreatedDate: time.Date(2024, 1, 14, 0, 0, 0, 0, time.Local),
			},
			expected: "x 2024-01-14 buy milk",
		},
		{
			name: "with project and context in message",
			item: Item{
				Message:     "buy milk +groceries @errands",
				CreatedDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local),
				Projects:    []string{"groceries"},
				Contexts:    []string{"errands"},
			},
			expected: "2024-01-15 buy milk +groceries @errands",
		},
		{
			name: "with priority",
			item: Item{
				Message:  "urgent task",
				Priority: Priority('A'),
			},
			expected: "(A) urgent task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := tt.item.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText error: %v", err)
			}
			if got := string(text); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestReadWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "todo.txt")

	// Reading a non-existent file returns an empty list, not an error.
	list, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile on missing file: %v", err)
	}
	if got := list.GetAll(); len(got) != 0 {
		t.Fatalf("expected empty list, got %d items", len(got))
	}

	// Write some items and read them back.
	items := []Item{
		{
			Message:     "buy milk +groceries @errands",
			CreatedDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local),
			Projects:    []string{"groceries"},
			Contexts:    []string{"errands"},
		},
		{
			Message:       "call dentist @health",
			Done:          true,
			CreatedDate:   time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local),
			CompletedDate: time.Date(2024, 1, 12, 0, 0, 0, 0, time.Local),
			Contexts:      []string{"health"},
		},
	}

	for _, item := range items {
		list.Add(item)
	}

	if err := WriteFile(path, list); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Verify the file exists and has content.
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if stat.Size() == 0 {
		t.Fatal("file is empty")
	}

	// Read back and compare.
	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after write: %v", err)
	}
	gotItems := got.GetAll()
	if len(gotItems) != len(items) {
		t.Fatalf("got %d items, want %d", len(gotItems), len(items))
	}

	for i, want := range items {
		g := gotItems[i]
		if g.Message != want.Message {
			t.Errorf("item %d: message got %q, want %q", i, g.Message, want.Message)
		}
		if g.Done != want.Done {
			t.Errorf("item %d: done got %v, want %v", i, g.Done, want.Done)
		}
		if !g.CreatedDate.Equal(want.CreatedDate) {
			t.Errorf("item %d: created got %v, want %v", i, g.CreatedDate, want.CreatedDate)
		}
		if !g.CompletedDate.Equal(want.CompletedDate) {
			t.Errorf("item %d: completed got %v, want %v", i, g.CompletedDate, want.CompletedDate)
		}
	}
}

// TestWriteFile_Atomic checks that WriteFile does not leave a .tmp file on success.
func TestWriteFile_Atomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "todo.txt")
	list := &List{list: make(itemList, 0)}
	list.Add(Item{Message: "test item"})

	if err := WriteFile(path, list); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error(".tmp file should not exist after successful write")
	}
}
