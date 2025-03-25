package redis_lock_code

import (
	"context"
	"errors"
	"time"

	//  Redis 客户端，提供连接池和命令执行功能
	"github.com/gomodule/redigo/redis"
)

// 分布式锁的核心操作接口
type LockClient interface {
	// SetNEX 原子性的设置键值（仅当键不存在时），用于获取锁
	SetNEX(ctx context.Context, key, value string, expireSeconds int64) (int64, error)

	// 执行 lua 脚本，用于原子性的释放锁
	Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error)
}

// Client Redis 客户端
type Client struct {
	ClientOptions // 连接配置，通过选项模式注入
	pool          *redis.Pool
}

func NewClient(network, address, password string, opts ...ClientOption) *Client {
	c := Client{
		ClientOptions: ClientOptions{
			network:  network,
			address:  address,
			password: password,
		},
	}

	for _, opt := range opts { // 应用配置选项
		opt(&c.ClientOptions)
	}

	repairClient(&c.ClientOptions) // 修复配置默认值

	pool := c.getRedisPool()
	return &Client{
		pool: pool,
	}
}

// 连接池配置
func (c *Client) getRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.maxIdle,                                         // 最大空闲连接数
		IdleTimeout: time.Duration(c.idleTimeoutSeconds) * time.Second, // 空闲连接超时时间
		Dial: func() (redis.Conn, error) { // 创建新连接的函数
			c, err := c.getRedisConn()
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		MaxActive: c.maxActive, // 最大活跃连接数
		Wait:      c.wait,      // 连接耗尽时，是否等待
		TestOnBorrow: func(c redis.Conn, t time.Time) error { // 连接健康检查
			_, err := c.Do("PING")
			return err
		},
	}
}

// 连接创建，用于连接池的 Dial 函数
func (c *Client) getRedisConn() (redis.Conn, error) {
	if c.address == "" {
		panic("Cannot get redis affress from config")
	}

	var dialOpts []redis.DialOption
	if len(c.password) > 0 {
		dialOpts = append(dialOpts, redis.DialPassword(c.password))
	}
	conn, err := redis.DialContext(context.Background(), c.network, c.address, dialOpts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetConn 获取一个 redis 连接
func (c *Client) GetConn(ctx context.Context) (redis.Conn, error) {
	return c.pool.GetContext(ctx)
}

// 只有 key 不存在时才能 set 成功，set 时携带上超时时间，单位秒
func (c *Client) SetNEX(ctx context.Context, key, value string, expireSeconds int64) (int64, error) {
	if key == "" || value == "" {
		return -1, errors.New("redis SET keyNX or value can't be empty")
	}

	conn, err := c.pool.GetContext(ctx) // 从池中获取连接
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	reply, err := conn.Do("SET", key, value, "EX", expireSeconds, "NX")
	if err != nil {
		return -1, err
	}

	r, _ := reply.(int64)
	return r, nil
}

// Eval 支持使用 lua 脚本
func (c *Client) Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error) {
	args := make([]interface{}, 2+len(keysAndArgs))
	args[0] = src
	args[1] = keyCount
	copy(args[2:], keysAndArgs)

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	return conn.Do("EVAL", args...)
}
