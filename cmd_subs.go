package main

import (
	"fmt"
	"steam-discount/storage"
	t "steam-discount/types"
	"strings"

	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

func subs(b *tb.Bot, m *tb.Message) {
	log.WithField("chat_id", m.Chat.ID).Trace("/subs")
	entry, err := strg.Load(t.ChatId(m.Chat.ID))
	if err == storage.Nil {
		b.Send(m.Sender, "You have not any subscriptions yet")
		return
	} else if err != nil {
		log.WithError(err).Error("getting entry failed")
		b.Send(m.Sender, err.Error())
		return
	}
	log.WithField("entry", entry).Trace("Got entry from a database")
	if len(entry.Subscriptions) == 0 {
		b.Send(m.Sender, "You have not any subscriptions yet")
		return
	}
	builder := strings.Builder{}
	for i, sub := range entry.Subscriptions {
		fmt.Fprintf(&builder, "%d) %d", i+1, sub)
	}
	b.Send(m.Sender, builder.String())
}
