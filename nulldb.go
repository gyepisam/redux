package redo

// NullDb is a blackhole database, used for source files outside the redo project directory structure..
// It never fails, all writes disappear, and reads return nothing.
type NullDb struct {
}

// Open requires a project root argument
func NullDbOpen(ignored string) (DB, error) {
  return &NullDb{}, nil
}

func (db *NullDb) IsNull() bool { return true }

// Put stores value under key
func (db *NullDb) Put(key string, value []byte) error {
  return nil
}

// Get returns the value stored under key and a boolean indicating
// whether the returned value was actually found in the database.
func (db *NullDb) Get(key string) ([]byte, bool, error) {
	return []byte{}, false, nil
}

// Delete removes the value stored under key.
// If key does not exist, the operation is a noop.
func (db *NullDb) Delete(key string) error {
  return nil
}

// GetKeys returns a list of keys which have the specified key as a prefix.
func (db *NullDb) GetKeys(prefix string) ([]string, error) {
  return []string{}, nil
}

// GetValues returns a list of values whose keys matching the specified key prefix.
func (db *NullDb) GetValues(prefix string) ([][]byte, error) {
  return [][]byte{}, nil
}

// GetRecords returns a list of records (keys and data) matchign the specified key prefix.
func (db *NullDb) GetRecords(prefix string) ([]Record, error) {
  return []Record{}, nil
}

func (db *NullDb) Close() error {
  return nil
}
