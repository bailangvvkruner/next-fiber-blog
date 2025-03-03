package auth

import (
	"context"
	"fmt"
	"go-fiber-ent-web-layout/internal/data"
	"go-fiber-ent-web-layout/internal/usercase"
	"log/slog"
	"math"
	"strconv"
	"sync"
)

const redisKeyPrefix = "CLASSIC:USER:"

type ClassicLoginUser interface {
	LoginUser
	SetLabels(labels []string) error // SetLabels 更新用户标签
}

type ClassicUserCache interface {
	AddUser(token string, user ClassicLoginUser) error

	RemoveUser(token string) error

	RemoveUserById(userId uint64) error

	GetUser(token string) ClassicLoginUser

	GetUserById(userId uint64) ClassicLoginUser
}

type InRedisClassicUserCache struct {
	redisTemplate *data.RedisTemplate
}

func NewClassicUserCache() ClassicUserCache {
	return &InRedisClassicUserCache{
		redisTemplate: data.DefaultRedisTemplate(),
	}
}

func (self *InRedisClassicUserCache) AddUser(token string, user ClassicLoginUser) error {
	key := redisKeyPrefix + token + ":" + strconv.FormatUint(user.GetUserId(), 10)
	return self.redisTemplate.Set(context.Background(), key, user, math.MaxInt64)
}

func (self *InRedisClassicUserCache) RemoveUser(token string) error {
	key := self.FindUserCacheKey(token, 0)
	if "" == key {
		return fmt.Errorf("获取用户登录信息缓存Key失败")
	}
	err := self.redisTemplate.Delete(context.Background(), key)
	return err
}

func (self *InRedisClassicUserCache) RemoveUserById(userId uint64) error {
	key := self.FindUserCacheKey("", userId)
	if "" == key {
		return fmt.Errorf("获取用户登录信息缓存Key失败")
	}
	err := self.redisTemplate.Delete(context.Background(), key)
	return err
}

func (self *InRedisClassicUserCache) GetUser(token string) ClassicLoginUser {
	key := self.FindUserCacheKey(token, 0)
	if "" == key {
		return nil
	}
	result, err := data.RedisGetStruct[usercase.User](context.Background(), key)
	if err != nil {
		slog.Error("获取博客登录用户失败", "key", key, "err", err.Error())
		return nil
	}
	return &result
}

func (self *InRedisClassicUserCache) GetUserById(userId uint64) ClassicLoginUser {
	key := self.FindUserCacheKey("", userId)
	if "" == key {
		return nil
	}
	result, err := data.RedisGetStruct[usercase.User](context.Background(), key)
	if err != nil {
		slog.Error("获取博客登录用户失败", "key", key, "err", err.Error())
		return nil
	}
	return &result
}

func (self *InRedisClassicUserCache) FindUserCacheKey(token string, userId uint64) string {
	if "" != token && userId > 0 {
		return redisKeyPrefix + token + ":" + strconv.FormatUint(userId, 10)
	}
	var key string
	if userId != 0 {
		key = redisKeyPrefix + "*" + ":" + strconv.FormatUint(userId, 10)
	} else {
		key = redisKeyPrefix + token + ":*"
	}
	result, err := self.redisTemplate.Client().Keys(context.Background(), key).Result()
	if err != nil {
		slog.Error("查询博客登录用户信息缓存Key失败", "err", err.Error())
		return ""
	}
	if len(result) == 0 {
		slog.Error("当前博客登录用户Token不存在，登录已过期")
		return ""
	}
	if len(result) > 1 {
		slog.Error("博客登录用户信息缓存key存在多个", "size", len(result))
		return ""
	}
	return result[0]
}

// 默认博客用户登录管理缓存
var (
	defaultClassicUserCache ClassicUserCache
	classicUserCacheOnce    sync.Once
)

func initDefaultClassicUserCache() {
	defaultClassicUserCache = NewClassicUserCache()
}

func AddClassicLoginUser(token string, user ClassicLoginUser) error {
	classicUserCacheOnce.Do(initDefaultClassicUserCache)
	return defaultClassicUserCache.AddUser(token, user)
}

func RemoveClassicLoginUser(token string) error {
	classicUserCacheOnce.Do(initDefaultClassicUserCache)
	return defaultClassicUserCache.RemoveUser(token)
}

func RemoveClassicLoginUserById(userId uint64) error {
	classicUserCacheOnce.Do(initDefaultClassicUserCache)
	return defaultClassicUserCache.RemoveUserById(userId)
}

func GetClassicLoginUser(token string) ClassicLoginUser {
	classicUserCacheOnce.Do(initDefaultClassicUserCache)
	return defaultClassicUserCache.GetUser(token)
}

func GetClassicLoginUserById(userId uint64) ClassicLoginUser {
	classicUserCacheOnce.Do(initDefaultClassicUserCache)
	return defaultClassicUserCache.GetUserById(userId)
}
