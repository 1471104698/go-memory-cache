package main

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	c := MemoryCacheConstructor(2)
	c.Set("1", 1, -1)
	c.Set("2", 2, -1)
	v := c.Get("1")
	if v != 1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, 1)
	}
	c.Set("3", 3, -1)
	v = c.Get("2")
	if v != -1 {
		t.Errorf("Getting key=2 err, act: %v, expect: %v", v, -1)
	}
	c.Set("4", 4, -1)
	v = c.Get("1")
	if v != -1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, -1)
	}
	v = c.Get("3")
	if v != 3 {
		t.Errorf("Getting key=3 err, act: %v, expect: %v", v, 3)
	}
	v = c.Get("4")
	if v != 4 {
		t.Errorf("Getting key=4 err, act: %v, expect: %v", v, 4)
	}
}

func TestMemoryCacheScanAndDelete(t *testing.T) {
	c := MemoryCacheConstructor(2)
	c.Set("1", 1, 1)
	c.Set("2", 2, -1)
	v := c.Get("1")
	if v != 1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, 1)
	}
	v = c.Get("2")
	if v != 2 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, 2)
	}
	time.Sleep(2 * time.Second)
	v = c.Get("1")
	if v != -1 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, -1)
	}
	v = c.Get("2")
	if v != 2 {
		t.Errorf("Getting key=1 err, act: %v, expect: %v", v, 2)
	}
}
