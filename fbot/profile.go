package fbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// URL to fetch the profile from;
// is relative to the API URL.
const profileURL = "%s/%d?fields=first_name,locale,timezone&access_token=%s"

// Profile has all public user information we need;
// needs to be in sync with the URL abouve.
type Profile struct {
	Name     string  `json:"first_name"`
	Locale   string  `json:"locale"`
	Timezone float64 `json:"timezone"`
}

// GetProfile fetches a user profile for an ID.
func (c Client) GetProfile(id int64) (Profile, error) {
	var p Profile

	url := fmt.Sprintf(profileURL, c.api, id, c.token)
	resp, err := http.Get(url)
	if err != nil {
		return p, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p, err
	}

	if err = json.Unmarshal(content, &p); err != nil {
		return p, err
	}

	return p, checkError(bytes.NewReader(content))
}
