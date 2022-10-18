package main

import (
	"sync"
	"time"
)

/*
	key 淘汰策略：
	1、LRU
	2、LFU
*/

type KeyMgr interface {
	Get(key string) interface{}
	Set(key string, value interface{}, expire int64) bool
	Delete(key string) bool
	ScanAndDelete(keyPercent float64) // 扫描并删除过期 key
}

type LRUNode struct {
	key       string
	val       interface{}
	pre, next *LRUNode
}

func newLRUNode(key string, value interface{}) *LRUNode {
	return &LRUNode{
		key: key,
		val: value,
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
	nodeMap    map[string]*LRUNode // 存储所有 node
	expireMap  map[string]int64    // 存储设置 key 对应的过期时间
	lock       *sync.Mutex
}

func LRUConstructor(capacity int) *LRUCache {
	head, tail := emptyLRUNode(), emptyLRUNode()
	head.next = tail
	tail.pre = head
	return &LRUCache{
		cap:       capacity,
		size:      0,
		head:      head,
		tail:      tail,
		nodeMap:   map[string]*LRUNode{},
		expireMap: map[string]int64{},
		lock:      &sync.Mutex{},
	}
}

func (this *LRUCache) Get(key string) interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	node := this.nodeMap[key]
	if node == nil {
		return -1
	}
	// 判断是否过期，如果过期那么删除
	if expire, ok := this.expireMap[key]; ok {
		if expire > 0 && expire < time.Now().Unix() {
			this.deleteAndClean(node)
			return -1
		}
	}
	moveToHead(this, node)
	return node.val
}

func (this *LRUCache) Set(key string, value interface{}, expire int64) bool {
	if this.cap == 0 {
		return true
	}
	if key == "" {
		return false
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	// 从 map 中获取
	node := this.nodeMap[key]
	if node == nil {
		node = newLRUNode(key, value)
		this.nodeMap[key] = node
		removeNode := addNode(this, node)
		this.deleteAndClean(removeNode)
	} else {
		node.val = value // 由于是指针，map 里也会更新
		moveToHead(this, node)
	}
	if expire > 0 {
		this.expireMap[key] = getExpireTime(expire)
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
	this.deleteAndClean(node)
	return true
}

func (this *LRUCache) ScanAndDelete(keyPercent float64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	keySize := int(float64(len(this.expireMap))*keyPercent) + 1
	for keySize > 0 {
		var deleteKey string
		// 利用 for range 返回 key 的随机性来进行随机 key 的扫描
		for key, expire := range this.expireMap {
			if expire < time.Now().Unix() {
				deleteKey = key
				break
			}
		}
		// 如果 key 不为空，那么进行删除，如果为空，那么表示不存在过期 key，直接返回
		if deleteKey != "" {
			node := this.nodeMap[deleteKey]
			this.deleteAndClean(node)
		} else {
			break
		}
		keySize--
	}
}

func (this *LRUCache) deleteAndClean(node *LRUNode) {
	if node == nil {
		return
	}
	deleteNode(this, node)
	delete(this.nodeMap, node.key)
	delete(this.expireMap, node.key)
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

func deleteNode(this *LRUCache, node *LRUNode) {
	if node.pre == nil || node.next == nil {
		return
	}
	node.pre.next = node.next
	node.next.pre = node.pre
	node.pre = nil
	node.next = nil
	this.size--
}

// -------------------------------------------------------- LFU start -----------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------------------------------------

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
	if removeNode.pre == nil || removeNode.next == nil {
		return
	}
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
	expireMap  map[string]int64    // 存储设置 key 对应的过期时间
	lock       *sync.Mutex
}

func LFUConstructor(capacity int) *LFUCache {
	head, tail := &LFUNode{}, &LFUNode{}
	head.next = tail
	tail.pre = head
	return &LFUCache{
		cap:       capacity,
		size:      0,
		nodeMap:   map[string]*LFUNode{},
		timeMap:   map[int]*LinkedList{},
		times:     &sortList{},
		expireMap: map[string]int64{},
		lock:      &sync.Mutex{},
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
	// 判断是否过期，如果过期那么删除
	if expire, ok := this.expireMap[key]; ok {
		if expire > 0 && expire < time.Now().Unix() {
			this.deleteAndClean(node, true)
			return -1
		}
	}
	this.deleteAndClean(node, false)
	node.time++
	this.addNewNode(node)
	return node.val
}

func (this *LFUCache) Set(key string, value interface{}, expire int64) bool {
	if this.cap == 0 {
		return true
	}
	if key == "" {
		return false
	}
	if expire > 0 {
		this.expireMap[key] = getExpireTime(expire)
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
			this.deleteAndClean(removeNode, true)
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
	this.deleteAndClean(node, false)
	node.val = value
	node.time++
	return this.addNewNode(node)
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
	this.deleteAndClean(node, true)
	return true
}

func (this *LFUCache) ScanAndDelete(keyPercent float64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	keySize := int(float64(len(this.expireMap))*keyPercent) + 1
	for keySize > 0 {
		var deleteKey string
		// 利用 for range 返回 key 的随机性来进行随机 key 的扫描
		for key, expire := range this.expireMap {
			if expire < time.Now().Unix() {
				deleteKey = key
				break
			}
		}
		// 如果 key 不为空，那么进行删除，如果为空，那么表示不存在过期 key，直接返回
		if deleteKey != "" {
			node := this.nodeMap[deleteKey]
			this.deleteAndClean(node, true)
		} else {
			break
		}
		keySize--
	}
}

func (this *LFUCache) deleteAndClean(node *LFUNode, isClean bool) {
	if node == nil {
		return
	}
	if isClean {
		delete(this.nodeMap, node.key)
		delete(this.expireMap, node.key)
	}
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
}

func (this *LFUCache) addNewNode(node *LFUNode) bool {
	// 更新 node.val
	time := node.time
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

func (this *LFUCache) getMinTime() int {
	return this.times.getMin()
}

func getExpireTime(expire int64) int64 {
	return time.Now().Add(time.Second * time.Duration(expire)).Unix()
}
