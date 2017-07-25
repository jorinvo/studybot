// Package fbot can be used to communicate with a Facebook Messenger bot.
// The supported API is limited to only the required use cases
// and the data format is abstracted accordingly.
package fbot

import (
	"encoding/json"
	"fmt"
	"io"
)

const defaultAPI = "https://graph.facebook.com/v2.6"

// Client can be used to communicate with a Messenger bot.
type Client struct {
	token string
	api   string
}

// API can be passed to New for sending requests to a different URL.
// Must not contain trailing slash.
func API(url string) func(*Client) {
	return func(c *Client) {
		c.api = url
	}
}

// New rerturns a new client with credentials set up.
func New(token string, options ...func(*Client)) Client {
	c := Client{
		token: token,
		api:   defaultAPI,
	}
	for _, option := range options {
		option(&c)
	}
	return c
}

// Button describes a text quick reply.
type Button struct {
	// Text is the text on the button visible to the user
	Text string
	// Payload is a string to identify the quick reply event internally in your application.
	Payload string
}

// Helper to check for errors in reply
func checkError(r io.Reader) error {
	var qr queryResponse
	err := json.NewDecoder(r).Decode(&qr)
	if qr.Error != nil {
		err = fmt.Errorf("Facebook error : %s", qr.Error.Message)
	}
	return err
}

type queryResponse struct {
	Error  *queryError `json:"error"`
	Result string      `json:"result"`
}

type queryError struct {
	Message   string `json:"message"`
	Type      string `json:"type"`
	Code      int    `json:"code"`
	FBTraceID string `json:"fbtrace_id"`
}
