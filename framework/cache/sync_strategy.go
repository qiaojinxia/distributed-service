package cache

// SyncStrategy 同步策略
type SyncStrategy string

const (
	// SyncStrategyWriteThrough 写透策略 - 写入时同时写入L1和L2
	SyncStrategyWriteThrough SyncStrategy = "write_through"
	
	// SyncStrategyWriteBack 写回策略 - 只写入L1，定期写回L2
	SyncStrategyWriteBack SyncStrategy = "write_back"
	
	// SyncStrategyWriteAround 写绕过策略 - 写入时只写L2，绕过L1
	SyncStrategyWriteAround SyncStrategy = "write_around"
)