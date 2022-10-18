package main

import (
	"time"
)

/*
	过期 key 处理策略：
	1、绑定定时器，过期了直接调用 delete 函数删除
	2、get 时再删除
	3、定时任务随机扫描部分 key 删除（洗牌算法）
*/

const (
	TimeToDelete  = 1 << 0 // 定时删除
	LazyToDelete  = 1 << 1 // 惰性删除
	RegularlyScan = 1 << 2 // 定期扫描
)

const (
	LRU = 1
	LFU = 2
)

const (
	RegularlyScanTime = 100  // 定期扫描间隔，单位为 ms
	DeleteKeyPercent  = 0.25 // 每次扫描过期 key 的占比， 1/4
)

type MemoryCache struct {
	keyMgr KeyMgr
}

func MemoryCacheConstructor(cap int) *MemoryCache {
	c := &MemoryCache{
		keyMgr: LFUConstructor(cap),
	}
	// 开启线程定期扫描清除过期 key
	go func() {
		tick := time.Tick(RegularlyScanTime * time.Millisecond)
		for {
			select {
			case <-tick:
				c.keyMgr.ScanAndDelete(DeleteKeyPercent)
			}
		}
	}()
	return c
}

func (c *MemoryCache) Get(key string) interface{} {
	return c.keyMgr.Get(key)
}

func (c *MemoryCache) Set(key string, val interface{}, expire int64) bool {
	return c.keyMgr.Set(key, val, expire)
}

func (c *MemoryCache) Delete(key string) bool {
	return c.keyMgr.Delete(key)
}
