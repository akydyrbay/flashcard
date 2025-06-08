package telegram

const (
	msgUnknownCommand = "Sorry, I didn't understand that. Type /help for usage."
	msgHelp           = `I can save and retrieve named texts:
/save <name> <text> - save a text under a name
/get <name>        - retrieve saved text
/list - list all saved names
`
	msgHello         = "Welcome! Use /help to see commands."
	msgAlreadyExists = "An entry with that name already exists."
	msgSaved         = "Saved!"
	msgNoSavedItems  = "No entries found by that name."
	msgUsageSave     = "Usage: /save <name> <text>"
	msgUsageGet      = "Usage: /get <name>"
)
