package main

import (
	"sync"
)

/*
	key 淘汰策略：
	1、LRU
	2、LFU
*/

type KeyMgr interface {
	Get(key string) interface{}
	Set(key string, value interface{}) bool
	Delete(key string) bool
}

type LRUNode struct {
	key       string
	val       interface{}
	expire    int64
	pre, next *LRUNode
}

func newLRUNode(key string, value interface{}, expire int64) *LRUNode {
	return &LRUNode{
		key:    key,
		val:    value,
		expire: expire,
	}
}

func emptyLRUNode() *LRUNode {
	return &LRUNode{}
}

type LRUCache struct {
	/*
	   O(1) 查询
	*/
	cap        int
	size       int
	head, tail *LRUNode
	nodeMap    map[string]*LRUNode
	lock       *sync.Mutex
}

func LRUConstructor(capacity int) LRUCache {
	head, tail := emptyLRUNode(), emptyLRUNode()
	head.next = tail
	tail.pre = head
	return LRUCache{
		cap:     capacity,
		size:    0,
		head:    head,
		tail:    tail,
		nodeMap: map[string]*LRUNode{},
		lock:    &sync.Mutex{},
	}
}

func (this *LRUCache) Get(key string) interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	node := this.nodeMap[key]
	if node == nil {
		return -1
	}
	moveToHead(this, node)
	return node.val
}

func (this *LRUCache) Set(key string, value interface{}, expire int64) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	// 从 map 中获取
	node := this.nodeMap[key]
	if node == nil {
		node = newLRUNode(key, value, expire)
		this.nodeMap[key] = node
		removeNode := addNode(this, node)
		if removeNode != nil {
			this.nodeMap[removeNode.key] = nil
			// delete(nodeMap, key)
		}
	} else {
		node.val = value // 由于是指针，map 里也会更新
		moveToHead(this, node)
	}
	return true
}

func (this *LRUCache) Delete(key string) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	// 找到 key 对应 Node
	node := this.nodeMap[key]
	if node == nil {
		return true
	}
	deleteNode(this, node)
	delete(this.nodeMap, node.key)
	return true
}

func addNode(this *LRUCache, node *LRUNode) *LRUNode {
	this.size++
	moveToHead(this, node)
	if this.size > this.cap {
		return removeTail(this)
	}
	return nil
}

func moveToHead(this *LRUCache, node *LRUNode) {
	if node.pre != nil {
		node.pre.next = node.next
		node.next.pre = node.pre
	}
	this.head.next.pre = node
	node.next = this.head.next
	node.pre = this.head
	this.head.next = node
}

func removeTail(this *LRUCache) *LRUNode {
	node := this.tail.pre
	if node == this.head {
		return nil
	}
	deleteNode(this, node)
	return node
}

func deleteNode(this *LRUCache, node *LRUNode) bool {
	if node == nil {
		return true
	}
	node.pre.next = node.next
	node.next.pre = node.pre
	node.pre = nil
	node.next = nil
	this.size--
	return true
}

type LFUNode struct {
	key       string
	val       interface{}
	time      int
	pre, next *LFUNode
}

func newLFUNode(key string, val interface{}, time int) *LFUNode {
	return &LFUNode{
		key:  key,
		val:  val,
		time: time,
	}
}

type LinkedList struct {
	head, tail *LFUNode
}

func newLinkedList() *LinkedList {
	head, tail := &LFUNode{}, &LFUNode{}
	head.next = tail
	tail.pre = head
	return &LinkedList{
		head: head,
		tail: tail,
	}
}

func (this *LinkedList) moveToHead(node *LFUNode) {
	if node.pre != nil {
		node.pre.next = node.next
		node.next.pre = node.pre
	}
	this.head.next.pre = node
	node.next = this.head.next
	node.pre = this.head
	this.head.next = node
}

func (this *LinkedList) removeNode(removeNode *LFUNode) {
	removeNode.pre.next = removeNode.next
	removeNode.next.pre = removeNode.pre
	removeNode.pre = nil
	removeNode.next = nil
	return
}

func (this *LinkedList) removeTail() *LFUNode {
	removeNode := this.tail.pre
	if removeNode == this.head {
		return nil
	}
	removeNode.pre.next = this.tail
	this.tail.pre = removeNode.pre
	removeNode.pre = nil
	removeNode.next = nil
	return removeNode
}

func (this *LinkedList) isEmpty() bool {
	return this.head.next == this.tail
}

type sortList []int

func (t *sortList) lte(val int) int {
	l := len(*t)
	left, right := 0, l-1
	for left < right {
		mid := (left + right + 1) >> 1
		if (*t)[mid] > val {
			right = mid - 1
		} else {
			left = mid
		}
	}
	return left
}
func (t *sortList) add(val int) {
	idx := t.lte(val)
	// 已经存在 val，那么不需要再插入
	if idx < len(*t) && (*t)[idx] == val {
		return
	}
	if idx == len(*t) {
		*t = append(*t, val)
		return
	}
	*t = append((*t)[0:idx+1], val)
	*t = append(*t, (*t)[idx+2:]...)
}

func (t *sortList) delete(val int) {
	idx := t.lte(val)
	// 不存在 val，那么不需要删除
	if idx == len(*t) || idx < len(*t) && (*t)[idx] != val {
		return
	}
	*t = append((*t)[:idx], (*t)[idx+1:]...)
}

