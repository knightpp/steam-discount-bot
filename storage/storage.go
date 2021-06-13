package storage

import (
	t "steam-discount/types"
)

var Nil = nilStruct{}

type nilStruct struct{}

func (s nilStruct) Error() string {
	return "no such item"
}

type Storager interface {
	Delete(key t.ChatId, game t.GameId) error
	Store(key t.ChatId, value t.Entry) error
	Load(key t.ChatId) (t.Entry, error)
	Iterator() Iterator
}
type Iterator interface {
	Next() (t.Entry, error)
}
