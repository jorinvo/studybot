package messenger

const (
	messageStartMenu = `What would you like to do next?
Please use the buttons below.`
	messageHelp     = "How can I help you?"
	messageIdle     = "Good, just send me a \U0001F44D to continue with your studies."
	messageStartAdd = `Please send me a phrase and its explanation.
Separate them with a linebreak.`
	messageWelcome = `Hello %s!

Whenever you pick up a new phrase, just add it to your Studybot and remember it forever.

You begin by adding phrases and later Studybot will test your memories in a natural schedule.`
	messageWelcome2 = messageStartAdd + `
Don't worry if you send something wrong. You can delete phrases later.

If your mother tongue is English and you're studying Spanish, a message would look like this:

Hola
Hello

Give it a try:`
	messageErr              = "Sorry, something went wrong."
	messageExplanationEmpty = "The phrase is missing an explanation. Please send it again with explanation."
	messagePhraseEmpty      = "Please send a phrase."
	messageStudyDone        = `Congrats, you finished all your studies for now!
Come back in %s.`
	messageStudyCorrect = "Correct!"
	messageStudyWrong   = `Sorry, the right version is:

%s`
	messageStudyEmpty = `You have added no phrases yet.
Click the button below and get started.`
	messageStudyQuestion = `%d. Do you remember how to say this?

%s

Use the buttons or type the phrase.`
	messageExplanationExists = `You already saved a phrase with the same explanation:
%s
%s

Please send it again with an explanation you can distinguish from the existing one.`
	messageAddDone = `Saved phrase:
%s

With explanation:
%s`
	messageAddNext        = "Add next phrase."
	messageStudiesDue     = `Hey %s, you have %d phrases ready for review!`
	messageConfirmDelete  = "Are you sure, you want to delete this phrase?"
	messageDeleted        = "The phrase has been deleted. Let's continue studying other phrases."
	messageCancelDelete   = "Good, let's keep that phrase and continue studying."
	messageAskToSubscribe = `

Would you like me to send you a message when there are phrases ready for studying?`
	messageSubscribed = `Good, I will send you a message when your phrases are ready.

` + messageStartMenu
	messageUnsubscribed = `Sure, you won't receive any more notifications.

` + messageStartMenu
	messageNoSubscription = `Sure, you won't receive any notifications.

` + messageStartMenu
	messageFedback      = "If you run into a problem, have any feedback for the people behind Studybot or just like to say hello, you can send a message now and we will get back to you as soon as possible."
	messageFeedbackDone = "Thanks, you will hear from us soon."
	greeting            = `Studybot helps you with our language studies.
Master the language you encounter in your every day life instead of being limited to a textbook.`
)
