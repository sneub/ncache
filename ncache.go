package ncache

import (
    //"fmt"
    "os"
    "bytes"
    "sync"
    "errors"
    "container/list"
)

type Cache struct {
    size    int                           // Size of cache (in MB)
    tracker int                           // Counter for tracking current size of cache
    index   *list.List                    // List of pointers to the keys
    data    map[interface{}]*list.Element // Hash map with the actual data
    lock    sync.RWMutex
}

type key struct {
    key        interface{}
    value      interface{}
    popularity int
}

func New (size int) (*Cache, error) {
    if size <= 0 {
        return nil, errors.New("Invalid cache size")
    }
    mapsize := (size * 1024 * 1024) / os.Getpagesize()

    //fmt.Printf("Initialising cache with size %d MB / %d pages (system page size %d)\n", size, mapsize, os.Getpagesize())

    c := &Cache{
        size:    size,
        tracker: 0,
        index:   list.New(),
        data:    make(map[interface{}]*list.Element, mapsize),
    }
    return c, nil
}

func (c *Cache) Set (keyname string, value *bytes.Buffer) (found bool) {
    c.lock.Lock()
    defer c.lock.Unlock()

    //fmt.Printf("nCache SET: %s (%d bytes)\n", keyname, value.Len())

    k := &key{keyname, value, 0}
    c.data[keyname] = c.index.PushFront(k)
    return true
}

func (c *Cache) Get (keyname string) (value interface{}, found bool) {
    c.lock.Lock()
    defer c.lock.Unlock()

    //fmt.Printf("nCache GET: %s\n", keyname)

    k, found := c.data[keyname]
    if found {
        return k.Value.(*key).value, true
    }
    return nil, false
}
