// worker 是异步任务进程（二期：消费 Valkey 队列，调用 ffmpeg 转码）。
// 当前为骨架，仅保活并打印等待日志。
package main

import (
	"log"
	"time"

	"roadside-league/backend/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("worker started, valkey=%s media=%s；转码逻辑将在二期接入", cfg.ValkeyAddr, cfg.MediaRoot)

	for range time.Tick(30 * time.Second) {
		log.Println("worker alive: waiting for transcode tasks")
	}
}
