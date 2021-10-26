package dataloader

import (
	"context"
	"math/rand"
	"tgd/models"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nzlov/dataloader"
	"gorm.io/gorm"
)

var _dataloader *Loaders

type loadCache interface {
	Save(string, []byte)
	Get(string) ([]byte, bool)
	SaveExpire(string, time.Duration, []byte)
	GetExpire(string, time.Duration) ([]byte, bool)
	Clear(...string)
}

type Loaders struct {
	rc         *redis.Client
	to         time.Duration
	rcls       *redis.Script
	UserLoader *dataloader.Loader[uint, models.User]
}

const _ctxKey = "nzlov@dataloader"

func NewLoaders(db *gorm.DB, rc *redis.Client, to time.Duration) *Loaders {
	if to < time.Second {
		to = time.Second
	}

	l := &Loaders{
		rc: rc,
		to: to,
	}

	l.rcls = redis.NewScript(`local cursor = 0
local keyNum = 0  
repeat
   local res = redis.call('scan',cursor,'MATCH',KEYS[1],'COUNT',ARGV[1])
   if(res ~= nil and #res>=0) 
   then
      cursor = tonumber(res[1])
      local ks = res[2]
      if(ks ~= nil and #ks>0) 
      then
         for i=1,#ks,1 do
            local key = tostring(ks[i])
            redis.call('UNLINK',key)
         end
         keyNum = keyNum + #ks
      end
     end
until( cursor <= 0 )
return keyNum`)

	l.init(db)
	return l
}

func (l *Loaders) init(db *gorm.DB) {
	lowconfig := dataloader.Config{
		Wait:      time.Millisecond * 100,
		CacheTime: time.Minute * 10,
		MaxBatch:  100,
		Prefix:    "nzlov@Cache:",
	}
	l.UserLoader = dataloader.NewLoader(lowconfig.WithPrefix("User:"), l, userLoader(db))
}

func GetLoader() *Loaders {
	return _dataloader
}

func InitLoader(db *gorm.DB, redis *redis.Client, to time.Duration) *Loaders {
	_dataloader = NewLoaders(db, redis, to)
	return _dataloader
}

func (l *Loaders) Ctx(ctx context.Context) context.Context {
	return context.WithValue(ctx, _ctxKey, l)
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(_ctxKey).(*Loaders)
}

func es(l int, err error) []error {
	if err == nil {
		return nil
	}
	return []error{err}
	//ess := make([]error, l)
	//for i := 0; i < l; i++ {
	//	ess[i] = err
	//}
	//return ess
}

func (l *Loaders) Clear(key ...string) {
	l.rc.Del(context.Background(), key...).Err()
}

func randt() time.Duration {
	return time.Duration(rand.Intn(500))
}

func (l *Loaders) Clean(pattern string) {
	l.rcls.Run(context.Background(), l.rc, []string{pattern}, 1000).Result()
}

func (l *Loaders) SaveExpire(key string, expire time.Duration, value []byte) {
	l.rc.Set(context.Background(), key, value, expire+randt()).Err()
}

func (l *Loaders) GetExpire(key string, expire time.Duration) ([]byte, bool) {
	data, _ := l.rc.GetEx(context.Background(), key, expire+randt()).Bytes()
	return data, true
}

var user = &models.User{}

func userLoader(db *gorm.DB) func(keys []uint) ([]*models.User, []error) {
	return func(keys []uint) ([]*models.User, []error) {
		objs := []models.User{}
		if err := db.Where("id in (?)", keys).Find(&objs).Error; err != nil {
			return nil, es(len(keys), err)
		}

		m := map[uint]*models.User{}
		for i, v := range objs {
			m[v.ID] = &objs[i]
		}
		rs := make([]*models.User, len(keys))
		for i, v := range keys {
			if _, ok := m[v]; ok {
				rs[i] = m[v]
			} else {
				rs[i] = user
			}
		}
		return rs, es(len(keys), nil)
	}
}
