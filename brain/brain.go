// Package brain handles all business logic.
// It handles data storage, retrieving and updating.
// It's independent from the used bot platform and user interaction.
package brain

import "time"

const (
	// Time to wait for first study in hours
	firstStudytime = 2
	// base time in hours to use to calculate next study time
	baseStudytime = 6
	// Time in minutes
	// When study times are updated they are randomly placed
	// somewhere between the new time and new time + studyTimeDiffusion
	// to mix up the order in which words are studied.
	studyTimeDiffusion = 30
	// Maximum number of new studies per day
	newPerDay = 30
	// Minimum number of studies needed to be due before notifying user
	dueMinCount = 9
	// Time user has to be inactive before being notified
	dueMinInactive = 10 * time.Minute
)

var (
	bucketModes         = []byte("modes")
	bucketPhrases       = []byte("phrases")
	bucketStudytimes    = []byte("studytimes")
	bucketReads         = []byte("reads")
	bucketActivities    = []byte("activities")
	bucketSubscriptions = []byte("subscriptions")
)

// Mode is the state of a chat.
// We need to keep track of the state each chat is in.
type Mode int

const (
	// ModeMenu shows the main menu.
	ModeMenu Mode = iota
	// ModeAdd lets the user add new phrases.
	ModeAdd
	// ModeStudy goes to phrases ready to study.
	ModeStudy
	// ModeGetStarted sends an introduction to the user.
	ModeGetStarted
	// ModeFeedback allows the user to send a message that is ready by a human.
	ModeFeedback
)

// Study is a study the current study the user needs to answer.
type Study struct {
	// Phrase is the phrase the user needs to guess.
	Phrase string
	// Explanation is the explanation displayed to the user.
	Explanation string
	// Total is the total number of studies ready, including the current one.
	Total int
	// Next contains the time until the next study is available;
	// it's only set if Total is 0.
	Next time.Duration
}

// Phrase describes a phrase the user saved.
type Phrase struct {
	Phrase      string
	Explanation string
	Score       int
}
