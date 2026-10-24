package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dmhl/internal/api"
	"dmhl/internal/config"
)

var (
	version = "0.1.0"
)

func main() {
	port := flag.Int("port", 0, "覆盖 API 端口")
	flag.Parse()

	home, _ := os.UserHomeDir()
	cfgPath := filepath.Join(home, ".dmhl", "config.json")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if *port > 0 {
		cfg.ServerPort = *port
	}

	fmt.Printf("DMHL v%s — Darling Minecraft Hello Launcher\n", version)
	fmt.Printf("数据目录: %s\n", cfg.DataDir)

	srv := api.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
