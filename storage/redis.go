package storage

import (
	"context"
	"fmt"
	t "steam-discount/types"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
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
	logrus.WithField("redis", r).WithField("key", key).Info("loading by key")
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
func (r *redisStorage) Delete(key t.ChatId) error {
	return r.rdb.Del(r.ctx, key.String()).Err()
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
			logrus.WithError(err).Panic("failed to parse chat_id")
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
		logrus.WithError(err).Panic("couldn't parse redis url")
	}
	return &redisStorage{rdb: redis.NewClient(opts), ctx: context.Background()}
}
