package main

import "testing"

func Test(t *testing.T) {
	_, err := initialize()

	if err != nil {
		t.Error("failed initialize()")
	}
}
