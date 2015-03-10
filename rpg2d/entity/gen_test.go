package entity

import "testing"

func TestIdGenerator(t *testing.T) {
	nextId := NewIdGenerator()

	for i := 0; i < 100; i++ {
		id := nextId()
		if id != Id(i) {
			t.Error("unexpected id value: ", id, " at i =", i)
		}
	}
}
