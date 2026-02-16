package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr        string
	DSN               string
	AccrualSystemAddr string
}

func GetDefaultConfig() *Config {
	return &Config{
		ServerAddr:        ":8080",
		DSN:               "",
		AccrualSystemAddr: "",
	}
}

func GetConfig() *Config {
	cfg := GetDefaultConfig()

	if envServerAddr := os.Getenv("RUN_ADDRESS"); envServerAddr != "" {
		cfg.ServerAddr = envServerAddr
	}
	if envDSN := os.Getenv("DATABASE_URI"); envDSN != "" {
		cfg.DSN = envDSN
	}
	if envAccrualSystem := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystem != "" {
		cfg.AccrualSystemAddr = envAccrualSystem
	}

	flag.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr,
		"server address in host:port format (default :8080)")
	flag.StringVar(&cfg.DSN, "d", cfg.DSN,
		"Database DSN (default \"\")")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", cfg.AccrualSystemAddr,
		"accrual system endpoint (default \"\")")
	flag.Parse()

	return cfg
}
