package main

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

type MemoryCache struct {
	keyMgr KeyMgr
}
