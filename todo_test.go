package todo

import (
	"reflect"
	"testing"
	"time"
)

func TestItem_MarshalText(t *testing.T) {

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
	}

	for _, test := range tests {
		t.Run(test.Name, func(tt *testing.T) {
			var item Item
			err := item.UnmarshalText([]byte(test.Item))
			if err != nil && test.Valid {
				tt.Errorf("Item UnmarshalText failed on valid string: %v", err)
			}

			if err == nil && !test.Valid {
				tt.Errorf("Item UnmarshalText succeeded on invalid string: %s - item: %v", test.Item, item)
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

func TestList_Add(t *testing.T) {

}
