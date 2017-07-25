package fbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// URL to send settings to;
// is relative to the API URL.
const settingsURL = "%s/me/messenger_profile?access_token=%s"

// SetGreetings sets the text displayed in the bot description.
// Pass a map of locale to greeting text.
// Include "default" locale as fallback for missing locales.
func (c Client) SetGreetings(greetings map[string]string) error {
	g := []greeting{}
	for k, v := range greetings {
		g = append(g, greeting{Locale: k, Text: v})
	}
	return c.postSetting(greetingSettings{Greeting: g})
}

// SetGetStartedPayload displays a "Get Started" button for new users.
// When a users pushes the button, a postback with the given payload is triggered.
func (c Client) SetGetStartedPayload(p string) error {
	return c.postSetting(getStartedSettings{GetStarted: getStartedPayload{p}})
}

// Helper to send settings to the settings endpoint.
func (c Client) postSetting(data interface{}) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(settingsURL, c.api, c.token)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return checkError(resp.Body)
}

type greetingSettings struct {
	Greeting []greeting `json:"greeting,omitempty"`
}

type greeting struct {
	Locale string `json:"locale"`
	Text   string `json:"text"`
}

type getStartedSettings struct {
	GetStarted getStartedPayload `json:"get_started,omitempty"`
}

type getStartedPayload struct {
	Payload string `json:"payload,omitempty"`
}
