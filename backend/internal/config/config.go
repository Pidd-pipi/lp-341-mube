package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config 集中解析环境变量。
type Config struct {
	Port           string `env:"PORT" envDefault:"8080"`
	DBHost         string `env:"DB_HOST" envDefault:"localhost"`
	DBPort         string `env:"DB_PORT" envDefault:"5432"`
	DBName         string `env:"DB_NAME" envDefault:"gbcheckup_db"`
	DBUser         string `env:"DB_USER" envDefault:"gbcheckup_user"`
	DBPassword     string `env:"DB_PASSWORD" envDefault:"gbcheckup_pwd"`
	DBSSLMode      string `env:"DB_SSLMODE" envDefault:"disable"`
	JWTSecret      string `env:"JWT_SECRET" envDefault:"change_me_to_a_long_random_string"`
	JWTExpireHours int    `env:"JWT_EXPIRE_HOURS" envDefault:"72"`
	RateLimitPerMin int   `env:"RATE_LIMIT_PER_MIN" envDefault:"120"`
	UploadDir      string `env:"UPLOAD_DIR" envDefault:"/app/uploads"`
	CORSOrigins    string `env:"APP_CORS_ORIGINS" envDefault:"http://localhost:18941"`
}

// Load 解析环境变量。
func Load() (Config, error) { return env.ParseAs[Config]() }

// DSN 构造 PostgreSQL 连接串。
func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

// CORSOriginsList 解析逗号分隔的 CORS 来源白名单，生产默认不允许通配符。
func (c Config) CORSOriginsList() []string {
	raw := strings.TrimSpace(c.CORSOrigins)
	if raw == "" || raw == "*" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}
