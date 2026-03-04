package utils

import (
	"fmt"
	"sync"
	"time"
)

// Snowflake 雪花算法生成器
type Snowflake struct {
	mu        sync.Mutex
	timestamp int64
	workerID  int64
	serial    int64
}

var (
	snowflake *Snowflake
	once      sync.Once
)

const (
	workerBits  = 5
	serialBits  = 12
	workerMax   = -1 ^ (-1 << workerBits)
	serialMax   = -1 ^ (-1 << serialBits)
	timestampShift = workerBits + serialBits
	workerShift    = serialBits
)

// InitSnowflake 初始化雪花算法
func InitSnowflake(workerID int64) {
	once.Do(func() {
		if workerID < 0 || workerID > workerMax {
			panic(fmt.Sprintf("workerID must be between 0 and %d", workerMax))
		}
		snowflake = &Snowflake{
			timestamp: time.Now().UnixMilli(),
			workerID:  workerID,
			serial:    0,
		}
	})
}

// GenerateID 生成唯一ID
func GenerateID() int64 {
	if snowflake == nil {
		InitSnowflake(0)
	}

	snowflake.mu.Lock()
	defer snowflake.mu.Unlock()

	now := time.Now().UnixMilli()
	if now == snowflake.timestamp {
		snowflake.serial++
		if snowflake.serial > serialMax {
			for now <= snowflake.timestamp {
				now = time.Now().UnixMilli()
			}
			snowflake.serial = 0
		}
	} else {
		snowflake.serial = 0
	}
	snowflake.timestamp = now

	return (now << timestampShift) | (snowflake.workerID << workerShift) | snowflake.serial
}
