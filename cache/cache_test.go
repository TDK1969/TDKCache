package mycache

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

var db = map[string]string{
	"Tom":  "100",
	"Jack": "101",
	"Sam":  "102",
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatal("callback failed")
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	g := NewGroup("score", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			fmt.Printf("[SlowDB] search key %s\n", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("key [%s] not exist", key)
		}))

	for k, v := range db {
		if view, err := g.Get(k); err != nil || view.String() != v {
			t.Fatalf("get error: %v\nfailed to get value of key [%s]", err, k)
		}
		if _, err := g.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache [%s] miss", k)
		}
	}

	if view, err := g.Get("unknown"); err == nil {
		t.Fatalf("the value of unknown should be empty, but got %s", view)
	}

}

func TestExpire(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	g := NewGroup("score", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			fmt.Printf("[SlowDB] search key %s\n", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("key [%s] not exist", key)
		}))
	for k, v := range db {
		if view, err := g.Get(k); err != nil || view.String() != v {
			t.Fatalf("get error: %v\nfailed to get value of key [%s]", err, k)
		}
		if _, err := g.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache [%s] miss", k)
		}
		time.Sleep(time.Second * 3)
		if _, err := g.Get(k); err != nil || loadCounts[k] < 2 {
			t.Fatalf("cache [%s] does not expire", k)
		}
	}
}
