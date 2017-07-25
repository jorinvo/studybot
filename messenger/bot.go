package messenger

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/jorinvo/studybot/brain"
	"github.com/jorinvo/studybot/fbot"
)

// Everything that is not in the unicode character classes
// for letters or numeric values
// See: http://www.fileformat.info/info/unicode/category/index.htm
var specialChars = regexp.MustCompile(`[^\p{Ll}\p{Lm}\p{Lo}\p{Lu}\p{Nd}\p{Nl}\p{No}]`)

var inParantheses = regexp.MustCompile(`\(.*?\)`)

// HandleEvent handles a Messenger event.
func (b Bot) HandleEvent(e fbot.Event) {
	if e.Type == fbot.EventError {
		b.err.Println(e.Text)
		return
	}

	if e.Type == fbot.EventUnknown {
		b.err.Println("received unknown event", e)
		return
	}

	if e.Type == fbot.EventRead {
		if err := b.store.SetRead(e.ChatID, e.Time); err != nil {
			b.err.Println(err)
		}
		return
	}

	b.scheduleNotify(e.ChatID)

	if e.Type == fbot.EventPayload {
		b.handlePayload(e.ChatID, e.Payload)
		return
	}

	b.handleMessage(e.ChatID, e.Text)
}

func (b Bot) handleMessage(id int64, msg string) {
	mode, err := b.store.GetMode(id)
	if err != nil {
		b.send(id, messageErr, buttonsMenuMode, fmt.Errorf("failed to get mode for id %v: %v", id, err))
		return
	}
	switch mode {
	case brain.ModeStudy:
		study, err := b.store.GetStudy(id)
		if err != nil {
			b.send(id, messageErr, buttonsStudyMode, fmt.Errorf("failed to get study: %v", err))
			return
		}
		// Score user unput and pick appropriate reply
		msgNormalized := normPhrase(msg)
		if msgNormalized == "" {
			study, err := b.store.GetStudy(id)
			if err != nil {
				b.send(id, messageErr, buttonsShow, fmt.Errorf("failed to get study: %v", err))
				return
			}
			b.send(id, study.Phrase, buttonsScore, nil)
			return
		}
		var score = 1
		reply := messageStudyCorrect
		if msgNormalized != normPhrase(study.Phrase) {
			score = -1
			reply = fmt.Sprintf(messageStudyWrong, study.Phrase)
		}
		b.send(id, reply, nil, nil)
		b.send(b.scoreAndStudy(id, score))

	case brain.ModeAdd:
		parts := strings.SplitN(strings.TrimSpace(msg), "\n", 2)
		phrase := strings.TrimSpace(parts[0])
		if phrase == "" {
			b.send(id, messagePhraseEmpty, buttonsAddMode, nil)
			return
		}
		if len(parts) == 1 {
			b.send(id, messageExplanationEmpty, buttonsAddMode, nil)
			return
		}
		explanation := strings.TrimSpace(parts[1])
		// Check for existing explanation
		p, err := b.store.FindPhrase(id, func(p brain.Phrase) bool {
			return p.Explanation == explanation
		})
		if err != nil {
			b.send(id, messageErr, nil, fmt.Errorf("failed to lookup phrase: %v", err))
			return
		}
		if p.Phrase != "" {
			b.send(id, fmt.Sprintf(messageExplanationExists, p.Phrase, p.Explanation), buttonsAddMode, nil)
			return
		}
		// Save phrase
		if err = b.store.AddPhrase(id, phrase, explanation); err != nil {
			b.send(id, messageErr, buttonsAddMode, fmt.Errorf("failed to save phrase: %v", err))
			return
		}
		b.send(id, fmt.Sprintf(messageAddDone, phrase, explanation), nil, nil)
		b.send(id, messageAddNext, buttonsAddMode, nil)

	case brain.ModeGetStarted:
		b.messageWelcome(id)

	case brain.ModeFeedback:
		p, err := b.client.GetProfile(id)
		name := p.Name
		if err != nil {
			name = "there"
			b.err.Printf("failed to get profile for %d: %v", id, err)
		}
		if b.feedback != nil {
			b.feedback <- Feedback{ChatID: id, Username: name, Message: msg}
		} else {
			b.err.Printf("got unhandled feedback from %s (%d): %s", name, id, msg)
		}
		b.send(id, messageFeedbackDone, nil, nil)
		b.send(b.messageStartMenu(id))

	default:
		b.send(b.messageStartMenu(id))
	}
}

