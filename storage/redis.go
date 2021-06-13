package storage

import (
	"context"
	"fmt"
	t "steam-discount/types"
	"strconv"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"
)

type redisStorage struct {
	ctx context.Context
	rdb *redis.Client
}

func (r *redisStorage) Store(key t.ChatId, value t.Entry) error {
	bytes, err := msgpack.Marshal(value)
	if err != nil {
		return fmt.Errorf("Store(): msgpack marshal failed: %w", err)
	}

	err = r.rdb.Set(r.ctx, key.String(), bytes, 0).Err()
	if err != nil {
		return fmt.Errorf("Store(): redis returned error: %w", err)
	}
	return nil
}

func (r *redisStorage) Load(key t.ChatId) (t.Entry, error) {
	log.WithField("redis", r).WithField("key", key).Info("loading by key")
	var entry t.Entry
	bytes, err := r.rdb.Get(r.ctx, key.String()).Bytes()
	if err == redis.Nil {
		return entry, nilStruct{}
	}
	if err != nil {
		return entry, fmt.Errorf("Load(): redis error: %w", err)
	}
	err = msgpack.Unmarshal(bytes, &entry)
	if err != nil {
		return entry, fmt.Errorf("Load(): msgpack unmarshal failed: %w", err)
	}
	return entry, nil
}
func (r *redisStorage) Delete(key t.ChatId, game t.GameId) error {
	entry, err := r.Load(key)
	if err != nil {
		return err
	}
	log.WithField("entry", entry).Trace("loading entry")
	indexToDelete := -1
	for i, g := range entry.Subscriptions {
		if g == game {
			indexToDelete = i
		}
	}
	log.WithField("indexToDelete", indexToDelete).Trace("about to delete")
	if indexToDelete == -1 {
		// nothing to delete, just return
		return nil
	}
	a := entry.Subscriptions
	copy(a[indexToDelete:], a[indexToDelete+1:]) // Shift a[i+1:] left one index.
	entry.Subscriptions = a[:len(a)-1]           // Truncate slice.
	if len(entry.Subscriptions) == 0 {
		log.WithField("chat_id", key).Trace("deleting entry")
		// if there are no subs just delete all entry
		return r.rdb.Del(r.ctx, key.String()).Err()
	} else {
		log.WithFields(
			log.Fields{
				"chat_id": key,
				"entry":   entry,
			}).Trace("storing modified entry")
		// otherwise store back entry
		return r.Store(key, entry)
	}
}

func (r *redisStorage) Iterator() Iterator {
	ctx := context.Background()
	return redisIterator{
		r:    r,
		iter: r.rdb.Scan(ctx, 0, "*", 0).Iterator(),
	}
}

type redisIterator struct {
	r    *redisStorage
	iter *redis.ScanIterator
}

func (it redisIterator) Next() (t.Entry, error) {
	if it.iter.Next(it.r.ctx) {
		chatId, err := strconv.ParseInt(it.iter.Val(), 10, 0)
		if err != nil {
			log.WithError(err).Panic("failed to parse chat_id")
			return t.Entry{}, err
		}
		entry, err := it.r.Load(t.ChatId(chatId))
		if err != nil {
			return entry, err
		}
		return entry, nil
	} else {
		return t.Entry{}, Nil
	}
}

func NewRedis(url string) Storager {
	opts, err := redis.ParseURL(url)
	if err != nil {
		log.WithError(err).Panic("couldn't parse redis url")
	}
	return &redisStorage{rdb: redis.NewClient(opts), ctx: context.Background()}
}
