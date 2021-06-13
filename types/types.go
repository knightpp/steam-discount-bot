package types

import "fmt"

type ChatId int64

func (ci ChatId) Recipient() string {
	return ci.String()
}

func (ci ChatId) String() string {
	return fmt.Sprintf("%d", ci)
}

type GameId uint64

func (gi GameId) String() string {
	return fmt.Sprintf("%d", gi)
}

type Entry struct {
	Id            ChatId
	Subscriptions []GameId
}

func NewEntry() Entry {
	return Entry{
		Subscriptions: make([]GameId, 0),
	}
}