func (b Bot) handlePayload(id int64, payload string) {
	switch payload {
	case payloadGetStarted:
		b.messageWelcome(id)

	case payloadIdle:
		b.send(id, messageIdle, nil, nil)

	case payloadStartStudy:
		if err := b.store.SetMode(id, brain.ModeStudy); err != nil {
			b.send(id, messageErr, buttonsMenuMode, err)
			return
		}
		b.send(b.startStudy(id))

	case payloadStartAdd:
		if err := b.store.SetMode(id, brain.ModeAdd); err != nil {
			b.send(id, messageErr, buttonsMenuMode, err)
			return
		}
		b.send(id, messageStartAdd, buttonsAddMode, nil)

	case payloadShowHelp:
		isSubscribed, err := b.store.IsSubscribed(id)
		if err != nil {
			b.err.Println(err)
		}
		buttons := buttonsHelp
		if !isSubscribed {
			buttons = buttons[1:]
		}
		b.send(id, messageHelp, buttons, nil)

	case payloadShowStudy:
		study, err := b.store.GetStudy(id)
		if err != nil {
			b.send(id, messageErr, buttonsShow, fmt.Errorf("failed to get study: %v", err))
			return
		}
		b.send(id, study.Phrase, buttonsScore, nil)

	case payloadScoreBad:
		b.send(b.scoreAndStudy(id, -1))

	case payloadScoreOk:
		b.send(b.scoreAndStudy(id, 0))

	case payloadScoreGood:
		b.send(b.scoreAndStudy(id, 1))

	case payloadDelete:
		b.send(id, messageConfirmDelete, buttonsConfirmDelete, nil)

	case payloadConfirmDelete:
		if err := b.store.DeleteStudyPhrase(id); err != nil {
			b.send(id, messageErr, nil, nil)
		} else {
			b.send(id, messageDeleted, nil, nil)
		}
		b.send(b.startStudy(id))

	case payloadCancelDelete:
		b.send(id, messageCancelDelete, nil, nil)
		b.send(b.startStudy(id))

	case payloadSubscribe:
		if err := b.store.Subscribe(id); err != nil {
			b.send(id, messageErr, nil, nil)
			return
		}
		b.send(id, messageSubscribed, buttonsMenuMode, nil)

	case payloadUnsubscribe:
		if err := b.store.Unsubscribe(id); err != nil {
			b.send(id, messageErr, nil, nil)
			return
		}
		b.send(id, messageUnsubscribed, buttonsMenuMode, nil)

	case payloadNoSubscription:
		b.send(id, messageNoSubscription, buttonsMenuMode, nil)

	case payloadFeedback:
		if err := b.store.SetMode(id, brain.ModeFeedback); err != nil {
			b.send(id, messageErr, buttonsMenuMode, err)
			return
		}
		b.send(id, messageFedback, buttonsFeedback, nil)

	case payloadStartMenu:
		fallthrough
	default:
		b.send(b.messageStartMenu(id))
	}
}

func (b Bot) messageStartMenu(id int64) (int64, string, []fbot.Button, error) {
	if err := b.store.SetMode(id, brain.ModeMenu); err != nil {
		return id, messageErr, buttonsMenuMode, err
	}
	return id, messageStartMenu, buttonsMenuMode, nil
}

func (b Bot) messageWelcome(id int64) {
	p, err := b.client.GetProfile(id)
	name := p.Name
	if err != nil {
		name = "there"
		b.err.Printf("failed to get profile for %d: %v", id, err)
	}
	b.send(id, fmt.Sprintf(messageWelcome, name), nil, nil)
	time.Sleep(6 * time.Second)
	b.send(id, messageWelcome2, nil, b.store.SetMode(id, brain.ModeAdd))
}

func (b Bot) startStudy(id int64) (int64, string, []fbot.Button, error) {
	study, err := b.store.GetStudy(id)
	if err != nil {
		return id, messageErr, buttonsStudyMode, err
	}
	// No studies ready
	if study.Total == 0 {
		// Go to menu mode
		if err = b.store.SetMode(id, brain.ModeMenu); err != nil {
			return id, messageErr, buttonsStudyMode, err
		}
		// There are not studies yet
		if study.Next == 0 {
			return id, messageStudyEmpty, buttonsStudyEmpty, nil
		}
		// Display time until next study is ready
		msg := fmt.Sprintf(messageStudyDone, formatDuration(study.Next))
		isSubscribed, err := b.store.IsSubscribed(id)
		if err != nil {
			b.err.Println(err)
		}
		if isSubscribed || err != nil {
			return id, msg, buttonsMenuMode, nil
		}
		// Ask to subscribe to notifications
		return id, msg + messageAskToSubscribe, buttonsSubscribe, nil
	}
	// Send study to user
	return id, fmt.Sprintf(messageStudyQuestion, study.Total, study.Explanation), buttonsShow, nil
}

func (b Bot) scoreAndStudy(id int64, score int) (int64, string, []fbot.Button, error) {
	err := b.store.ScoreStudy(id, score)
	if err != nil {
		return id, messageErr, buttonsStudyMode, err
	}
	return b.startStudy(id)
}

// Send replies and log errors
func (b Bot) send(id int64, reply string, buttons []fbot.Button, err error) {
	if err != nil {
		b.err.Println(err)
	}
	if err = b.client.Send(id, reply, buttons); err != nil {
		b.err.Println("failed to send message:", err)
	}
}

// Format like "X hour[s] X minute[s]".
// Returns empty string for negativ durations.
func formatDuration(d time.Duration) string {
	// Precision in minutes
	d = time.Duration(math.Ceil(float64(d)/float64(time.Minute))) * time.Minute
	s := ""
	h := d / time.Hour
	m := (d - h*time.Hour) / time.Minute
	if h > 1 {
		s += fmt.Sprintf("%d", h) + " hours "
	} else if h == 1 {
		s += "1 hour "
	}
	if m > 1 {
		s += fmt.Sprintf("%d", m) + " minutes"
	} else if m > 0 {
		s += "1 minute"
	} else if s != "" {
		// No minutes, only hours, remove trailing space
		s = s[:len(s)-1]
	}
	return s
}
func normPhrase(s string) string {
	s = inParantheses.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return specialChars.ReplaceAllString(s, "")
}
