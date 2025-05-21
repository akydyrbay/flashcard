package telegram

import (
	"log"
	"strings"
)

func (p *Processor) doCMD(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command '%s' from '%s'", text, username)
	//add page
	//rnd
	//help
	//start
	return nil
}
