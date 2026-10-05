package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	for _, key := range []string{"DB_HOST", "DB_PORT", "HTTP_ADDRESS", "CORS_ALLOWED_ORIGINS"} {
		t.Setenv(key, "")
	}
	return path
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
		name, source, env, value string
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
