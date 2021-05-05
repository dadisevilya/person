package gettStorages

import (
	"github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	"strconv"
	"sync"
	"time"

	"github.com/gtforge/global_services_common_go/gett-config"
	"github.com/gtforge/global_services_common_go/gett-ops"
	tracingHelper "github.com/gtforge/global_services_common_go/gett-ops/opentracing"
	"github.com/newrelic/go-agent"
	"github.com/gtforge/gls"
	"gopkg.in/redis.v5"
)

const (
	GET              = "GET"
	SETEX            = "SETEX"
	SET              = "SET"
	MGET             = "MGET"
	HGETALL          = "HGETALL"
	HMGET            = "HMGET"
	HDEL             = "HDEL"
	HSET             = "HSET"
	HGET             = "HGET"
	DEL              = "DEL"
	EXISTS           = "EXISTS"
	INCR             = "INCR"
	INCRBY           = "INCRBY"
	APPEND           = "APPEND"
	LPUSH            = "LPUSH"
	EXPIRE           = "EXPIRE"
	MSET             = "MSET"
	HMSET            = "HMSET"
	SADD             = "SADD"
	SMEMBERS         = "SMEMBERS"
	LRANGE           = "LRANGE"
	SETNX            = "SETNX"
	ZADD             = "ZADD"
	ZCARD            = "ZCARD"
	ZRANGEBYSCORE    = "ZRANGEBYSCORE"
	ZREMRANGEBYSCORE = "ZREMRANGEBYSCORE"
	TXPIPELINE       = "TXPIPELINE"
	PIPELINE         = "PIPELINE"
	GEOADD           = "GEOADD"
	GEODIST          = "GEODIST"
	GEOHASH          = "GEOHASH"
	GEOPOS           = "GEOPOS"
	GEORADIUS        = "GEORADIUS"
)

type GtRedisClient struct {
	*redis.Client

	// for opentracing tags
	addr string
	db   int
}

var RedisClient *GtRedisClient
var once sync.Once

func newGtRedisClient(client *redis.Client, addr string, db int) *GtRedisClient {
	return &GtRedisClient{
		Client: client,
		addr:   addr,
		db:     db,
	}
}

func InitRedis(config *gettConfig.Config, appEnv string) *GtRedisClient {
	once.Do(func() {
		options := &redis.Options{
			Addr:         config.GetString(appEnv+".main.server") + ":6379",
			Password:     "", // no password set
			MaxRetries:   3,
			DB:           config.GetInt(appEnv + ".main.database"), // use default DB
			PoolSize:     config.GetInt(appEnv + ".main.pool"),
			PoolTimeout:  getDurationSettingIfSet(appEnv, config, "pool_timeout", time.Second),
			ReadTimeout:  getDurationSettingIfSet(appEnv, config, "read_timeout", time.Second),
			WriteTimeout: getDurationSettingIfSet(appEnv, config, "write_timeout", time.Second),
			DialTimeout:  getDurationSettingIfSet(appEnv, config, "dial_timeout", 2*time.Second),
		}
		client := redis.NewClient(options)

		RedisClient = newGtRedisClient(client, options.Addr, options.DB)
	})
	return RedisClient
}

func getDurationSettingIfSet(appEnv string, config *gettConfig.Config, name string, defaultValue time.Duration) time.Duration {
	key := appEnv + ".main." + name
	if config.IsSet(key) {
		val := config.GetInt(key)
		return time.Duration(val) * time.Second
	}
	return defaultValue
}

func (c GtRedisClient) trackSegment(operation string) newrelic.DatastoreSegment {
	if content := gls.Get(gettOps.GlsNewRelicTxnKey); content != nil {
		txn := content.(newrelic.Transaction)
		return newrelic.DatastoreSegment{
			StartTime:  newrelic.StartSegmentNow(txn),
			Product:    newrelic.DatastoreRedis,
			Collection: "",
			Operation:  operation,
		}
	}
	return newrelic.DatastoreSegment{}
}

func (c GtRedisClient) TxPipelined(fn func(*redis.Pipeline) error) ([]redis.Cmder, error) {
	end := c.track(TXPIPELINE)
	res, err := c.Client.TxPipelined(fn)
	end(err)
	return res, err
}

