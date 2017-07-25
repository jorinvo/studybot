// Package admin provides an admin server that can be used to make backups
// and to communicate with users via Slack.
package admin

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jorinvo/studybot/brain"
)

// Admin is a HTTP handler that can be used for backups
// and to communicate with users via Slack.
type Admin struct {
	store        brain.Store
	err          *log.Logger
	slackHook    string
	slackToken   string
	replyHandler func(int64, string) error
}

// SlackReply isa n option to enable /slack to receive replies from Slack.
// token is used to validate posts to the webhook.
// fn is called with a chatID and a message.
func SlackReply(token string, fn func(int64, string) error) func(*Admin) {
	return func(a *Admin) {
		a.slackToken = token
		a.replyHandler = fn
	}
}

// LogErr is an option to set the error logger.
func LogErr(l *log.Logger) func(*Admin) {
	return func(a *Admin) {
		a.err = l
	}
}

// New returns a new Admin which can be used as an http.Handler.
func New(store brain.Store, slackHook string, options ...func(*Admin)) Admin {
	a := Admin{
		store:     store,
		slackHook: slackHook,
	}
	for _, option := range options {
		option(&a)
	}
	if a.err == nil {
		a.err = log.New(ioutil.Discard, "", 0)
	}
	return a
}

// ServeHTTP serves the different endpoints the admin server provides.
func (a Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			a.err.Println(err)
		}
	}()

	switch r.URL.Path {
	case "/":
		_, err := w.Write([]byte(`
GET     /backup    Stream a backup of the current state of the database.
DELETE  /phrase    Delete phrases. Combine query parameters 'chatid', 'phrase', 'explanation' and 'score' to select phrases.
GET     /studynow  Reset all study times to now. Note that this doesn't reset the notification timers.
POST    /slack     Register in Slack as Outgoing Webhook to send responses back to users.
`))
		if err != nil {
			a.err.Println("failed to send '/' response")
		}
	case "/backup":
		if r.Method != "GET" {
			return
		}
		a.store.BackupTo(w)

	case "/phrase":
		if r.Method != "DELETE" {
			return
		}
		qChatID := r.URL.Query().Get("chatid")
		hasChatID := qChatID != ""
		phrase := r.URL.Query().Get("phrase")
		hasPhrase := phrase != ""
		explanation := r.URL.Query().Get("explanation")
		hasExplanation := explanation != ""
		score := r.URL.Query().Get("score")
		hasScore := score != ""
		if !(hasChatID || hasPhrase || hasExplanation || hasScore) {
			http.Error(w, "no query specified", 400)
			return
		}
		var chatID int64
		var err error
		if hasChatID {
			chatID, err = strconv.ParseInt(qChatID, 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid chatid: '%s'", qChatID), 400)
				return
			}
		}
		count, err := a.store.DeletePhrases(func(id int64, p brain.Phrase) bool {
			if hasChatID && id != chatID {
				return false
			}
			if hasPhrase && !strings.Contains(p.Phrase, phrase) {
				return false
			}
			if hasExplanation && !strings.Contains(p.Explanation, explanation) {
				return false
			}
			if hasScore && strconv.Itoa(int(p.Score)) != score {
				return false
			}
			return true
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed deleting phrases: %v", err), 500)
			return
		}
		fmt.Fprintf(w, "Deleted %d phrases.", count)

	case "/studynow":
		if r.Method != "GET" {
			return
		}
		if err := a.store.StudyNow(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "studies updated")

	case "/slack":
		if r.Method != "POST" {
			return
		}
		if a.slackToken == "" {
			slackError(w, fmt.Errorf("webhook is disabled"))
		}
		if r.FormValue("token") != a.slackToken {
			slackError(w, fmt.Errorf("invalid token"))
			return
		}
		if r.FormValue("bot_id") != "" {
			return
		}
		text := strings.TrimSpace(r.FormValue("text"))
		fields := strings.Fields(text)
		if len(fields) < 2 {
			slackError(w, fmt.Errorf("missing ID"))
			return
		}
		firstField := fields[0]
		id, err := strconv.Atoi(firstField)
		if err != nil {
			slackError(w, fmt.Errorf("failed parsing ID: %v", err))
			return
		}
		msg := strings.TrimSpace(strings.TrimPrefix(text, firstField))
		if err := a.replyHandler(int64(id), msg); err != nil {
			slackError(w, err)
			return
		}
	}
}

// HandleMessage can be called to send a user message to Slack.
func (a Admin) HandleMessage(id int64, name, msg string) {
	slackMsg := struct {
		Username string `json:"username"`
		Text     string `json:"text"`
	}{
		Username: name,
		Text:     fmt.Sprintf("%d\n\n%s", id, msg),
	}
	buf, err := json.Marshal(slackMsg)
	if err != nil {
		a.err.Printf("failed to marshal slack message (%v): %v", slackMsg, err)
		return
	}
	resp, err := http.Post(a.slackHook, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		a.err.Printf("failed to post message from %s (%d) to Slack: %s", name, id, msg)
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			a.err.Printf("failed to read response for Slack message '%s' from %s (%d): %v", msg, name, id, err)
			return
		}
		a.err.Printf("HTTP status code is not OK (%d) for Slack message '%s' from %s (%d): %s", resp.StatusCode, msg, name, id, body)
		return
	}
}

func slackError(w http.ResponseWriter, err error) {
	fmt.Fprint(w, fmt.Sprintf(`{ "text": "Error sending message: %s." }`, err))
}

// Currently not in use
func csvImport(store brain.Store, errLogger, infoLogger *log.Logger, toImport string) {
	// CSV import
	parts := strings.Split(toImport, ":")
	i, err := strconv.Atoi(parts[0])
	if err != nil {
		errLogger.Fatal(err)
	}
	chatID := int64(i)
	file := parts[1]
	infoLogger.Printf("Importing to chat ID %d from CSV file %s", chatID, file)
	f, err := os.Open(file)
	if err != nil {
		errLogger.Fatalln(err)
	}
	count := 0
	reader := csv.NewReader(f)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			infoLogger.Printf("Imported %d phrases", count)
			return
		}
		if err != nil {
			errLogger.Fatalln(err)
		}
		if len(row) != 2 {
			errLogger.Printf("line %d has wrong number of fields, expected 2, had %d", count+1, len(row))
		} else {
			count++
			p := strings.TrimSpace(row[0])
			e := strings.TrimSpace(row[1])
			if err = store.AddPhrase(chatID, p, e); err != nil {
				errLogger.Fatalln(err)
			}
		}
	}
}
