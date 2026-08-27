package database

import (
	"context"
	"fmt"
	"time"

	"arlo-admin/internal/config"
	"arlo-admin/pkg/logger"
	"go.uber.org/zap"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		return nil, fmt.Errorf("failed to connect redis (%s): %w", addr, err)
	}

	RDB = rdb
	logger.Logger.Info("redis connected successfully",
		zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
		zap.Int("db", cfg.DB),
	)
	return rdb, nil
}

func CloseRedis() {
	if RDB != nil {
		RDB.Close()
		logger.Logger.Info("redis connection closed")
	}
}
