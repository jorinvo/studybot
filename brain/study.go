package brain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/boltdb/bolt"
)

// GetStudy returns the current study the user needs to do.
func (store Store) GetStudy(chatID int64) (Study, error) {
	var study Study
	err := store.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketStudytimes).Cursor()
		now := time.Now().Unix()
		prefix := itob(chatID)
		total := 0
		var keyTime int64
		var key []byte

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp, err := btoi(v)
			if err != nil {
				return err
			}
			if timestamp < keyTime || keyTime == 0 {
				keyTime = timestamp
				key = k
			}
			if timestamp <= now {
				total++
			}
		}

		// No studies found
		if total == 0 {
			if keyTime > 0 {
				study = Study{Next: time.Second * time.Duration(keyTime-now)}
			}
			return nil
		}

		// Get study from phrase
		var p Phrase
		if err := json.Unmarshal(tx.Bucket(bucketPhrases).Get(key), &p); err != nil {
			return err
		}
		study = Study{
			Phrase:      p.Phrase,
			Explanation: p.Explanation,
			Total:       total,
		}
		return nil
	})

	if err != nil {
		return study, fmt.Errorf("failed to study with chatID %d: %v", chatID, err)
	}
	return study, nil
}

// ScoreStudy sets the score of the current study and moves to the next study.
func (store Store) ScoreStudy(chatID int64, score int) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		bs := tx.Bucket(bucketStudytimes)
		c := bs.Cursor()
		now := time.Now()
		uNow := now.Unix()
		prefix := itob(chatID)
		var keyTime int64
		var key []byte

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp, err := btoi(v)
			if err != nil {
				return err
			}
			if timestamp > uNow {
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

		// Get phrase
		var p Phrase
		bp := tx.Bucket(bucketPhrases)
		if err := json.Unmarshal(bp.Get(key), &p); err != nil {
			return err
		}

		// Update score
		p.Score += score
		newScore := p.Score
		if newScore < 0 {
			newScore = 0
		}

		// Save phrase
		buf, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err = bp.Put(key, buf); err != nil {
			return err
		}

		// Update study time
		// Randomize order by spreading studies over a period of time
		diffusion := time.Duration(rand.Intn(studyTimeDiffusion)) * time.Minute
		next := itob(now.Add((baseStudytime<<uint(newScore))*time.Hour + diffusion).Unix())
		if err = bs.Put(key, next); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to study with chatID %d: %v", chatID, err)
	}
	return nil
}

// GetNotifyTime gets the time until the user should be notified to study.
// Returns the time until the next studies are ready and a count of the ready studies.
// The returned duration is always at least dueMinInactive.
// The count is 0 if the chat has no phrases yet.
func (store Store) GetNotifyTime(chatID int64) (time.Duration, int, error) {
	due := 0
	now := time.Now()
	minTime := now.Add(dueMinInactive).Unix()
	var next sortableInts

	err := store.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketStudytimes).Cursor()
		prefix := itob(chatID)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp, err := btoi(v)
			if err != nil {
				return err
			}
			if timestamp < minTime {
				due++
			}
			if due >= dueMinCount {
				continue
			}
			l := len(next)
			if l < dueMinCount {
				next = append(next, timestamp)
				sort.Sort(next)
				continue
			}
			if timestamp < next[l-1] {
				next = append(next[:l-1], timestamp)
				sort.Sort(next)
			}
		}
		return nil
	})

	if err != nil {
		return 0, 0, fmt.Errorf("failed to get next studies for chat %d: %v", chatID, err)
	}

	minCount := dueMinCount
	l := len(next)
	if minCount > l {
		minCount = l
	}
	if due >= minCount {
		return dueMinInactive, due, nil
	}
	return time.Unix(next[l-1], 0).Sub(now), l, nil
}

// EachActiveChat runs a function for each chat
// where the user has been active since the last notification has been sent.
func (store Store) EachActiveChat(fn func(int64)) error {
	return store.db.View(func(tx *bolt.Tx) error {
		active := tx.Bucket(bucketActivities)
		return tx.Bucket(bucketReads).ForEach(func(k, v []byte) error {
			timeRead, err := btoi(v)
			if err != nil {
				return err
			}
			a := active.Get(k)
			if a == nil {
				a = itob(0)
			}
			timeActive, err := btoi(a)
			if err != nil {
				return err
			}
			if timeRead > timeActive {
				chatID, err := btoi(k)
				if err != nil {
					return err
				}
				fn(chatID)
			}
			return nil
		})
	})
}

type sortableInts []int64

func (b sortableInts) Len() int {
	return len(b)
}

func (b sortableInts) Less(i, j int) bool {
	return b[i] < b[j]
}

func (b sortableInts) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}
