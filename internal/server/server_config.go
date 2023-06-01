package server

type Config struct {
	ListenAddr  string `long:"address" short:"a" env:"ADDRESS" default:"127.0.0.1:8080" description:"set listen address"`
	DatabaseDSN string `long:"database" short:"d" env:"DATABASE_DSN" description:"set database dsn"`
	LogLevel    string `long:"log_level" env:"LOG_LEVEL" default:"info" description:"set log level"`
	LogFile     string `long:"log_file" env:"LOG_FILE" default:"" description:"set log file"`
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr: "127.0.0.1:8080",
	}
}