func (c GtRedisClient) Pipelined(fn func(*redis.Pipeline) error) ([]redis.Cmder, error) {
	end := c.track(PIPELINE)
	res, err := c.Client.Pipelined(fn)
	end(err)
	return res, err
}

func (c GtRedisClient) Get(key string) *redis.StringCmd {
	end := c.track(GET)
	res := c.Client.Get(key)
	end(res.Err())
	return res
}

func (c GtRedisClient) MGet(key ...string) *redis.SliceCmd {
	end := c.track(MGET)
	res := c.Client.MGet(key...)
	end(res.Err())
	return res
}

func (c GtRedisClient) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	operation := SET
	if expiration.Seconds() > 0 {
		operation = SETEX
	}
	end := c.track(operation)
	res := c.Client.Set(key, value, expiration)
	end(res.Err())
	return res
}

func (c GtRedisClient) HGet(key, field string) *redis.StringCmd {
	end := c.track(HGET)
	res := c.Client.HGet(key, field)
	end(res.Err())
	return res
}

func (c GtRedisClient) HMGet(key string, fields ...string) *redis.SliceCmd {
	end := c.track(HMGET)
	res := c.Client.HMGet(key, fields...)
	end(res.Err())
	return res
}

func (c GtRedisClient) HGetAll(key string) *redis.StringStringMapCmd {
	end := c.track(HGETALL)
	res := c.Client.HGetAll(key)
	end(res.Err())
	return res
}

func (c GtRedisClient) Del(key ...string) *redis.IntCmd {
	end := c.track(DEL)
	res := c.Client.Del(key...)
	end(res.Err())
	return res
}

func (c GtRedisClient) HDel(key string, fields ...string) *redis.IntCmd {
	end := c.track(HDEL)
	res := c.Client.HDel(key, fields...)
	end(res.Err())
	return res
}

func (c GtRedisClient) HSet(key, field, value string) *redis.BoolCmd {
	end := c.track(HSET)
	res := c.Client.HSet(key, field, value)
	end(res.Err())
	return res
}

func (c GtRedisClient) Exists(key string) *redis.BoolCmd {
	end := c.track(EXISTS)
	res := c.Client.Exists(key)
	end(res.Err())
	return res
}

func (c GtRedisClient) Incr(key string) *redis.IntCmd {
	end := c.track(INCR)
	res := c.Client.Incr(key)
	end(res.Err())
	return res
}

func (c GtRedisClient) IncrBy(key string, val int64) *redis.IntCmd {
	end := c.track(INCRBY)
	res := c.Client.IncrBy(key, val)
	end(res.Err())
	return res
}

func (c GtRedisClient) Append(key, field string) *redis.IntCmd {
	end := c.track(APPEND)
	res := c.Client.Append(key, field)
	end(res.Err())
	return res
}

func (c GtRedisClient) LPush(key string, values ...interface{}) *redis.IntCmd {
	end := c.track(LPUSH)
	res := c.Client.LPush(key, values...)
	end(res.Err())
	return res
}

func (c GtRedisClient) Expire(key string, expiration time.Duration) *redis.BoolCmd {
	end := c.track(EXPIRE)
	res := c.Client.Expire(key, expiration)
	end(res.Err())
	return res
}

func (c GtRedisClient) MSet(pairs ...interface{}) *redis.StatusCmd {
	end := c.track(MSET)
	res := c.Client.MSet(pairs...)
	end(res.Err())
	return res
}

func (c GtRedisClient) HMSet(key string, fields map[string]string) *redis.StatusCmd {
	end := c.track(HMSET)
	res := c.Client.HMSet(key, fields)
	end(res.Err())
	return res
}

func (c GtRedisClient) SAdd(key string, members ...interface{}) *redis.IntCmd {
	end := c.track(SADD)
	res := c.Client.SAdd(key, members...)
	end(res.Err())
	return res
}

func (c GtRedisClient) SMembers(key string) *redis.StringSliceCmd {
	end := c.track(SMEMBERS)
	res := c.Client.SMembers(key)
	end(res.Err())
	return res
}

