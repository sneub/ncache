

# ncache - a smart cache library for Go
[![Build Status](https://travis-ci.org/sneub/ncache.svg?branch=master)](https://travis-ci.org/sneub/ncache)

**ncache** bases its cache size on storage capacity (MB) rather than number of pages or keys. As such, it is not possible to use a straight-swap LRU policy, as the key that is evicted may be smaller than the key being added, resulting in a cache size that is larger than planned. Doing a one-for-one swap just doesn't cut it around here.

**ncache** will iteratively evict enough keys to cover the size of the new key being added.

Other eviction methods are also used to ensure cached data is not stale, such as a time-to-live (TTL) value stored with the keys. A janitor process will clean up stale keys.

### To do:
- Track popularity of keys - use this to proactively refresh/re-cache popular keys. This could be done asynchronously before the key expires.
- Method to return key meta data - could be useful for building a scaleable cluster of cache nodes, or simply saving the state of cache
