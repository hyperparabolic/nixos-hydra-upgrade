package assert

import (
	"testing"
)

// Simple test assertions with generics.
// This package really doesn't need much, will pull in
// https://github.com/google/go-cmp if I find myself
// writing too many more of these helpers, but even that
// feels overkill for now.

func Equal[T comparable](t *testing.T, got, expected T) {
	t.Helper()

	if got != expected {
		t.Errorf(`assert.Equal(t, ...)
got:
%v
expected:
%v`, got, expected)
	}
}

func ArrayEqual[T comparable](t *testing.T, got, expected []T) {
	if len(got) != len(expected) {
		t.Errorf(`assert.ArrayEqual(t, ...)
got:
%v
expected:
%v`, got, expected)
	}

	for index, element := range got {
		if element != expected[index] {
			t.Errorf(`assert.ArrayEqual(t, ...)
got:
%v
expected:
%v`, got, expected)
		}
	}
}
