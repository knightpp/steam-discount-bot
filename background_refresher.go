package main

import (
	"fmt"
	"steam-discount/storage"
	t "steam-discount/types"

	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

func backgroundRefresher2(b *tb.Bot) error {
	it := strg.Iterator()
	m := make(map[t.GameId][]t.ChatId)
	for {
		entry, err := it.Next()
		if err == storage.Nil {
			break
		} else if err != nil {
			return fmt.Errorf("db iterator returned error: %w", err)
		}
		log.WithField("entry", entry).Debug("iterator returned entry")
		for _, gi := range entry.Subscriptions {
			m[gi] = append(m[gi], entry.Id)
		}
	}
	gameIds := make([]t.GameId, 0, 100)
	for k := range m {
		gameIds = append(gameIds, k)
		if len(gameIds) == 100 {
			resp, err := requestPriceOverview(gameIds)
			if err != nil {
				return fmt.Errorf("reques to steam returned: %w", err)
			}
			process(b, resp, m)

			gameIds = make([]t.GameId, 0, 100)
		}

	}
	if len(gameIds) != 0 {
		resp, err := requestPriceOverview(gameIds)
		if err != nil {
			return fmt.Errorf("request to steam returned: %w", err)
		}
		process(b, resp, m)
	}

	return nil
}

// Given slice of chat ids, sends the same message to all.
func sendToChats(b *tb.Bot, chats []t.ChatId, message string) {
	log.WithField("chats", chats).WithField("message", message).
		Debug("sending message to chats")
	for _, c := range chats {
		b.Send(c, message)
	}
}

func deleteGameFromChats(gameId t.GameId, chats map[t.GameId][]t.ChatId) {
	for _, chat := range chats[gameId] {
		err := strg.Delete(chat, gameId)
		if err != nil {
			log.WithError(err).Warn("failed to delete")
		}
	}
}

// Iterates over responses, on error unsubscribes.
// If DiscountPercent > 0, sends message to all subscribed
// chats and unsubscribes the chats.
func process(b *tb.Bot, steamResponses map[t.GameId]SteamResponse, chats map[t.GameId][]t.ChatId) {
	for gameId, sr := range steamResponses {
		if !sr.Success {
			deleteGameFromChats(gameId, chats)
			sendToChats(b, chats[gameId],
				fmt.Sprintf("Failed to check (%s), you have been unsubscribed from the game.", gameId))
			continue
		}
		if sr.Data.PriceOverview.DiscountPercent > 0 {
			deleteGameFromChats(gameId, chats)
			sendToChats(b, chats[gameId],
				fmt.Sprintf("(%s) is on sale now, discount %d%%\n%s%s", gameId,
					sr.Data.PriceOverview.DiscountPercent,
					"https://store.steampowered.com/app/",
					gameId))
		}
	}
}
