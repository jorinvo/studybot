package brain

import (
	"fmt"

	"github.com/boltdb/bolt"
)

// GetMode fetches the mode for a chat.
func (store Store) GetMode(chatID int64) (Mode, error) {
	var mode Mode
	err := store.db.View(func(tx *bolt.Tx) error {
		if bm := tx.Bucket(bucketModes).Get(itob(chatID)); bm != nil {
			iMode, err := btoi(bm)
			if err != nil {
				return err
			}
			mode = Mode(iMode)
		} else {
			mode = ModeGetStarted
		}
		return nil
	})
	if err != nil {
		return mode, fmt.Errorf("failed to get mode for chatID %d: %v", chatID, err)
	}
	return mode, nil
}

// SetMode updates the mode for a chat.
func (store Store) SetMode(chatID int64, mode Mode) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketModes).Put(itob(chatID), itob(int64(mode)))
	})
	if err != nil {
		return fmt.Errorf("failed to set mode for chatID %d: %d: %v", chatID, mode, err)
	}
	return nil
}
