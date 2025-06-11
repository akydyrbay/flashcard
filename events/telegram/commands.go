package telegram

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"flashcard/lib/e"
	"flashcard/storage"
)

const (
	SaveCmd  = "/save" // /save <name> <text>
	GetCmd   = "/get"  // /get <name>
	HelpCmd  = "/help"
	StartCmd = "/start"
	ListCmd  = "/list"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)
	log.Printf("got new command '%s' from '%s'", text, username)

	parts := strings.SplitN(text, " ", 3)
	switch parts[0] {
	case SaveCmd:
		if len(parts) < 3 {
			return p.tg.SendMessage(chatID, msgUsageSave)
		}
		return p.saveItem(chatID, username, parts[1], parts[2])

	case GetCmd:
		if len(parts) < 2 {
			return p.tg.SendMessage(chatID, msgUsageGet)
		}
		return p.getItem(chatID, username, parts[1])
	case ListCmd:
		return p.listItems(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	default:
		return p.tg.SendMessage(chatID, msgUnknownCommand)
	}
}

func (p *Processor) saveItem(chatID int, user, name, content string) (err error) {
	defer func() { err = e.WrapIfErr("save item", err) }()

	// 1) Verify flashcard format
	qaMap := extractQA(content)
	if len(qaMap) == 0 {
		return p.tg.SendMessage(chatID, msgInvalidFormat)
	}

	// 2) Prepare item
	item := &storage.Item{
		Name:     name,
		Content:  content,
		UserName: user,
	}

	// 3) Check for duplicates
	exists, err := p.storage.IsExists(context.Background(), item)
	if err != nil {
		return err
	}
	if exists {
		return p.tg.SendMessage(chatID, msgAlreadyExists)
	}

	// 4) Save to storage
	if err := p.storage.Save(context.Background(), item); err != nil {
		return err
	}

	// 5) Acknowledge
	return p.tg.SendMessage(chatID, msgSaved)
}

func (p *Processor) getItem(chatID int, user, name string) (err error) {
	defer func() { err = e.WrapIfErr("get item", err) }()
	item, err := p.storage.Get(context.Background(), user, name)
	if err != nil {
		if errors.Is(err, storage.ErrNoSavedItems) {
			return p.tg.SendMessage(chatID, msgNoSavedItems)
		}
		return err
	}
	contentMap := extractQA(item.Content)
	for q, a := range contentMap {
		// send question
		if err = p.tg.SendMessage(chatID, q); err != nil {
			return err
		}

		// pause for a second
		time.Sleep(5 * time.Second)

		// send answer
		if err = p.tg.SendMessage(chatID, a); err != nil {
			return err
		}
	}
	return nil
}

func extractQA(text string) map[string]string {
	result := make(map[string]string)

	// 1) Drop the first line if it’s a "/save" command
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "/save") {
		lines = lines[1:]
	}

	// 2) Walk the rest of the lines, pairing q: → a:
	var currentQ string
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "q:") {
			currentQ = strings.TrimSpace(strings.TrimPrefix(line, "q:"))
		} else if strings.HasPrefix(line, "a:") && currentQ != "" {
			answer := strings.TrimSpace(strings.TrimPrefix(line, "a:"))
			result[currentQ] = answer
			currentQ = "" // reset until next question
		}
	}

	return result
}

func (p *Processor) listItems(chatID int, user string) (err error) {
	defer func() { err = e.WrapIfErr("list items", err) }()
	names, err := p.storage.List(context.Background(), user)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return p.tg.SendMessage(chatID, msgNoSavedItems)
	}
	return p.tg.SendMessage(chatID, "Your items:\n"+strings.Join(names, "\n"))
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}
