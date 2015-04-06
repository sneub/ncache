package ncache

import (
    "fmt"
    "testing"
    "time"
    "bytes"
    "github.com/dustin/randbo"
)

func TestSetGet(t *testing.T) {
    cache, err := New(1, 30, 120)

    if err != nil {
        t.Fatalf("Error creating cache: %v", err)
    }

    var data []byte
    data = make([]byte, 512)
    randbo.New().Read(data)

    buf := bytes.NewBuffer(data)
    success := cache.Set("foo", buf)

    if ! success {
        t.Fatalf("couldn't set key")
    }

    expected, found := cache.Get("foo")

    if ! found {
        t.Fatalf("get key was not found")
    }

    if expected == nil {
        t.Fatalf("get data came back empty")
    }

    if fmt.Sprintf("%s", data) != fmt.Sprintf("%s", expected) {
        t.Fatalf("returned data doesn't match original key")
    }
}

func TestLRUEvict(t *testing.T) {
    cache, err := New(1, 30, 120)

    if err != nil {
        t.Fatalf("error creating cache: %v", err)
    }

    var data []byte
    data = make([]byte, 51200, 51200)

    buf := bytes.NewBuffer(data)
    cache.Set("foo", buf)

    // Fill the cache
    for i := 0; i < 20; i++ {
        cache.Set(string(i), buf)
    }

    // 'foo' should have been evicted
    _, found := cache.Get("foo")
    if found {
        t.Fatalf("LRU key should have been evicted")
    }
}

func TestExpiry(t *testing.T) {
    cache, err := New(1, 1, 1)

    if err != nil {
        t.Fatalf("error creating cache: %v", err)
    }

    var data []byte
    data = make([]byte, 51200, 51200)

    buf := bytes.NewBuffer(data)
    cache.Set("foo", buf)

    time.Sleep(2 * time.Second)

    // 'foo' should have been evicted
    _, found := cache.Get("foo")
    if found {
        t.Fatalf("Expired key should have been evicted")
    }
}
