package executor

// TestSSH 对全部 hosts 执行连接——鉴权检查，任一失败则返回聚合错误
// 使用场景：批量执行命令前检查，
/*
	参数说明：
		- ctx: 上下文，用于取消和超时控制
		- hosts: 目标主机列表
		- opts: 并发数和SSH连接参数(config 模式和携带私钥认证)
*/

func