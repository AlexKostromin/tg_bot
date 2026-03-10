package config

import "github.com/caarlos0/env/v11"

type Config struct {
	BotToken    string `env:"BOT_TOKEN,required"`
	WebhookURL  string `env:"WEBHOOK_URL"` // пусто = режим Polling
	WebhookPort int    `env:"WEBHOOK_PORT"  envDefault:"8443"`

	DatabaseURL string `env:"DATABASE_URL,required"`

	RedisAddr     string `env:"REDIS_ADDR"     envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB"       envDefault:"0"`

	RabbitMQURL string `env:"RABBITMQ_URL" envDefault:"amqp://guest:guest@localhost:5672/"`

	AdminChatID int64  `env:"ADMIN_CHAT_ID,required"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	MaxActiveBookings int `env:"MAX_ACTIVE_BOOKINGS" envDefault:"3"`
	HTTPPort       string `env:"HTTP_PORT"          envDefault:"8080"`
	AdminJWTSecret string `env:"ADMIN_JWT_SECRET,required"`
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
	return cfg
}
