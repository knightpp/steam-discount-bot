package main

import (
	"fmt"
	"regexp"
	"steam-discount/storage"
	t "steam-discount/types"
	"strconv"

	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

const MAX_ALLOWED_SUBSCRIPTIONS = 10

var gameUrlRegex = regexp.MustCompile(`store\.steampowered\.com/app/(\d{1,9})`)

func sub(b *tb.Bot, m *tb.Message) {
	var gameId t.GameId
	{
		var toParse string

		matches := gameUrlRegex.FindStringSubmatch(m.Payload)
		if matches == nil {
			toParse = m.Payload
		} else {
			toParse = matches[1]
		}

		parsed, err := strconv.ParseUint(toParse, 10, 0)
		if err != nil {
			msg := fmt.Sprintf("incorrect syntax, example: /sub <620|https://store.steampowered.com/app/620/Portal_2/>\nand you wrote: %s", m.Payload)
			b.Send(m.Sender, msg)
			return
		}
		gameId = t.GameId(parsed)
	}

	chatId := t.ChatId(m.Chat.ID)
	log.WithField("game_id", gameId).Trace("Subscribe")
	entry, err := strg.Load(chatId)
	if err == storage.Nil {
		log.Trace("Not in database")
		// user haven't added anything yet
		entry.Id = t.ChatId(m.Chat.ID)
		entry.Subscriptions = []t.GameId{t.GameId(gameId)}
	} else if err != nil {
		b.Send(m.Sender, fmt.Sprintf("loading failed: %s", err))
		return
	} else {
		AddDedup(&entry.Subscriptions, t.GameId(gameId))
	}
	log.WithField("entry", entry).Trace("Got entry from a database")
	if len(entry.Subscriptions) > MAX_ALLOWED_SUBSCRIPTIONS {
		b.Send(m.Sender, "You have reached the limit of subscriptions."+
			"The maximum allowed number of subscriptions is 10.")
		return
	}
	err = strg.Store(chatId, entry)
	if err != nil {
		b.Send(m.Sender, fmt.Sprintf("storing failed: %s", err))
		return
	}
	b.Send(m.Sender, fmt.Sprintf("Subscribed to %s", gameId))
}

func AddDedup(subs *[]t.GameId, gameId t.GameId) {
	var found bool
	for _, gi := range *subs {
		if gi == gameId {
			found = true
			break
		}
	}
	if !found {
		*subs = append(*subs, gameId)
	}
}
