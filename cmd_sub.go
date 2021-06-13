package main

import (
	"fmt"
	"steam-discount/storage"
	t "steam-discount/types"
	"strconv"

	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

func sub(b *tb.Bot, m *tb.Message) {
	gameId, err := strconv.ParseUint(m.Payload, 10, 0)
	chatId := t.ChatId(m.Chat.ID)
	if err != nil {
		b.Send(m.Sender, fmt.Sprintf("couldn't parse game id: %s", err))
		return
	}
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
	err = strg.Store(chatId, entry)
	if err != nil {
		b.Send(m.Sender, fmt.Sprintf("storing failed: %s", err))
		return
	}
	b.Send(m.Sender, fmt.Sprintf("Subscribed to %s", gameId))
}

func AddDedup(subs *[]t.GameId, gi t.GameId) {
	*subs = append(*subs, gi)
}
