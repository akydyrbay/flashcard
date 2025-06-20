package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"flashcard/lib/e"
	"flashcard/storage"
)

const (
	SaveCmd   = "/save"
	GetCmd    = "/get"
	HelpCmd   = "/help"
	StartCmd  = "/start"
	ListCmd   = "/list"
	DeleteCmd = "/delete"
	NextCmd   = "/next"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)
	log.Printf("got new command '%s' from '%s'", text, username)

	// 1) If chat is waiting for a name → treat text as the name
	if st, ok := p.pendingSaveName[chatID]; ok {
		delete(p.pendingSaveName, chatID)
		return p.finishSave(chatID, username, st.rawQA, text)
	}

	// 2) If chat is waiting for the QA → treat text as the raw QA
	if _, ok := p.pendingSaveQA[chatID]; ok {
		// store the QA, move to next step
		p.pendingSaveName[chatID] = &saveState{rawQA: text}
		delete(p.pendingSaveQA, chatID)
		return p.tg.SendMessage(chatID, msgSaveName)
	}
	if p.pendingDelete[chatID] {
		delete(p.pendingDelete, chatID)
		return p.handleDeleteContent(chatID, username, text)
	}
	if p.pendingGet[chatID] {
		delete(p.pendingGet, chatID)
		return p.handleGet(chatID, username, text)
	}

	// parts := strings.SplitN(text, " ", 3)
	switch text {
	case DeleteCmd:
		p.pendingDelete[chatID] = true
		return p.tg.SendMessage(chatID, msgDeleteResponse)

	case SaveCmd:
		p.pendingSaveQA[chatID] = &saveState{}
		return p.tg.SendMessage(chatID, msgSaveCmdResponse)

	case GetCmd:
		p.pendingGet[chatID] = true
		return p.tg.SendMessage(chatID, msgGetCmdResponse)
	case NextCmd:
		return p.advanceSession(chatID)

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

func (p *Processor) finishSave(chatID int, user, rawQA, name string) error {
	// rawQA is the Q&A string, name is the final name
	// 1) validate and parse rawQA

	if rawQA == "" {
		return p.tg.SendMessage(chatID, msgInvalidFormat)
	}

	// 2) now call your existing saveItem logic:
	return p.saveItem(chatID, user, name, rawQA)
}

func (p *Processor) handleGet(chatID int, user, name string) error {
	item := &storage.Item{UserName: user, Name: name}
	exists, err := p.storage.IsExists(context.Background(), item)
	if err != nil {
		return e.Wrap("get item", err)
	}
	if !exists {
		return p.tg.SendMessage(chatID, msgNoSavedItems) // or a new msgNoSuchItem
	}
	return p.startSession(chatID, user, name)
}

func (p *Processor) handleDeleteContent(chatID int, user, name string) error {
	// 1) Check existence
	item := &storage.Item{UserName: user, Name: name}
	exists, err := p.storage.IsExists(context.Background(), item)
	if err != nil {
		return e.Wrap("delete item", err)
	}
	if !exists {
		return p.tg.SendMessage(chatID, msgNoSavedItems) // or a new msgNoSuchItem
	}

	// 2) Delete
	if err := p.storage.Remove(context.Background(), item); err != nil {
		return e.Wrap("delete item", err)
	}

	// 3) Confirm
	return p.tg.SendMessage(chatID, fmt.Sprintf("Deleted deck “%s”.", name))
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

func (p *Processor) startSession(chatID int, user, name string) (err error) {
	defer func() { err = e.WrapIfErr("start session", err) }()

	item, err := p.storage.Get(context.Background(), user, name)
	if err != nil {
		if errors.Is(err, storage.ErrNoSavedItems) {
			return p.tg.SendMessage(chatID, msgNoSavedItems)
		}
		return err
	}

	// parse all Q&A pairs
	qaMap := extractQA(item.Content)
	if len(qaMap) == 0 {
		return p.tg.SendMessage(chatID, msgInvalidFormat)
	}

	// build slice in deterministic order
	pairs := make([]qaPair, 0, len(qaMap))
	for q, a := range qaMap {
		pairs = append(pairs, qaPair{Q: q, A: a})
	}

	// save session: start at idx=0
	p.sessions[chatID] = &session{pairs: pairs, idx: 0}

	// send first question
	return p.tg.SendMessage(chatID, pairs[0].Q)
}

func (p *Processor) advanceSession(chatID int) error {
	sess, ok := p.sessions[chatID]
	if !ok {
		return p.tg.SendMessage(chatID, msgNoActive)
	}

	// send the answer to the previous question
	prev := sess.idx
	if prev < len(sess.pairs) {
		if err := p.tg.SendMessage(chatID, sess.pairs[prev].A); err != nil {
			return err
		}
	}

	// move to next
	sess.idx++
	if sess.idx >= len(sess.pairs) {
		delete(p.sessions, chatID)
		return p.tg.SendMessage(chatID, msgQuizComplete)
	}

	// send next question
	return p.tg.SendMessage(chatID, sess.pairs[sess.idx].Q)
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
		if strings.HasPrefix(line, "q:") || strings.HasPrefix(line, "Q:") {
			currentQ = strings.TrimSpace(strings.TrimPrefix(line, "q:"))
		} else if strings.HasPrefix(line, "a:") && currentQ != "" || strings.HasPrefix(line, "A:") && currentQ != "" {
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