func (c GtRedisClient) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	end := c.track(LRANGE)
	res := c.Client.LRange(key, start, stop)
	end(res.Err())
	return res
}

func (c GtRedisClient) SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	end := c.track(SETNX)
	res := c.Client.SetNX(key, value, expiration)
	end(res.Err())
	return res
}

func (c *GtRedisClient) ZAdd(key string, members ...redis.Z) *redis.IntCmd {
	end := c.track(ZADD)
	res := c.Client.ZAdd(key, members...)
	end(res.Err())
	return res
}

func (c *GtRedisClient) ZCard(key string) *redis.IntCmd {
	end := c.track(ZCARD)
	res := c.Client.ZCard(key)
	end(res.Err())
	return res
}

func (c *GtRedisClient) ZRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	end := c.track(ZRANGEBYSCORE)
	res := c.Client.ZRangeByScore(key, opt)
	end(res.Err())
	return res
}

func (c *GtRedisClient) ZRemRangeByScore(key, min, max string) *redis.IntCmd {
	end := c.track(ZREMRANGEBYSCORE)
	res := c.Client.ZRemRangeByScore(key, min, max)
	end(res.Err())
	return res
}

func (c *GtRedisClient) GeoAdd(key string, geoLocations []*redis.GeoLocation) *redis.IntCmd {
	end := c.track(GEOADD)
	res := c.Client.GeoAdd(key, geoLocations...)
	end(res.Err())
	return res
}

func (c *GtRedisClient) GeoDist(key, member1, member2, unit string) *redis.FloatCmd {
	end := c.track(GEODIST)
	res := c.Client.GeoDist(key, member1, member2, unit)
	end(res.Err())
	return res
}

func (c *GtRedisClient) GeoHash(key string, members ...string) *redis.StringSliceCmd {
	end := c.track(GEOHASH)
	res := c.Client.GeoHash(key, members...)
	end(res.Err())
	return res
}

func (c *GtRedisClient) GeoPos(key string, members ...string) *redis.GeoPosCmd {
	end := c.track(GEOPOS)
	res := c.Client.GeoPos(key, members...)
	end(res.Err())
	return res
}

func (c *GtRedisClient) GeoRadius(key string, lon float64, lat float64, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	end := c.track(GEORADIUS)
	res := c.Client.GeoRadius(key, lon, lat, query)
	end(res.Err())
	return res
}

type endFunc func(error)

func (c GtRedisClient) track(operation string) endFunc {
	segment := c.trackSegment(operation)
	span := startSpan(operation, c.addr, c.db)

	return func(err error) {
		if err != nil {
			otext.Error.Set(span, true)
			span.SetTag("error.details", err)
		}
		span.Finish()
		segment.End()
	}
}

const redisSpanPrefix = "redis.query"

func startSpan(method string, addr string, db int) opentracing.Span {
	parentSpan := tracingHelper.SpanFromGLS()
	span := opentracing.StartSpan(redisSpanPrefix+" "+method, opentracing.ChildOf(parentSpan.Context()))
	otext.DBType.Set(span, "redis")
	otext.PeerAddress.Set(span, addr)
	otext.DBInstance.Set(span, strconv.Itoa(db))
	otext.SpanKindRPCClient.Set(span)
	return span
}

type RedisMutex struct{}

const mutexTTL = 10 * time.Second

func (m RedisMutex) Lock(key string) (bool, error) {
	response, err := RedisClient.SetNX(key, "value", mutexTTL).Result()
	if err != nil {
		return false, err
	}
	return response, nil
}

func (m RedisMutex) Unlock(key string) error {
	_, err := RedisClient.Del(key).Result()
	return err
}

type TxPipeline struct {
	*redis.Pipeline
	client GtRedisClient
}

func (t *TxPipeline) Exec() ([]redis.Cmder, error) {
	end := t.client.track(TXPIPELINE)
	res, err := t.Pipeline.Exec()
	end(err)
	return res, err
}
func (c GtRedisClient) TxPipeline() TxPipeline {
	return TxPipeline{client: c, Pipeline: c.Client.TxPipeline()}
}
