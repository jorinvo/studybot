package brain

import (
	"fmt"

	"github.com/boltdb/bolt"
)

// IsSubscribed checks if a user has notifications enabled.
func (store Store) IsSubscribed(chatID int64) (bool, error) {
	var isSubscribed bool
	err := store.db.View(func(tx *bolt.Tx) error {
		isSubscribed = tx.Bucket(bucketSubscriptions).Get(itob(chatID)) != nil
		return nil
	})
	if err != nil {
		return isSubscribed, fmt.Errorf("failed to check subscription for chat %d: %v", chatID, err)
	}
	return isSubscribed, nil
}

// Subscribe enables notifications for a user.
func (store Store) Subscribe(chatID int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSubscriptions).Put(itob(chatID), []byte{'1'})
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe chatID %d: %v", chatID, err)
	}
	return nil
}

// Unsubscribe disables notifications for a user.
func (store Store) Unsubscribe(chatID int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSubscriptions).Delete(itob(chatID))
	})
	if err != nil {
		return fmt.Errorf("failed to unsubscribe chatID %d: %v", chatID, err)
	}
	return nil
}
