package gettStorages

import (
	"fmt"
	"time"

	"gopkg.in/redis.v5"
)

type LuaScript struct {
	*redis.Script
	name string
}

func NewLuaScript(script, name string) *LuaScript {
	return &LuaScript{
		Script: redis.NewScript(script),
		name:   name,
	}
}

func (c *GtRedisClient) RunLua(script *LuaScript, keys []string, args ...interface{}) (interface{}, error) {
	if script.name != "" {
		segment := c.trackSegment(script.name)
		defer segment.End()
	}
	return script.Run(c, keys, args...).Result()
}

const compareAndSetString = `local value = redis.call('GET', KEYS[1])
if (not value or value == ARGV[1]) then
	return redis.call('SETEX', KEYS[2], ARGV[3], ARGV[2]) and 1 or 0
end
return 0`

var compareAndSetScript = NewLuaScript(compareAndSetString, "CompareXAndSetY")

func (c *GtRedisClient) CompareXAndSetY(xKey string, xValue string, yKey string, yValueToSet string, duration time.Duration) (bool, error) {
	var (
		res     interface{}
		err     error
		seconds string
	)

	seconds = fmt.Sprintf("%d", duration/time.Second)
	res, err = c.RunLua(compareAndSetScript, []string{xKey, yKey}, xValue, yValueToSet, seconds)
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, err
}

const compareAndSwapString = "return redis.call('GET', KEYS[1]) == ARGV[1] and redis.call('SETEX', KEYS[1], ARGV[3], ARGV[2]) and 1 or 0;"

var compareAndSwapScript = NewLuaScript(compareAndSwapString, "CompareAndSwap")

func (c *GtRedisClient) CompareAndSwap(key, old, new string, dur time.Duration) (bool, error) {
	var (
		res interface{}
		err error
	)

	seconds := fmt.Sprintf("%d", dur/time.Second)
	res, err = c.RunLua(compareAndSwapScript, []string{key}, old, new, seconds)
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, err
}

func (c *GtRedisClient) DeleteIfValueMatch(key, oldValue string) (bool, error) {
	isDeleted, err := c.RunLua(deleteIfMatchScript, []string{key}, oldValue)
	if err != nil {
		return false, err
	}

	return isDeleted.(int64) == 1, nil
}

const deleteIfValueMatch = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
	else
	return 0
	end
`

var deleteIfMatchScript = NewLuaScript(deleteIfValueMatch, "deleteIfMatch")
