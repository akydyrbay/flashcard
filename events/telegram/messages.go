package telegram

const (
	msgUnknownCommand = "Sorry, I didn't understand that. Type /help for usage."
	msgHelp           = `Usage:
/save - save a flashcards under a name
/get - retrieve saved text
/list - list all saved names
/help - for usage
/delete - to delete saved cards
/next - to show answer
`
	msgHello           = "Welcome! Use /help to see commands."
	msgAlreadyExists   = "An entry with that name already exists."
	msgSaved           = "Saved!"
	msgNoSavedItems    = "No entries found by that name."
	msgUsageGet        = "Usage: /get"
	msgInvalidFormat   = "Invalid format!"
	msgSaveCmdResponse = "Great! Please send your Q&A in this format:\n" +
		"q:<question1> \n a:<answer1> \n q:<question2> \n a:<answer2> ..."
	msgNoSuchItem     = "I couldn't find a flashcards by that name."
	msgDeleteResponse = "Sure! Please send me the name of the flashcard you want to delete."
	msgGetCmdResponse = "Please send a name of the flashcards to get."
	msgQuizComplete   = "Quiz is finished"
	msgSaveName       = "Got your Q&A. Now please send the **name** you want to save this under."
	msgNoActive       = "No active quiz—send /get first."
)
