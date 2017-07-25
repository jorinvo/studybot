package brain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// AddPhrase stores a new phrase.
func (store Store) AddPhrase(chatID int64, phrase, explanation string) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket(bucketPhrases)

		// Get phrase id
		sequence, err := bp.NextSequence()
		if err != nil {
			return err
		}
		prefix := itob(chatID)
		phraseID := append(prefix, itob(int64(sequence))...)

		// Phrase to JSON
		buf, err := json.Marshal(Phrase{Phrase: phrase, Explanation: explanation})
		if err != nil {
			return err
		}

		// Save Phrase
		if err = bp.Put(phraseID, buf); err != nil {
			return err
		}

		// Limit number of new studies per day
		newPhrases := 0
		c := tx.Bucket(bucketPhrases).Cursor()
		var p Phrase
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			if p.Score == 0 {
				newPhrases++
			}
		}

		// Save study time
		bs := tx.Bucket(bucketStudytimes)
		next := itob(time.Now().Add(time.Duration(newPhrases/newPerDay*24+firstStudytime) * time.Hour).Unix())
		return bs.Put(phraseID, next)
	})

	if err != nil {
		return fmt.Errorf("failed to add phrase for chatID %d: %s - %s: %v", chatID, phrase, explanation, err)
	}
	return nil
}

// FindPhrase returns a phrase belonging to the passed user that matches the passed function.
func (store Store) FindPhrase(chatID int64, fn func(Phrase) bool) (Phrase, error) {
	var p Phrase
	err := store.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketPhrases).Cursor()
		prefix := itob(chatID)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var tmp Phrase
			if err := json.Unmarshal(v, &tmp); err != nil {
				return err
			}
			if fn(tmp) {
				p = tmp
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return p, fmt.Errorf("failed to find phrase with chatid %d: %v", chatID, err)
	}
	return p, nil
}

// DeleteStudyPhrase deletes the phrase the passed user currently has to study.
func (store Store) DeleteStudyPhrase(chatID int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		bs := tx.Bucket(bucketStudytimes)
		c := bs.Cursor()
		now := time.Now().Unix()
		prefix := itob(chatID)
		var keyTime int64
		var key []byte

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp, err := btoi(v)
			if err != nil {
				return err
			}
			if timestamp > now {
				continue
			}
			if timestamp < keyTime || keyTime == 0 {
				keyTime = timestamp
				key = k
			}
		}

		// No studies found
		if key == nil {
			return errors.New("no study found")
		}

		// Delete study time
		if err := bs.Delete(key); err != nil {
			return err
		}

		// Delete phrase
		return tx.Bucket(bucketPhrases).Delete(key)
	})

	if err != nil {
		return fmt.Errorf("failed to delete study phrase for chatID %d: %v", chatID, err)
	}
	return nil
}
