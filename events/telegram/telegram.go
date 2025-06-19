package telegram

import (
	"errors"
	"flashcard/clients/telegram"
	"flashcard/events"
	"flashcard/lib/e"
	"flashcard/storage"
)

type Processor struct {
	tg              *telegram.Client
	offset          int
	storage         storage.Storage
	pendingDelete   map[int]bool
	pendingGet      map[int]bool
	pendingSaveQA   map[int]*saveState // waiting for the Q&A
	pendingSaveName map[int]*saveState // waiting for the final name
	sessions        map[int]*session   // chatID → current session
}
type saveState struct {
	rawQA string // the Q&A text the user sent
}

// a single Q&A pair
type qaPair struct{ Q, A string }

// holds an in‐progress flashcard session
type session struct {
	pairs []qaPair // all Q&A
	idx   int      // next index to reveal
}

type Meta struct {
	ChatID   int
	UserName string
}

var ErrUnknownEventType = errors.New("unknown event type")
var ErrUnknownMetaType = errors.New("unknown meta type")

func New(client *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{tg: client,
		storage:         storage,
		pendingDelete:   make(map[int]bool),
		pendingGet:      make(map[int]bool),
		pendingSaveQA:   make(map[int]*saveState),
		pendingSaveName: make(map[int]*saveState),
		sessions:        make(map[int]*session),
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}
	res := make([]events.Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(u))
	}

	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}
func (p *Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	default:
		return e.Wrap("can't process message", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	if err := p.doCmd(event.Text, meta.ChatID, meta.UserName); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get Meta", ErrUnknownMetaType)
	}

	return res, nil
}
func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)
	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			UserName: upd.Message.From.UserName,
		}
	}
	return res
}

func fetchType(upd telegram.Update) events.Type {
	if upd.Message == nil {
		return events.Unknown
	}
	return events.Message
}

func fetchText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""
	}
	return upd.Message.Text
}
