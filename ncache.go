package ncache

import (
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

func New (size int, poll time.Duration, ttl int) (*Cache, error) {
    if size <= 0 {
        return nil, errors.New("Invalid cache size")
    }
    mapsize := (size * 1024 * 1024) / os.Getpagesize()

    c := &Cache{
        size:    (size*1024*1024),
        tracker: 0,
        index:   list.New(),
        data:    make(map[interface{}]*list.Element, mapsize),
        poll:    poll,
        ttl:     ttl,
    }
    go c.evictor(0)
    return c, nil
}

func (c *Cache) Set (keyname string, value *bytes.Buffer) (success bool) {
    keysize := value.Len()
    if (c.tracker + keysize) >= c.size {
        c.evictor(keysize)
    }

    c.lock.Lock()
    defer c.lock.Unlock()

    c.tracker += keysize
    k := &key{keyname, value, value.Len(), time.Now(), 0}
    c.data[keyname] = c.index.PushFront(k)

    return true
}

func (c *Cache) Get (keyname string) (value interface{}, found bool) {
    c.lock.RLock()
    defer c.lock.RUnlock()

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
    c.lock.Lock()
    defer c.lock.Unlock()

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
                if int(keyage) >= c.ttl {
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
