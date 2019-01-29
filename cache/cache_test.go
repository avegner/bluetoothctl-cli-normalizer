package cache_test

import (
	"fmt"
	"testing"

	"github.com/avegner/bluetoothctl-cli-normalizer/cache"
)

const (
	cacheSize = 3
)

func new(t *testing.T, size int) *cache.Cache {
	t.Helper()

	c := cache.New(cacheSize)
	if c == nil {
		t.Fatalf("New: have nil, want not nil")
	}

	return c
}

func getPrev(t *testing.T, c *cache.Cache, want string) {
	t.Helper()

	if have := c.GetPrev(); have != want {
		t.Fatalf("GetPrev: have '%s', want '%s'", have, want)
	}
}

func getNext(t *testing.T, c *cache.Cache, want string) {
	t.Helper()

	if have := c.GetNext(); have != want {
		t.Fatalf("GetNext: have '%s', want '%s'", have, want)
	}
}

func TestNew(t *testing.T) {
	new(t, cacheSize)
}

func TestAddGetPrev(t *testing.T) {
	c := new(t, cacheSize)

	for i := 0; i < cacheSize; i++ {
		c.Add(fmt.Sprintf("command line %d", i))
	}

	for i := cacheSize - 1; i >= 0; i-- {
		getPrev(t, c, fmt.Sprintf("command line %d", i))
	}
	getPrev(t, c, fmt.Sprintf("command line %d", cacheSize-1))
}

func TestAddGetNext(t *testing.T) {
	c := new(t, cacheSize)

	for i := 0; i < cacheSize; i++ {
		c.Add(fmt.Sprintf("command line %d", i))
	}

	for i := 0; i < cacheSize; i++ {
		getNext(t, c, fmt.Sprintf("command line %d", i))
	}
	getNext(t, c, fmt.Sprintf("command line %d", 0))
}

func TestGetEmpty(t *testing.T) {
	c := new(t, cacheSize)
	empty := ""

	getNext(t, c, empty)
	getPrev(t, c, empty)
}

func TestAddSameLine(t *testing.T) {
	c := new(t, cacheSize)
	same := "the same line"
	other := "the other line"

	c.Add(other)
	c.Add(same)
	c.Add(same)

	getPrev(t, c, same)
	getPrev(t, c, other)
	getPrev(t, c, same)
}

func TestAddOverflow(t *testing.T) {
	c := new(t, cacheSize)

	for i := 0; i <= cacheSize; i++ {
		c.Add(fmt.Sprintf("command line %d", i))
	}

	for i := cacheSize; i > 0; i-- {
		getPrev(t, c, fmt.Sprintf("command line %d", i))
	}
	getPrev(t, c, fmt.Sprintf("command line %d", cacheSize))
}
