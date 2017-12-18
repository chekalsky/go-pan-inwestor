package main

import (
	"testing"
	"reflect"
)

func TestDifference(t *testing.T) {
	slice1 := []string{"A", "B", "C"}
	slice2 := []string{"B", "C", "D"}

	diff := difference(slice1, slice2)

	s := []string{"D"}
	if !reflect.DeepEqual(diff, s) {
		t.Error("Wrong difference")
	}
}
