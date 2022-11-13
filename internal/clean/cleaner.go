package clean

type SystemDataCleaner struct {
	//sync.RWMutex
}

// Clean system data.
func (cleaner *SystemDataCleaner) Clean() error {
	// 开始清理时的游标，  keyspace中key的数量
	return nil
}

// Report task status to snrs
func (cleaner *SystemDataCleaner) Report() error {
	return nil
}

func NewCleaner() (SystemDataCleaner, error) {
	cleaner := SystemDataCleaner{}
	return cleaner, nil
}
