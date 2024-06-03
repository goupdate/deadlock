package structmap

import (
	"os"
	"testing"
)

type ExampleStruct struct {
	Field1 string
	Field2 int
}

func TestNew(t *testing.T) {
	storage, err := New[*ExampleStruct]("test_storage", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if storage == nil {
		t.Fatalf("expected non-nil storage")
	}
}

func TestAddAndGet(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example := &ExampleStruct{Field1: "value1", Field2: 42}
	id := storage.Add(example)

	item, ok := storage.Get(id)
	if !ok {
		t.Fatalf("expected item, got nil")
	}
	if item.Field1 != "value1" || item.Field2 != 42 {
		t.Fatalf("expected %v, got %v", example, item)
	}
}

func TestSetField(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example := &ExampleStruct{Field1: "value1", Field2: 42}
	id := storage.Add(example)

	updated := storage.SetField(id, "Field1", "newValue")
	if !updated {
		t.Fatalf("expected field to be updated")
	}

	item, ok := storage.Get(id)
	if !ok || item.Field1 != "newValue" {
		t.Fatalf("expected field1 to be 'newValue', got %v", item.Field1)
	}
}

func TestGetAll(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	items := storage.GetAll()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestFindEqual(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field1", Value: "value1", Op: "equal"})
	if len(results) != 1 || results[0].Field1 != "value1" {
		t.Fatalf("expected to find 1 item with Field1 'value1', got %d items", len(results))
	}
}

func TestFindGreaterThan(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field2", Value: 42, Op: "gt"})
	if len(results) != 1 || results[0].Field2 != 43 {
		t.Fatalf("expected to find 1 item with Field2 greater than 42, got %d items", len(results))
	}
}

func TestFindLessThan(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field2", Value: 43, Op: "lt"})
	if len(results) != 1 || results[0].Field2 != 42 {
		t.Fatalf("expected to find 1 item with Field2 less than 43, got %d items", len(results))
	}
}

func TestFindLike(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field1", Value: "value", Op: "like"})
	if len(results) != 2 {
		t.Fatalf("expected to find 2 items with Field1 containing 'value', got %d items", len(results))
	}
}

func TestDelete(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	id1 := storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field1", Value: "value", Op: "like"})
	if len(results) != 2 {
		t.Fatalf("expected to find 2 items with Field1 containing 'value', got %d items", len(results))
	}

	storage.Delete(id1)
	_, ok := storage.Get(id1)
	if ok {
		t.Fatal("value should be removed")
	}
}

func TestClear(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)
	storage.Clear()

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)

	results := storage.Find("", FindCondition{Field: "Field1", Value: "value", Op: "like"})
	if len(results) == 0 {
		t.Fatalf("expected to find items")
	}

	storage.Clear()

	items := storage.GetAll()
	if len(items) > 0 {
		t.Fatal("should be 0 items")
	}
}

func TestIterate(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}
	example3 := &ExampleStruct{Field1: "other", Field2: 44}

	storage.Add(example1)
	storage.Add(example2)
	storage.Add(example3)

	var items []*ExampleStruct
	storage.Iterate(func(v *ExampleStruct) bool {
		items = append(items, v)
		return true
	})

	if len(items) != 3 {
		t.Fatalf("expected to iterate over 3 items, got %d items", len(items))
	}

	// Test stopping iteration early
	var partialItems []*ExampleStruct
	storage.Iterate(func(v *ExampleStruct) bool {
		partialItems = append(partialItems, v)
		return len(partialItems) < 2 // Stop after collecting 2 items
	})

	if len(partialItems) != 2 {
		t.Fatalf("expected to iterate over 2 items before stopping, got %d items", len(partialItems))
	}
}

func TestFindWithOrConditions(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}
	example3 := &ExampleStruct{Field1: "other", Field2: 42}

	storage.Add(example1)
	storage.Add(example2)
	storage.Add(example3)

	// Test OR condition
	results := storage.Find("OR", FindCondition{Field: "Field1", Value: "value1", Op: "equal"}, FindCondition{Field: "Field2", Value: 43, Op: "equal"})
	if len(results) != 2 {
		t.Fatalf("expected to find 2 items with Field1 'value1' or Field2 43, got %d items", len(results))
	}
}

func TestFindWithAndConditions(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example1 := &ExampleStruct{Field1: "value1", Field2: 42}
	example2 := &ExampleStruct{Field1: "value2", Field2: 43}
	example3 := &ExampleStruct{Field1: "value1", Field2: 43}

	storage.Add(example1)
	storage.Add(example2)
	storage.Add(example3)

	// Test AND condition
	results := storage.Find("AND", FindCondition{Field: "Field1", Value: "value1", Op: "equal"}, FindCondition{Field: "Field2", Value: 42, Op: "equal"})
	if len(results) != 1 || results[0].Field1 != "value1" || results[0].Field2 != 42 {
		t.Fatalf("expected to find 1 item with Field1 'value1' and Field2 42, got %d items", len(results))
	}
}

func TestSave(t *testing.T) {
	storage, _ := New[*ExampleStruct]("test_storage", false)

	example := &ExampleStruct{Field1: "value1", Field2: 42}
	id := storage.Add(example)

	err := storage.Save()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	storage2, _ := New[*ExampleStruct]("test_storage", false)
	example2, ok := storage2.Get(id)
	if !ok {
		t.Fatal("not found")
	}
	if *example != *example2 {
		t.Fatalf("fail on load after save: %+v vs %+v", example, example2)
	}

	// Clean up test files
	os.Remove("test_storage")
	os.Remove("test_storagei")
}
