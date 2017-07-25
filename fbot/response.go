package fbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// URL to send messages to;
// is relative to the API URL.
const sendMessageURL = "%s/me/messages?access_token=%s"

// Send a text message with a set of quick reply buttons to a user.
func (c Client) Send(id int64, message string, buttons []Button) error {
	var replies []quickReply
	for _, b := range buttons {
		replies = append(replies, quickReply{
			ContentType: "text",
			Title:       b.Text,
			Payload:     b.Payload,
		})
	}
	m := sendMessage{
		Recipient: recipient{ID: id},
		Message: messageData{
			Text:         message,
			QuickReplies: replies,
		},
	}

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(sendMessageURL, c.api, c.token)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode == 200 {
		return nil
	}
	return checkError(resp.Body)
}

type sendMessage struct {
	Recipient recipient   `json:"recipient"`
	Message   messageData `json:"message"`
}

type messageData struct {
	Text         string       `json:"text"`
	QuickReplies []quickReply `json:"quick_replies,omitempty"`
}

type recipient struct {
	ID int64 `json:"id,string"`
}

type quickReply struct {
	ContentType string `json:"content_type,omitempty"`
	Title       string `json:"title,omitempty"`
	Payload     string `json:"payload"`
}
