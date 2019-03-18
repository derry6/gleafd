package snowflake

type Metadata struct {
	Name      string // 服务名称
	Addr      string // 监听IP:PORT
	MachineID int    // 机器ID 0 - 1023
	Timestamp int64  // 最后更新的时间
}
