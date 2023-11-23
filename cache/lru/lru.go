package lru

import (
	"container/list"
)

type Cache struct {
	capacity  int64                    // 缓存容量
	length    int64                    // 当前缓存大小
	linkList  *list.List               // 链表
	cache     map[string]*list.Element //字典
	onEvicted func(key string, value Value)
}

// Value包装实际的数据类型,通过Len方法返回数据的字节长度
type Value interface {
	Len() int // 返回数据占用的字节数
}

type entry struct {
	key   string // 键
	value Value  // 值
}

func NewCache(capacity int64, onEvicted func(key string, value Value)) *Cache {
	/*
		* New一个Cache
		* Params:
		 	* capcacity: 缓存的容量
			* onEvicted: 移除数据时的回调函数
	*/
	return &Cache{
		capacity:  capacity,
		length:    0,
		linkList:  list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (c *Cache) Add(key string, value Value) bool {
	/*
		* 向缓存中添加数据
		* Params:
		 	* key: 键
			* value: 值
	*/
	// 如果值的字节数大于缓存容量,则无法加入缓存,返回错误
	if int64(value.Len())+int64(len(key)) > c.capacity {
		return false
	}

	if element, ok := c.cache[key]; ok {
		// 如果key已经在缓存中
		// 将数据移动到链表头
		c.linkList.MoveToFront(element)
		kv := element.Value.(*entry)
		// 修改缓存大小
		c.length += int64(value.Len()) - int64(kv.value.Len())
		// 修改缓存的值
		kv.value = value
	} else {
		// 如果key不在缓存中
		// 从链表头插入新数据
		kv := c.linkList.PushFront(&entry{key: key, value: value})
		c.cache[key] = kv
		c.length += int64(value.Len()) + int64(len(key))
	}

	for c.length > c.capacity {
		c.RemoveLeastRecentElement()
	}
	return true
}

func (c *Cache) Get(key string) (Value, bool) {
	/*
		* 从缓存中获取key对应的value
		* Params
		 	* key: 键
		* Return
		 	* Value: 值
			* bool: 是否成功
	*/

	if elem, ok := c.cache[key]; ok {
		c.linkList.MoveToFront(elem)
		kv := elem.Value.(*entry)
		return kv.value, true
	} else {
		return nil, false
	}
}

func (c *Cache) Delete(key string) (Value, bool) {
	/*
		* 从缓存中删除key及其对应的value
		* Params
		 	* key: 键
		* Return
		 	* Value: 值
			* bool: 是否成功
	*/

	if elem, ok := c.cache[key]; ok {
		c.linkList.Remove(elem)
		kv := elem.Value.(*entry)
		// 修改缓存大小
		c.length -= int64(kv.value.Len())
		// 从哈希表中删除数据
		delete(c.cache, kv.key)
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
		return kv.value, true
	} else {
		return nil, false
	}
}

func (c *Cache) RemoveLeastRecentElement() {
	/*
	 * 移除最久未使用的数据
	 */
	element := c.linkList.Back()
	if element != nil {
		// 从链表尾移除元素
		c.linkList.Remove(element)
		kv := element.Value.(*entry)
		// 修改缓存大小
		c.length -= int64(kv.value.Len()) + int64(len(kv.key))
		// 从哈希表中删除数据
		delete(c.cache, kv.key)

		// 执行回调函数(如果有)
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}

	}
}

func (c *Cache) Len() int {
	return c.linkList.Len()
}
