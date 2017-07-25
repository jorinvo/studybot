package messenger

import "github.com/jorinvo/studybot/fbot"

const (
	iconOK     = "\U0001F44C"
	iconDelete = "\u274C"
)

var (
	buttonStudyDone = fbot.Button{Text: "done studying", Payload: payloadStartMenu}
	// School emoji
	buttonStudy = fbot.Button{Text: "\U0001F3EB study", Payload: payloadStartStudy}
	// Plus sign emoji
	buttonAdd = fbot.Button{Text: "\u2795 phrases", Payload: payloadStartAdd}
	// Waving hand emoji
	buttonDone   = fbot.Button{Text: "\u2714 done", Payload: payloadIdle}
	buttonHelp   = fbot.Button{Text: "\u2753 help", Payload: payloadShowHelp}
	buttonDelete = fbot.Button{Text: iconDelete, Payload: payloadDelete}
)

var (
	buttonsMenuMode = []fbot.Button{
		buttonStudy,
		buttonAdd,
		buttonHelp,
		buttonDone,
	}
	buttonsSubscribe = []fbot.Button{
		fbot.Button{Text: iconOK + " sounds good", Payload: payloadSubscribe},
		fbot.Button{Text: "no thanks", Payload: payloadNoSubscription},
	}
	buttonsHelp = []fbot.Button{
		fbot.Button{Text: "stop notifications", Payload: payloadUnsubscribe},
		fbot.Button{Text: "send feedback", Payload: payloadFeedback},
		fbot.Button{Text: "all good", Payload: payloadStartMenu},
	}
	buttonsFeedback = []fbot.Button{
		fbot.Button{Text: iconDelete + " cancel", Payload: payloadStartMenu},
	}
	buttonsAddMode = []fbot.Button{
		fbot.Button{Text: "stop adding", Payload: payloadStartMenu},
	}
	buttonsStudyMode = []fbot.Button{
		buttonStudyDone,
	}
	buttonsShow = []fbot.Button{
		buttonDelete,
		buttonStudyDone,
		fbot.Button{Text: "\U0001F449 show phrase", Payload: payloadShowStudy},
	}
	buttonsScore = []fbot.Button{
		buttonDelete,
		// Thumb down emoji
		fbot.Button{Text: "\U0001F44E didn't know", Payload: payloadScoreBad},
		// Thinking face emoji
		fbot.Button{Text: "\U0001F914", Payload: payloadScoreOk},
		fbot.Button{Text: iconOK + " got it", Payload: payloadScoreGood},
	}
	buttonsStudyEmpty = []fbot.Button{
		buttonAdd,
		// buttonHelp,
	}
	buttonsStudiesDue = []fbot.Button{
		buttonStudy,
		fbot.Button{Text: "not now", Payload: payloadStartMenu},
	}
	buttonsConfirmDelete = []fbot.Button{
		fbot.Button{Text: iconDelete + " delete phrase", Payload: payloadConfirmDelete},
		fbot.Button{Text: "cancel", Payload: payloadCancelDelete},
	}
)
