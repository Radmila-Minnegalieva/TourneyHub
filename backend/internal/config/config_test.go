package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const validYAML = `http:
  address: ":8080"
  cors_origins: [http://localhost:5173]
database:
  host: localhost
  port: 5433
  name: tourneyhub
  ssl_mode: disable
migrations:
  directory: ../migrations
`

func prepareConfig(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", path)
	t.Setenv("DB_USER", "пользователь@тест")
	t.Setenv("DB_PASSWORD", "секрет:/?#@% пароль")
	for _, key := range []string{
		"DB_HOST", "DB_PORT", "HTTP_ADDRESS", "CORS_ALLOWED_ORIGINS", "PUBLIC_WEB_URL",
		"SMTP_ADDRESS", "SMTP_FROM", "SMTP_USER", "SMTP_PASSWORD", "SMTP_ALLOW_INSECURE", "SMTP_DIAL_TIMEOUT", "SMTP_TIMEOUT",
		"WORKER_INTERVAL", "EVENT_RETENTION", "HTTP_READ_HEADER_TIMEOUT", "HTTP_READ_TIMEOUT", "HTTP_IDLE_TIMEOUT", "HTTP_SHUTDOWN_TIMEOUT", "HTTP_BODY_LIMIT", "SSE_POLL_INTERVAL", "SSE_HEARTBEAT_INTERVAL",
		"ADMIN_EMAIL", "ADMIN_NICKNAME", "ADMIN_PASSWORD",
	} {
		t.Setenv(key, "")
	}
	return path
}

func TestRuntimeConfigDefaultsYAMLAndEnvironment(t *testing.T) {
	prepareConfig(t, validYAML)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Public.WebURL != "http://localhost:5173" || cfg.SMTP.Timeout != 15*time.Second || cfg.Worker.Interval != time.Minute || cfg.HTTP.SSEHeartbeatInterval != 15*time.Second {
		t.Fatal("runtime defaults missing")
	}
	if defaults := Default(); defaults.Database.Host != "" || defaults.Database.Port != 0 || defaults.HTTP.Address != "" {
		t.Fatal("required infrastructure settings must not have defaults")
	}
	prepareConfig(t, validYAML+`public:
  web_url: https://example.com/tourney
smtp:
  address: relay.example.com:2525
  from: sender@example.com
  allow_insecure: false
  dial_timeout: 3s
  timeout: 8s
worker:
  interval: 30s
  event_retention: 48h
`)
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Public.WebURL != "https://example.com/tourney" || cfg.SMTP.Address != "relay.example.com:2525" || cfg.SMTP.DialTimeout != 3*time.Second || cfg.Worker.EventRetention != 48*time.Hour {
		t.Fatal("runtime YAML settings missing")
	}
	t.Setenv("PUBLIC_WEB_URL", "https://frontend.example.org")
	t.Setenv("SMTP_ADDRESS", "localhost:1025")
	t.Setenv("SMTP_FROM", "local@example.com")
	t.Setenv("SMTP_USER", "mail-user")
	t.Setenv("SMTP_PASSWORD", "mail-secret")
	t.Setenv("SMTP_ALLOW_INSECURE", "true")
	t.Setenv("SMTP_TIMEOUT", "2s")
	t.Setenv("HTTP_BODY_LIMIT", "8M")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Public.WebURL != "https://frontend.example.org" || cfg.SMTP.Address != "localhost:1025" || cfg.SMTP.From != "local@example.com" || cfg.SMTP.User != "mail-user" || cfg.SMTP.Password != "mail-secret" || !cfg.SMTP.AllowInsecure || cfg.SMTP.Timeout != 2*time.Second || cfg.HTTP.BodyLimit != "8M" {
		t.Fatal("runtime environment did not override YAML")
	}
}

func TestRejectInvalidRuntimeConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		env    string
		value  string
	}{
		{"SMTP secret in YAML", validYAML + "smtp:\n  password: secret\n", "", ""},
		{"SMTP user in YAML", validYAML + "smtp:\n  user: login\n", "", ""},
		{"bootstrap in YAML", validYAML + "bootstrap:\n  password: secret\n", "", ""},
		{"bad public URL", validYAML, "PUBLIC_WEB_URL", "javascript:alert(1)"},
		{"relative public URL", validYAML, "PUBLIC_WEB_URL", "/frontend"},
		{"bad SMTP address", validYAML, "SMTP_ADDRESS", "relay.invalid"},
		{"bad SMTP port", validYAML, "SMTP_ADDRESS", "localhost:70000"},
		{"bad SMTP bool", validYAML, "SMTP_ALLOW_INSECURE", "sometimes"},
		{"bad SMTP duration", validYAML, "SMTP_TIMEOUT", "soon"},
		{"zero timeout", validYAML, "SMTP_TIMEOUT", "0s"},
		{"negative SSE interval", validYAML, "SSE_POLL_INTERVAL", "-1s"},
		{"invalid worker duration", validYAML + "worker:\n  interval: later\n", "", ""},
		{"bad body limit", validYAML, "HTTP_BODY_LIMIT", "0M"},
		{"mail header injection", validYAML, "SMTP_FROM", "sender@example.com\r\nBcc: another@example.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prepareConfig(t, tc.source)
			if tc.env != "" {
				t.Setenv(tc.env, tc.value)
			}
			if _, err := Load(); err == nil {
				t.Fatal("invalid runtime configuration accepted")
			}
		})
	}
}

func TestBootstrapConfigValidation(t *testing.T) {
	prepareConfig(t, validYAML)
	t.Setenv("ADMIN_EMAIL", " Admin@Example.com ")
	t.Setenv("ADMIN_NICKNAME", "admin_user")
	t.Setenv("ADMIN_PASSWORD", "password-123456")
	cfg, err := LoadBootstrap()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Bootstrap.Email != "admin@example.com" || cfg.Bootstrap.Nickname != "admin_user" {
		t.Fatal("bootstrap identity was not loaded")
	}
	for _, tc := range []struct {
		key   string
		value string
	}{
		{"ADMIN_EMAIL", "missing"}, {"ADMIN_NICKNAME", "bad nickname"}, {"ADMIN_PASSWORD", "short"}, {"ADMIN_PASSWORD", strings.Repeat("я", 37)},
	} {
		t.Run(tc.key+tc.value, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, err := LoadBootstrap(); err == nil {
				t.Fatal("invalid bootstrap setting accepted")
			}
		})
	}
}

func TestLoadYAMLAndEnvironment(t *testing.T) {
	path := prepareConfig(t, validYAML)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.Host != "localhost" || cfg.Database.Port != 5433 || cfg.HTTP.Address != ":8080" {
		t.Fatal("настройки YAML не загружены")
	}
	if cfg.Migrations.Directory != filepath.Clean(filepath.Join(filepath.Dir(path), "../migrations")) {
		t.Fatal("путь миграций должен считаться от YAML-файла")
	}
	dsn, err := url.Parse(cfg.Database.URL())
	if err != nil {
		t.Fatal(err)
	}
	password, _ := dsn.User.Password()
	if dsn.User.Username() != os.Getenv("DB_USER") || password != os.Getenv("DB_PASSWORD") || dsn.Query().Get("sslmode") != "disable" {
		t.Fatal("специальные символы секретов должны сохраняться после экранирования")
	}
	t.Setenv("DB_HOST", "postgres")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("HTTP_ADDRESS", ":8082")
	t.Setenv("CORS_ALLOWED_ORIGINS", " http://localhost:8081, ,http://localhost:5173 ")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.Host != "postgres" || cfg.Database.Port != 5432 || cfg.HTTP.Address != ":8082" || len(cfg.HTTP.CORSOrigins) != 2 {
		t.Fatal("переменные окружения должны переопределять настройки YAML")
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		env    string
		value  string
	}{
		{name: "логин обязателен", source: validYAML, env: "DB_USER"},
		{name: "пароль обязателен", source: validYAML, env: "DB_PASSWORD"},
		{name: "секреты запрещены в YAML", source: strings.Replace(validYAML, "  host: localhost", "  host: localhost\n  password: secret", 1)},
		{name: "неизвестное поле", source: validYAML + "unknown: value\n"},
		{name: "несколько документов", source: validYAML + "---\n" + validYAML},
		{name: "некорректный порт", source: validYAML, env: "DB_PORT", value: "invalid"},
		{name: "порт вне диапазона", source: validYAML, env: "DB_PORT", value: "65536"},
		{name: "некорректный SSL", source: strings.Replace(validYAML, "ssl_mode: disable", "ssl_mode: invalid", 1)},
		{name: "пустой путь миграций", source: strings.Replace(validYAML, "directory: ../migrations", "directory: ''", 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prepareConfig(t, tc.source)
			if tc.env != "" {
				t.Setenv(tc.env, tc.value)
			}
			if _, err := Load(); err == nil {
				t.Fatal("ожидалась ошибка конфигурации")
			}
		})
	}
}
