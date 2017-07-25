package messenger

import (
	"fmt"
	"time"

	"github.com/jorinvo/studybot/brain"
)

// Start a timer to notify the given chat.
// Only works when chat has notifications enabled
// and has added some phrases already.
func (b Bot) scheduleNotify(id int64) {
	if b.notifyTimers == nil {
		return
	}

	isSubscribed, err := b.store.IsSubscribed(id)
	if err != nil {
		b.err.Println(err)
		return
	}
	if !isSubscribed {
		return
	}

	if timer := b.notifyTimers[id]; timer != nil {
		// Don't care if timer is active or not
		_ = timer.Stop()
	}
	d, count, err := b.store.GetNotifyTime(id)
	if err != nil {
		b.err.Println(err)
		return
	}
	if count == 0 {
		return
	}

	b.info.Printf("Notify %d in %s with %d due studies", id, d.String(), count)
	b.notifyTimers[id] = time.AfterFunc(d, func() {
		b.notify(id, count)
	})
}

func (b Bot) notify(id int64, count int) {
	p, err := b.client.GetProfile(id)
	name := p.Name
	if err != nil {
		name = "there"
		b.err.Printf("failed to get profile for %d: %v", id, err)
	}
	msg := fmt.Sprintf(messageStudiesDue, name, count)
	if err := b.store.SetMode(id, brain.ModeMenu); err != nil {
		b.err.Printf("failed to activate menu mode while notifying %d: %v", id, err)
	}
	if err = b.client.Send(id, msg, buttonsStudiesDue); err != nil {
		b.err.Printf("failed to notify user %d: %v", id, err)
	}
	b.info.Printf("Notified %s (%d) with %d due studies", name, id, count)
	// Track last sending of a notification
	// to stop sending notifications
	// when user hasn't read the last notification.
	if err := b.store.SetActivity(id, time.Now()); err != nil {
		b.err.Println(err)
	}
}
