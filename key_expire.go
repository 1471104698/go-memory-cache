package main

/*
	过期 key 处理策略：
	1、通知直接删除
	2、get 时再删除
	3、定时任务随机扫描部分 key 删除（洗牌算法）
*/
