package bbb

import (
	"testing"
)

func TestCreate(t *testing.T) {
	b3, _ := New("http://localhost/", "secret")
	t.Log(b3.Create("123", EmptyOptions))
}

func TestJoinURL(t *testing.T) {
	b3, _ := New("http://localhost/", "secret")
	t.Log(b3.JoinURL("Tim Jurcka", "123", "123", EmptyOptions))
}