func (t *sortList) getMin() int {
	if t == nil || len(*t) <= 0 {
		return -1
	}
	return (*t)[0]
}

type LFUCache struct {
	/*
	   需要记录 key 对应的访问次数，相同访问次数的 key 根据最近访问时间来进行排序
	   使用一个 map + linkedlist 来记录访问次数对应的链表，同时使用一个 int 字段来记录最小的访问次数，
	   方便获取对应次数的链表，因为到达 cap 时需要移除的是最小访问次数的链表节点
	*/
	cap        int // 容量
	size       int // 插入的 key 个数
	head, tail *LFUNode
	nodeMap    map[string]*LFUNode // 存储 key-Node
	timeMap    map[int]*LinkedList // 存储 time-list
	times      *sortList           // 存储 key 的访问次数，是一个有序列表
	lock       *sync.Mutex
}

func LFUConstructor(capacity int) LFUCache {
	head, tail := &LFUNode{}, &LFUNode{}
	head.next = tail
	tail.pre = head
	return LFUCache{
		cap:     capacity,
		size:    0,
		nodeMap: map[string]*LFUNode{},
		timeMap: map[int]*LinkedList{},
		times:   &sortList{},
		lock:    &sync.Mutex{},
	}
}

func (this *LFUCache) Get(key string) interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	//
	node := this.nodeMap[key]
	if node == nil {
		return -1
	}
	// 获取 node 的访问次数 time
	time := node.time
	// 将 node 从当前 time list 中移除，同时判断当前 time list 是否已经为空，并且为空并且 time == minTime，那么将 minTime++
	ll := this.timeMap[time]
	ll.removeNode(node)
	if ll.isEmpty() {
		// 删除 time
		this.times.delete(time)
		delete(this.timeMap, time)
	}
	time++
	// 将 node.time+1
	node.time = time
	// 记录 time
	this.times.add(time)
	// 将 node 添加到新的 list
	newLL := this.timeMap[time]
	if newLL == nil {
		newLL = newLinkedList()
		this.timeMap[time] = newLL
	}
	newLL.moveToHead(node)
	return node.val
}

func (this *LFUCache) Set(key string, value interface{}) bool {
	if this.cap == 0 {
		return false
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	node := this.nodeMap[key]
	if node == nil {
		// 判断是否已经超过 cap
		if this.size == this.cap {
			minTime := this.getMinTime()
			// 超过的话需要将 minTime 对应的 list 移除末尾的节点
			ll := this.timeMap[minTime]
			// 将节点从 nodeMap 删除
			removeNode := ll.removeTail()
			delete(this.nodeMap, removeNode.key)
			// 再判断 list 是否为空，为空就去除
			if ll.isEmpty() {
				// 删除 time
				this.times.delete(minTime)
				// 删除 list
				delete(this.timeMap, minTime)
			}
		} else {
			this.size++
		}
		// 记录 time
		this.times.add(1)
		time := 1
		ll := this.timeMap[time]
		if ll == nil {
			ll = newLinkedList()
			this.timeMap[time] = ll
		}
		// 构造 node
		node = newLFUNode(key, value, 1)
		// 将 node 存储到 nodeMap
		this.nodeMap[key] = node
		// 将 node 添加到 time+1 list
		ll.moveToHead(node)
		return true
	}
	// 更新 node.val
	node.val = value
	// 获取 node 的访问次数 time
	time := node.time
	// 将 node 从当前 time list 中移除，同时判断当前 time list 是否已经为空，并且为空并且 time == minTime，那么将 minTime++
	ll := this.timeMap[time]
	ll.removeNode(node)
	if ll.isEmpty() {
		// 删除 time
		this.times.delete(time)
		// 删除 list
		delete(this.timeMap, time)
	}
	// time+1
	time++
	// 更新 node time
	node.time = time
	// 记录 time
	this.times.add(time)
	// 将 node 添加到新的 list
	newLL := this.timeMap[time]
	if newLL == nil {
		newLL = newLinkedList()
		this.timeMap[time] = newLL
	}
	newLL.moveToHead(node)
	return true
}

func (this *LFUCache) Delete(key string) bool {
	/*
		删除某个 key
		1、根据 key 获取 node
		2、根据 time 获取 list
		3、从 list 中删除 node
		4、如果 list 为空了，并且 time == minTime，那么将 minTime 设置为下一个更大的 time

		如何获取下一个 time ？ 用有序的 slice 存储当前存在的 time，添加或者删除的时候用二分
	*/
	this.lock.Lock()
	defer this.lock.Unlock()
	// 找到 key 对应 Node
	node := this.nodeMap[key]
	if node == nil {
		return true
	}
	delete(this.nodeMap, node.key)

	time := node.time
	// 根据 time 获取所在 list
	ll := this.timeMap[time]
	// 移除 node
	ll.removeNode(node)
	// list 为空
	if ll.isEmpty() {
		// 删除 time
		this.times.delete(time)
		// 删除 list
		delete(this.timeMap, time)
	}
	return true
}

func (this *LFUCache) getMinTime() int {
	return this.times.getMin()
}

func main() {

}
