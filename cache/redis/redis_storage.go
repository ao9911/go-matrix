package redis

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"

	"github.com/ao9911/go-matrix/log"
	"github.com/ao9911/go-matrix/util/xtime"
)

type RedisStorage struct {
	client        *redis.Client
	clusterClient *redis.ClusterClient

	cmd    redis.UniversalClient
	config *Config
}

type Config struct {
	Name         string         `json:"name"`
	Addrs        []string       `json:"addrs"`
	Password     string         `json:"password"`
	PoolSize     int            `json:"pool_size"`
	DB           int            `json:"db"`
	MinIdleConns int            `json:"minidle_conns"`
	Dial         xtime.Duration `json:"dial"`
	KeepAlive    xtime.Duration `json:"keep_alive"`
	Cluster      bool           `json:"cluster"`
	Trace        bool           `json:"trace"`
}

func CreateRedisStorage(option *Config) (*RedisStorage, error) {
	if option == nil || len(option.Addrs) == 0 {
		return nil, errors.New("addrs cannot be empty")
	}

	storage := &RedisStorage{
		config: option,
	}

	buildDialer := func(tlsConf *tls.Config) func(ctx context.Context, network, addr string) (net.Conn, error) {
		return func(ctx context.Context, network, addr string) (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   time.Duration(option.Dial),
				KeepAlive: time.Duration(option.KeepAlive),
			}
			if tlsConf == nil {
				return netDialer.DialContext(ctx, network, addr)
			}
			return tls.DialWithDialer(netDialer, network, addr, tlsConf)
		}
	}

	if len(option.Addrs) > 1 || option.Cluster {
		o := &redis.ClusterOptions{
			Addrs:        option.Addrs,
			ReadOnly:     true,
			MinIdleConns: option.MinIdleConns,
		}
		if option.PoolSize > 0 {
			o.PoolSize = option.PoolSize
		}
		o.Dialer = buildDialer(o.TLSConfig)

		client := redis.NewClusterClient(o)
		storage.clusterClient = client
		storage.cmd = client // [优化] 赋值给通用接口
	} else {
		o := &redis.Options{
			Addr:         option.Addrs[0],
			Password:     option.Password,
			DB:           option.DB,
			MinIdleConns: option.MinIdleConns,
		}
		if option.PoolSize > 0 {
			o.PoolSize = option.PoolSize
		}
		o.Dialer = buildDialer(o.TLSConfig)

		client := redis.NewClient(o).WithTimeout(time.Second)
		storage.client = client
		storage.cmd = client // [优化] 赋值给通用接口
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := storage.cmd.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "redis ping check failed")
	}

	if option.Trace {
		if err := redisotel.InstrumentTracing(storage.cmd); err != nil {
			log.Errorf("redisotel.InstrumentTracing error %v", err)
		}
	}

	return storage, nil
}

func (rs *RedisStorage) DB() *redis.Client {
	return rs.client
}

func (rs *RedisStorage) ClusterDB() *redis.ClusterClient {
	return rs.clusterClient
}

func (rs *RedisStorage) ZRevRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return rs.cmd.ZRevRange(ctx, key, start, stop)
}

func (rs *RedisStorage) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return rs.cmd.ZRevRangeWithScores(ctx, key, start, stop)
}

func (rs *RedisStorage) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd {
	return rs.cmd.ZRevRangeByScoreWithScores(ctx, key, opt)
}

func (rs *RedisStorage) ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd {
	return rs.cmd.ZRangeByScoreWithScores(ctx, key, opt)
}

func (rs *RedisStorage) ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	return rs.cmd.ZScore(ctx, key, member)
}

func (rs *RedisStorage) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return rs.cmd.SMembers(ctx, key)
}

func (rs *RedisStorage) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return rs.cmd.SAdd(ctx, key, members...)
}

func (rs *RedisStorage) SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return rs.cmd.SRem(ctx, key, members...)
}

func (rs *RedisStorage) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	return rs.cmd.SIsMember(ctx, key, member)
}

func (rs *RedisStorage) Get(ctx context.Context, key string) *redis.StringCmd {
	return rs.cmd.Get(ctx, key)
}

