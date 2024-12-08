package lru

import (
	"container/list"
)

type Cache struct {
	maxBytes  int64      //容量
	nBytes    int64      //已使用的空间
	ll        *list.List //双向链表作队列
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) //某条记录被删除时的回调函数，可以为nil
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 查找：第一步是从字典中找到对应的双向链表的节点，第二步将该节点移到队尾
func (cache *Cache) Get(key string) (value Value, ok bool) {
	if elem, ok := cache.cache[key]; ok {
		cache.ll.MoveToBack(elem)
		kv := elem.Value.(*entry)
		return kv.value, true
	}
	return
}

// 删除最近最少访问的节点（队首）
func (cache *Cache) removeOldest() {
	elem := cache.ll.Front()
	if elem != nil {
		cache.ll.Remove(elem)
		kv := elem.Value.(*entry)
		delete(cache.cache, kv.key)
		cache.nBytes -= int64(kv.value.Len()) + int64(len(kv.key))
		if cache.OnEvicted != nil {
			cache.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 新增/修改：键存在就是更新对应值，并把该节点移到队尾
// 不存在就是新增
func (cache *Cache) Add(key string, value Value) {
	if elem, ok := cache.cache[key]; ok {
		cache.ll.MoveToBack(elem)
		kv := elem.Value.(*entry)
		cache.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		elem := cache.ll.PushBack(&entry{key, value})
		cache.cache[key] = elem
		cache.nBytes += int64(value.Len()) + int64(len(key))
	}
	for cache.maxBytes != 0 && cache.maxBytes < cache.nBytes {
		cache.removeOldest()
	}
}

func (cache *Cache) Len() int {
	return cache.ll.Len()
}
