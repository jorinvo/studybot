package brain

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

// BackupTo streams backup as an HTTP response.
func (store Store) BackupTo(w http.ResponseWriter) {
	err := store.db.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="my.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// StudyNow resets all study times of all users to now.
func (store Store) StudyNow() error {
	return store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketStudytimes)
		now := itob(time.Now().Unix())
		return b.ForEach(func(k, v []byte) error {
			return b.Put(k, now)
		})
	})
}

// DeleteChat removes all records of a given chat.
func (store Store) DeleteChat(chatID int64) error {
	return store.db.Update(func(tx *bolt.Tx) error {
		key := itob(chatID)
		// Remove mode
		if err := tx.Bucket(bucketModes).Delete(key); err != nil {
			return err
		}
		// Remove phrases
		bp := tx.Bucket(bucketPhrases)
		c := bp.Cursor()
		for k, _ := c.Seek(key); k != nil && bytes.HasPrefix(k, key); k, _ = c.Next() {
			if err := bp.Delete(k); err != nil {
				return err
			}
		}
		// Remove study times
		bs := tx.Bucket(bucketStudytimes)
		c = bp.Cursor()
		for k, _ := c.Seek(key); k != nil && bytes.HasPrefix(k, key); k, _ = c.Next() {
			if err := bs.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetChatIDs returns chatIDs of all users.
func (store Store) GetChatIDs() ([]int64, error) {
	var ids []int64
	err := store.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketModes).ForEach(func(k, v []byte) error {
			id, err := btoi(k)
			if err != nil {
				return err
			}
			ids = append(ids, id)
			return nil
		})
	})
	return ids, err
}

// GetPhrasesAsJSON ...
func (store Store) GetPhrasesAsJSON(chatID int64) (io.Reader, error) {
	var phrases string
	err := store.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketPhrases).Cursor()
		prefix := itob(chatID)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			phrases += string(v) + "\n"
		}
		return nil
	})
	return strings.NewReader(phrases), err
}

// DeletePhrases removes all phrases fn matches.
func (store Store) DeletePhrases(fn func(int64, Phrase) bool) (int, error) {
	deleted := 0
	err := store.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket(bucketPhrases)
		bs := tx.Bucket(bucketStudytimes)
		return bp.ForEach(func(k, v []byte) error {
			id, err := btoi(k[:8])
			if err != nil {
				return err
			}
			var p Phrase
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			if !fn(int64(id), p) {
				return nil
			}
			if err := bs.Delete(k); err != nil {
				return err
			}
			if err := bp.Delete(k); err != nil {
				return err
			}
			deleted++
			return nil
		})
	})
	return deleted, err
}
