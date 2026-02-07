package facts

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

var factsBucket = []byte("facts")

// FactStore persists per-user facts in a bbolt database.
type FactStore struct {
	db *bolt.DB
}

// NewFactStore opens or creates a bbolt database at the given path.
func NewFactStore(path string) (*FactStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(factsBucket)
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create bucket: %w", err)
	}

	return &FactStore{db: db}, nil
}

// GetFacts returns the stored facts for a user, or nil if none exist.
func (s *FactStore) GetFacts(userID string) ([]string, error) {
	var facts []string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(factsBucket)
		data := b.Get([]byte(userID))
		if data == nil {
			return nil
		}
		return json.Unmarshal(data, &facts)
	})
	return facts, err
}

// AppendFacts appends new facts to the user's existing facts and returns the total count.
func (s *FactStore) AppendFacts(userID string, newFacts []string) (int, error) {
	var total int
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(factsBucket)
		var existing []string
		data := b.Get([]byte(userID))
		if data != nil {
			if err := json.Unmarshal(data, &existing); err != nil {
				return err
			}
		}
		existing = append(existing, newFacts...)
		total = len(existing)
		encoded, err := json.Marshal(existing)
		if err != nil {
			return err
		}
		return b.Put([]byte(userID), encoded)
	})
	return total, err
}

// SetFacts replaces a user's facts entirely (used for compaction).
func (s *FactStore) SetFacts(userID string, facts []string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(factsBucket)
		encoded, err := json.Marshal(facts)
		if err != nil {
			return err
		}
		return b.Put([]byte(userID), encoded)
	})
}

// Close closes the underlying database.
func (s *FactStore) Close() error {
	return s.db.Close()
}
