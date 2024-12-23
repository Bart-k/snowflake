package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	// 毫秒级时间戳偏移量，基准时间可以是任意设定的起始时间
	epoch          = 1629980400000 // 例如：2021-08-26 00:00:00 UTC
	machineBits    = 5             // 机器ID长度
	dataCenterBits = 5             // 数据中心ID长度
	sequenceBits   = 12            // 序列号长度
)

const (
	maxMachineID    = -1 ^ (-1 << machineBits)    // 最大机器 ID，5 位机器 ID，最大值为 31
	maxDataCenterID = -1 ^ (-1 << dataCenterBits) // 最大数据中心 ID，5 位数据中心 ID，最大值为 31
	maxSequence     = -1 ^ (-1 << sequenceBits)   // 最大序列号，12 位序列号，最大值为 4095
)

const (
	machineShift    = sequenceBits                                // 序列号偏移
	dataCenterShift = sequenceBits + machineBits                  // 机器 ID 偏移
	timestampShift  = sequenceBits + machineBits + dataCenterBits // 数据中心 ID 偏移
)

// Snowflake struct 用于管理 ID 生成
type Snowflake struct {
	mu            sync.Mutex
	machineID     int64
	dataCenterID  int64
	sequence      int64
	lastTimestamp int64
}

func NewSnowflake(machineID int64, dataCenterID int64) (*Snowflake, error) {
	if machineID < 0 || machineID > maxMachineID {
		return nil, fmt.Errorf("machine ID must be between 0 and %d", maxMachineID)
	}
	if dataCenterID < 0 || dataCenterID > maxDataCenterID {
		return nil, fmt.Errorf("data center ID must be between 0 and %d", maxDataCenterID)
	}
	return &Snowflake{
		machineID:    machineID,
		dataCenterID: dataCenterID,
	}, nil
}

// Generate 生成唯一的 Snowflake ID
func (s *Snowflake) Generate() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取当前时间戳（毫秒）
	timestamp := time.Now().UnixMilli() - epoch

	// 检查时间戳变化，处理序列号溢出
	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			// 如果序列号溢出，则等待下一个毫秒
			for timestamp == s.lastTimestamp {
				timestamp = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		// 重置序列号
		s.sequence = 0
	}

	// 更新最后时间戳
	s.lastTimestamp = timestamp

	// 构建唯一 ID
	id := (timestamp << timestampShift) | (s.dataCenterID << dataCenterShift) | (s.machineID << machineShift) | s.sequence

	return id, nil
}

func main() {
	// 创建 Snowflake 实例，设置机器 ID 为 1，数据中心 ID 为 1
	sf, err := NewSnowflake(1, 1)
	if err != nil {
		fmt.Println("Error creating snowflake:", err)
		return
	}

	// 生成 10 个唯一的 ID
	for i := 0; i < 10; i++ {
		id, err := sf.Generate()
		if err != nil {
			fmt.Println("Error generating ID:", err)
			return
		}
		fmt.Printf("Generated ID %d\n", id)
	}
}
