package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	mh := myHandler{}
	h := NoSurf(&mh)

	switch v := h.(type) {
	case http.Handler:
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}

func TestSessionLoad(t *testing.T) {
	mh := myHandler{}
	h := SessionLoad(&mh)

	switch v := h.(type) {
	case http.Handler:
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}
