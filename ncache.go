package ncache

import (
    "fmt"
    "os"
    "bytes"
    "sync"
    "errors"
    "time"
    "container/list"
)

type Cache struct {
    size    int                           // Size of cache (in MB)
    tracker int                           // Counter for tracking current size of cache
    index   *list.List                    // List of pointers to the keys
    data    map[interface{}]*list.Element // Hash map with the actual data
    poll    time.Duration
    ttl     int
    lock    sync.RWMutex
}

type key struct {
    key        interface{}
    value      interface{}
    size       int
    birthday   time.Time
    popularity int
}

func New (size int) (*Cache, error) {
    if size <= 0 {
        return nil, errors.New("Invalid cache size")
    }
    mapsize := (size * 1024 * 1024) / os.Getpagesize()

    fmt.Printf("Initialising cache with size %d MB / %d pages (system page size %d)\n", size, mapsize, os.Getpagesize())

    c := &Cache{
        size:    (size*1024*1024),
        tracker: 0,
        index:   list.New(),
        data:    make(map[interface{}]*list.Element, mapsize),
        poll:    30,
        ttl:     120,
    }
    go c.evictor(0)
    return c, nil
}

func (c *Cache) Set (keyname string, value *bytes.Buffer) (found bool) {
    c.lock.Lock()
    defer c.lock.Unlock()

    fmt.Printf("ncache SET: %s (%d bytes)\n", keyname, value.Len())
    keysize := value.Len()

    if (c.tracker + keysize) >= c.size {
        fmt.Println("Cache limit reached. Running evictor")
        c.evictor(keysize)
    }

    c.tracker += keysize
    fmt.Printf("Tracker at %d / %d bytes\n", c.tracker, c.size)

    k := &key{keyname, value, value.Len(), time.Now(), 0}
    c.data[keyname] = c.index.PushFront(k)
    return true
}

func (c *Cache) Get (keyname string) (value interface{}, found bool) {
    c.lock.Lock()
    defer c.lock.Unlock()

    fmt.Printf("ncache GET: %s\n", keyname)

    k, found := c.data[keyname]
    if found {
        return k.Value.(*key).value, true
    }
    return nil, false
}

func (c *Cache) freespace () (free int) {
    return (c.size - c.tracker)
}

func (c *Cache) removeElement (keyname *list.Element) {
    c.index.Remove(keyname)
    k := keyname.Value.(*key)
    delete(c.data, k.key)
    c.tracker -= k.size
}

func (c *Cache) removeOldest () {
    k := c.index.Back()
    if k != nil {
        c.removeElement(k)
    }
}

func (c *Cache) evictor (size int) {
    if size == 0 {
        for {
            for i := c.index.Front(); i != nil; i = i.Next() {
                keyage := time.Now().Sub(i.Value.(*key).birthday).Seconds()
                if (int(keyage) >= c.ttl) {
                    fmt.Printf("Key TTL expired %s\n", i.Value.(*key).key)
                    c.removeElement(i)
                }
            }
            time.Sleep(c.poll * time.Second)
        }
    } else {
        for c.freespace() < size {
            c.removeOldest()
        }
    }
}
