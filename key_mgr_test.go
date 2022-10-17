package main

import (
	"testing"
)

type TestStruct struct {
	Num      int
	Children []*TestStruct
}

func TestLRU(t *testing.T) {
	lru := LRUConstructor(2)
	lru.Set("1", 1, -1)
	lru.Set("2", 2, -1)
	v := lru.Get("1")
	if v != 1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, 1)
	}
	lru.Set("3", 3, -1)
	v = lru.Get("2")
	if v != -1 {
		t.Errorf("Getting key=2 err, act: %v, expect: %v", v, -1)
	}
	lru.Set("4", 4, -1)
	v = lru.Get("1")
	if v != -1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, -1)
	}
	v = lru.Get("3")
	if v != 3 {
		t.Errorf("Getting key=3 err, act: %v, expect: %v", v, 3)
	}
	v = lru.Get("4")
	if v != 4 {
		t.Errorf("Getting key=4 err, act: %v, expect: %v", v, 4)
	}
}

func TestLFUGetAndSet(t *testing.T) {
	lru := LFUConstructor(2)
	lru.Set("1", 1)
	lru.Set("2", 2)
	v := lru.Get("1")
	if v != 1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, 1)
	}
	lru.Set("3", 3)
	v = lru.Get("2")
	if v != -1 {
		t.Errorf("Getting key=2 err, act: %v, expect: %v", v, -1)
	}
	v = lru.Get("3")
	if v != 3 {
		t.Errorf("Getting key=3 err, act: %v, expect: %v", v, 3)
	}
	lru.Set("4", 4)
	v = lru.Get("1")
	if v != -1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, -1)
	}
	v = lru.Get("3")
	if v != 3 {
		t.Errorf("Getting key=3 err, act: %v, expect: %v", v, 3)
	}
	v = lru.Get("4")
	if v != 4 {
		t.Errorf("Getting key=4 err, act: %v, expect: %v", v, 4)
	}
}

func TestLFUDelete(t *testing.T) {
	lru := LFUConstructor(2)
	lru.Set("1", 1)
	lru.Set("2", 2)
	lru.Delete("1")
	v := lru.Get("1")
	if v != -1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, -1)
	}
	lru.Set("2", 22)
	v = lru.Get("2") // time = 3
	if v != 22 {
		t.Errorf("Getting key=2 err, act: %v, expect: %v", v, 22)
	}
	lru.Set("3", 3)
	lru.Set("4", 4)
	v = lru.Get("3")
	if v != -1 {
		t.Errorf("Getting key=3 err, act: %v, expect: %v", v, -1)
	}
	v = lru.Get("4")
	if v != 4 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, 4)
	}
	lru.Delete("4")
	v = lru.Get("4")
	if v != -1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, -1)
	}
}
