package redis_lock_code

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Chen-cc11/redis_lock_code/utils"
	"github.com/gomodule/redigo/redis"
)

const RedisLockKeyPrefix = "REDIS_LOCK_PREFIX_"

var ErrLockAcquiredByOthers = errors.New("lock is acquired by others")

var ErrNil = redis.ErrNil

// 基于redis实现的分布式锁，不可重入，但保证了对称性
type RedisLock struct {
	LockOptions            // 锁配置
	key         string     // 锁的唯一标识
	token       string     // 锁持有者的唯一标识（由进程ID+协程ID生成）
	client      LockClient // redis 客户端接口
}

// NewRedisLock 创建一个新的 RedisLock 实例
func NewRedisLock(key string, client LockClient, opts ...LockOption) *RedisLock {
	r := RedisLock{
		key:    key,
		token:  utils.GetProcessAndGoroutineIDStr(),
		client: client,
	}

	for _, opt := range opts {
		opt(&r.LockOptions)
	}

	repairLock(&r.LockOptions)
	return &r
}

// ------------------------------------------------------------------

// Lock 加锁
// 非阻塞模式直接尝试一次获取锁，失败立即返回错误
// 阻塞模式下通过blockingLock()方法轮询，直到成功或超时
func (r *RedisLock) Lock(ctx context.Context) error {
	// 不管是不是阻塞模式，先获取一次锁
	err := r.tryLock(ctx)
	if err == nil {
		return nil
	}

	// 非阻塞模式
	if !r.isBlock {
		return err
	}

	// 判断错误的类型是否允许重试
	if !IsRetryableErr(err) {
		return err
	}

	// 阻塞模式
	return r.blockingLock(ctx)
}

// ------------------------------------------------------------------

func IsRetryableErr(err error) bool {
	return errors.Is(err, ErrLockAcquiredByOthers)
}

// ------------------------------------------------------------------

//	---------原子性加锁---------
//
// 通过SetNEX实现原子性操作，若key不存在时则设置值，并设置过期时间（避免死锁）。
// 错误处理： 返回ErrLockAcquiredByOthers表示锁已被占用，该错误可通过IsRetryableErr()判断是否可重试
func (r *RedisLock) tryLock(ctx context.Context) error {

	reply, err := r.client.SetNEX(ctx, r.getLockKey(), r.token, r.expireSeconds)
	if err != nil {
		return err
	}
	if reply != 1 {
		return fmt.Errorf("reply: %d, err: %w", reply, ErrLockAcquiredByOthers)
	}
	return nil
}

func (r *RedisLock) getLockKey() string {
	return RedisLockKeyPrefix + r.key
}

// --------------------------------------------------------------------

// 阻塞模式下持续轮询获取锁
// 轮询机制：通过计时器ticker控制重试频率，避免高频请求redis
// 超时控制： 上下文取消ctx.Done()、最大阻塞时间blockWaitingSeconds
func (r *RedisLock) blockingLock(ctx context.Context) error {
	// 阻塞模式等锁时间上限
	timeoutCh := time.After(time.Duration(r.blockWaitingSeconds) * time.Second)
	// 轮询 定时器ticker ，每个 50 ms 尝试取锁一次
	ticker := time.NewTicker(time.Duration(50) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		select {

		// ctx 终止了
		case <-ctx.Done():
			return fmt.Errorf("lock failed, ctx timeout, err: %w", ctx.Err())

		// 阻塞等锁达到上限时间
		case <-timeoutCh:
			return fmt.Errorf("block waiting timeout, err: %w", ErrLockAcquiredByOthers)

		default:
		}

		// 尝试取锁
		err := r.tryLock(ctx)
		if err == nil {
			return nil
		}

		if !IsRetryableErr(err) {
			return err
		}

	}
	return nil
}

// --------------------------------------------------------------------

// Unlock 解锁
// Lua脚本： 确保检查锁归属+删除锁的原子性操作，防止误删他人持有的锁
func (r *RedisLock) Unlock(ctx context.Context) error {
	keysAndArgs := []interface{}{r.getLockKey(), r.token}
	reply, err := r.client.Eval(ctx, LuaCheckAndDeleteDistributionLock, 1, keysAndArgs)
	if err != nil {
		return err
	}

	if ret, _ := reply.(int64); ret != 1 {
		return errors.New("cannot unlock without ownership of lock")
	}
	return nil
}
