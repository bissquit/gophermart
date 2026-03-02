package config

import (
	"flag"
	"os"
)

// Config structure stores required parameters to run server
type Config struct {
	ServerAddr        string
	DSN               string
	AccrualSystemAddr string
}

// GetDefaultConfig returns default Config structure with
// preconfigured fields
func GetDefaultConfig() *Config {
	return &Config{
		ServerAddr:        ":8080",
		DSN:               "",
		AccrualSystemAddr: "",
	}
}

// GetConfig handles environment and cli options to return ready-to-run Config
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
