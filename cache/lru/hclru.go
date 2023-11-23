package lru

import (
	"container/list"
	"time"
)

type HCCache struct {
	heatCapacity int64                         // 热数据区缓存容量
	heatLength   int64                         // 热数据区当前缓存大小
	coldCapacity int64                         // 冷数据区缓存容量
	coldLength   int64                         // 冷数据区当前缓存大小
	heatLinklist *list.List                    // 热数据链表头
	coldLinklist *list.List                    // 冷数据链表哨兵
	heatCache    map[string]*list.Element      // 热数据哈希表
	coldCache    map[string]*list.Element      // 冷数据哈希表
	onEvicted    func(key string, value Value) // 回调函数
}

type hcEntry struct {
	key       string // 键
	value     Value  // 值
	timestamp int64  // 加入时间
}

func NewEntry(key string, value Value) *hcEntry {
	return &hcEntry{key: key, value: value, timestamp: time.Now().UnixMilli()}
}

func (e *hcEntry) Len() int {
	return len(e.key) + e.value.Len() + 8
}

func (e *hcEntry) ResetTs() {
	e.timestamp = time.Now().UnixMilli()
}

func NewHCCache(capacity int64, onEvicted func(key string, value Value)) *HCCache {
	return &HCCache{
		heatCapacity: capacity,
		heatLength:   0,
		coldCapacity: capacity / 2,
		coldLength:   0,
		heatLinklist: list.New(),
		coldLinklist: list.New(),
		heatCache:    make(map[string]*list.Element),
		coldCache:    make(map[string]*list.Element),
		onEvicted:    onEvicted,
	}
}

func (c *HCCache) Add(key string, value Value) bool {
	if element, ok := c.heatCache[key]; ok {
		// 如果数据在热数据区,移动到链表头
		c.heatLinklist.MoveToFront(element)
		e := element.Value.(*hcEntry)

		c.heatLength += int64(value.Len()) - int64(e.value.Len())
		e.value = value
	} else if element, ok := c.coldCache[key]; ok {
		// 如果数据在冷数据区,根据访问间隔判断是否需要移动到热数据区
		e := element.Value.(*hcEntry)
		if time.Now().UnixMilli()-e.timestamp < 1000 {
			// 如果间隔小于1s,加入热数据区
			c.coldLinklist.Remove(element)
			c.coldLength -= int64(e.Len())
			delete(c.coldCache, key)

			c.heatCache[key] = c.heatLinklist.PushFront(e)
			c.heatLength += int64(e.Len())
		}
		// 如果大于1s,则什么都不做
	} else {
		// 新数据加入冷数据区的链表头
		e := NewEntry(key, value)
		c.coldLength += int64(e.Len())
		c.coldCache[key] = c.coldLinklist.PushFront(e)
	}

	// 处理超出缓存
	c.replace()

	return true

}

func (c *HCCache) Get(key string) (Value, bool) {
	if elem, ok := c.heatCache[key]; ok {
		c.heatLinklist.MoveToFront(elem)
		return elem.Value.(*hcEntry).value, true
	} else if elem, ok := c.coldCache[key]; ok {
		e := elem.Value.(*hcEntry)
		if time.Now().UnixMilli()-e.timestamp < 1000 {
			// 如果间隔小于1s,加入热数据区
			c.coldLinklist.Remove(elem)
			c.coldLength -= int64(e.Len())
			delete(c.coldCache, key)

			c.heatCache[key] = c.heatLinklist.PushFront(e)
			c.heatLength += int64(e.Len())

			c.replace()
		}
		return e.value, true
	}

	return nil, false

}

func (c *HCCache) Delete(key string) (Value, bool) {
	if elem, ok := c.heatCache[key]; ok {
		c.heatLinklist.Remove(elem)
		e := elem.Value.(*hcEntry)
		c.heatLength -= int64(e.Len())
		delete(c.heatCache, e.key)
		if c.onEvicted != nil {
			c.onEvicted(e.key, e.value)
		}
		return e.value, true
	} else if elem, ok := c.coldCache[key]; ok {
		c.coldLinklist.Remove(elem)
		e := elem.Value.(*hcEntry)
		c.coldLength -= int64(e.Len())
		delete(c.coldCache, e.key)
		if c.onEvicted != nil {
			c.onEvicted(e.key, e.value)
		}
		return e.value, true
	}
	return nil, false
}

func (c *HCCache) replace() {
	// 进行淘汰策略

	for c.heatLength > c.heatCapacity {
		// 对热数据区进行淘汰
		element := c.heatLinklist.Back()
		e := element.Value.(*hcEntry)
		c.heatLinklist.Remove(element)
		c.heatLength -= int64(e.Len())
		delete(c.heatCache, e.key)

		// 加入冷数据区
		e.ResetTs()
		c.coldLength += int64(e.Len())
		c.coldCache[e.key] = c.coldLinklist.PushFront(e)
	}

	for c.coldLength > c.coldCapacity {
		// 对冷数据区进行淘汰
		element := c.coldLinklist.Back()
		e := element.Value.(*hcEntry)
		c.coldLinklist.Remove(element)
		c.coldLength -= int64(e.Len())
		delete(c.coldCache, e.key)
		if c.onEvicted != nil {
			c.onEvicted(e.key, e.value)
		}
	}
}

func (c *HCCache) Len() int {
	return c.coldLinklist.Len() + c.heatLinklist.Len()
}