func (rs *RedisStorage) MGet(ctx context.Context, keys []string) *redis.SliceCmd {
	return rs.cmd.MGet(ctx, keys...)
}

func (rs *RedisStorage) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	return rs.cmd.IncrBy(ctx, key, value)
}

func (rs *RedisStorage) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return rs.cmd.Set(ctx, key, value, expiration)
}

func (rs *RedisStorage) SCard(ctx context.Context, key string) *redis.IntCmd {
	return rs.cmd.SCard(ctx, key)
}

func (rs *RedisStorage) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return rs.cmd.Expire(ctx, key, expiration)
}

func (rs *RedisStorage) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return rs.cmd.ZAdd(ctx, key, members...)
}

func (rs *RedisStorage) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return rs.cmd.ZRem(ctx, key, members...)
}

func (rs *RedisStorage) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	return rs.cmd.ZRangeByScore(ctx, key, opt)
}

func (rs *RedisStorage) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return rs.cmd.ZRangeWithScores(ctx, key, start, stop)
}

func (rs *RedisStorage) ZCard(ctx context.Context, key string) *redis.IntCmd {
	return rs.cmd.ZCard(ctx, key)
}

func (rs *RedisStorage) ZPopMax(ctx context.Context, key string, count ...int64) *redis.ZSliceCmd {
	return rs.cmd.ZPopMax(ctx, key, count...)
}

func (rs *RedisStorage) ZPopMin(ctx context.Context, key string, count ...int64) *redis.ZSliceCmd {
	return rs.cmd.ZPopMin(ctx, key, count...)
}

func (rs *RedisStorage) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) *redis.ZWithKeyCmd {
	return rs.cmd.BZPopMax(ctx, timeout, keys...)
}

func (rs *RedisStorage) BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) *redis.ZWithKeyCmd {
	return rs.cmd.BZPopMin(ctx, timeout, keys...)
}

func (rs *RedisStorage) Del(ctx context.Context, key string) *redis.IntCmd {
	return rs.cmd.Del(ctx, key)
}

func (rs *RedisStorage) SetNx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return rs.cmd.SetNX(ctx, key, value, expiration)
}

func (rs *RedisStorage) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return rs.cmd.SetEx(ctx, key, value, expiration)
}

func (rs *RedisStorage) HSet(ctx context.Context, key string, field interface{}, value interface{}) *redis.IntCmd {
	return rs.cmd.HSet(ctx, key, field, value)
}

func (rs *RedisStorage) HGet(ctx context.Context, key string, field string) *redis.StringCmd {
	return rs.cmd.HGet(ctx, key, field)
}

func (rs *RedisStorage) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	return rs.cmd.HGetAll(ctx, key)
}

func (rs *RedisStorage) HDel(ctx context.Context, key string, field string) *redis.IntCmd {
	return rs.cmd.HDel(ctx, key, field)
}

func (rs *RedisStorage) HIncrBy(ctx context.Context, key string, field string, incr int64) *redis.IntCmd {
	return rs.cmd.HIncrBy(ctx, key, field, incr)
}

func (rs *RedisStorage) HIncrByFloat(ctx context.Context, key string, field string, incr float64) *redis.FloatCmd {
	return rs.cmd.HIncrByFloat(ctx, key, field, incr)
}

func (rs *RedisStorage) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return rs.cmd.LPush(ctx, key, values...)
}

func (rs *RedisStorage) RPop(ctx context.Context, key string) *redis.StringCmd {
	return rs.cmd.RPop(ctx, key)
}

func (rs *RedisStorage) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return rs.cmd.RPush(ctx, key, values...)
}

func (rs *RedisStorage) LPop(ctx context.Context, key string) *redis.StringCmd {
	return rs.cmd.LPop(ctx, key)
}

func (rs *RedisStorage) Publish(ctx context.Context, channel string, msg interface{}) (err error) {
	return rs.cmd.Publish(ctx, channel, msg).Err()
}

func (rs *RedisStorage) Subscribe(ctx context.Context, channels []string) *redis.PubSub {
	return rs.cmd.Subscribe(ctx, channels...)
}

func (rs *RedisStorage) ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd {
	return rs.cmd.ZIncrBy(ctx, key, increment, member)
}
