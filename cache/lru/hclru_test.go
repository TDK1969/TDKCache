package lru

import (
	"reflect"
	"testing"
)

func TestHCGet(t *testing.T) {
	lru := NewHCCache(int64(100), nil)
	lru.Add("key1", String("1234"), 0)
	if v, ok := lru.Get("key1", 0); !ok || string(v.(String)) != "1234" || lru.coldLength != 0 {
		t.Fatalf("cache hit key1=1234 failed 1")
	}
	if _, ok := lru.Get("key2", 0); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestHCGet2(t *testing.T) {
	lru := NewHCCache(int64(100), nil)
	lru.Add("key1", String("1234"), 0)
	if v, ok := lru.Get("key1", 2005); !ok || string(v.(String)) != "1234" || lru.heatLength != 0 {
		t.Fatalf("cache hit key1=1234 failed 1")
	}
	if _, ok := lru.Get("key2", 0); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestHCRemoveoldest(t *testing.T) {
	k1, k2, k3, k4 := "key1", "key2", "key3", "key4"
	v1, v2, v3, v4 := "value1", "value2", "value3", "value4"
	cap := len(k1+k2+v1+v2) + 16
	lru := NewHCCache(int64(cap), nil)

	lru.Add(k1, String(v1), 0)
	if _, ok := lru.Get("key1", 0); !ok || lru.Len() != 1 {
		t.Fatalf("store key1 failed")
	}

	lru.Add(k2, String(v2), 1)
	if _, ok := lru.Get("key2", 1); !ok || lru.Len() != 2 {
		t.Fatalf("store key2 failed")
	}

	lru.Add(k3, String(v3), 2)
	if _, ok := lru.Get("key3", 2); !ok || lru.Len() != 3 {
		t.Fatalf("store key3 failed")
	}

	lru.Add(k4, String(v4), 3)
	if _, ok := lru.Get("key4", 3); !ok || lru.Len() != 3 {
		t.Fatalf("store key4 failed")
	}

	if _, ok := lru.Get("key1", 3); ok || lru.Len() != 3 {
		t.Fatalf("Removeoldest key1 failed")
	}

}

func TestHCOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := NewHCCache(int64(24), callback)
	lru.Add("k1", String("k1"), 0)
	lru.Get("k1", 0)
	lru.Add("k2", String("k2"), 0)
	lru.Get("k2", 0)
	lru.Add("k3", String("k3"), 0)
	lru.Add("k4", String("k4"), 0)

	expect := []string{"k3"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s, but get %s", expect, keys)
	}
}
