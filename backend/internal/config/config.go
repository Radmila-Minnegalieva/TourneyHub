// Package config загружает настройки сервиса из YAML и переменных окружения.
package config

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config содержит настройки HTTP, PostgreSQL и миграций.
type Config struct {
	HTTP       HTTP       `yaml:"http"`
	Database   Database   `yaml:"database"`
	Migrations Migrations `yaml:"migrations"`
}

// HTTP задаёт адрес сервера и разрешённые источники запросов браузера.
type HTTP struct {
	Address     string   `yaml:"address"`
	CORSOrigins []string `yaml:"cors_origins"`
}

// Database содержит параметры подключения к PostgreSQL.
// Логин и пароль загружаются только из DB_USER и DB_PASSWORD.
type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
	User     string `yaml:"-"`
	Password string `yaml:"-"`
}

// Migrations задаёт директорию с SQL-файлами goose.
type Migrations struct {
	Directory string `yaml:"directory"`
}

// URL формирует строку подключения, экранируя специальные символы в секретах.
// Возвращённое значение содержит пароль и не должно попадать в логи.
func (db Database) URL() string {
	query := url.Values{"sslmode": {db.SSLMode}}
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(db.User, db.Password),
		Host:     net.JoinHostPort(db.Host, strconv.Itoa(db.Port)),
		Path:     "/" + db.Name,
		RawQuery: query.Encode(),
	}
	return dsn.String()
}

// Load читает CONFIG_PATH или config/config.yaml и проверяет настройки.
// Относительный путь миграций считается от директории YAML-файла.
// DB_HOST, DB_PORT, HTTP_ADDRESS и CORS_ALLOWED_ORIGINS позволяют переопределить
// настройки для контейнеров или локального запуска.
func Load() (Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/config.yaml"
	}
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("открыть конфигурацию %s: %w", path, err)
	}
	defer file.Close()
	var cfg Config
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("прочитать конфигурацию: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Config{}, errors.New("конфигурация должна содержать один YAML-документ")
	}
	cfg.Database.User = os.Getenv("DB_USER")
	cfg.Database.Password = os.Getenv("DB_PASSWORD")
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		cfg.Database.Port, err = strconv.Atoi(port)
		if err != nil {
			return Config{}, errors.New("DB_PORT должен быть целым числом")
		}
	}
	if address := os.Getenv("HTTP_ADDRESS"); address != "" {
		cfg.HTTP.Address = address
	}
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		cfg.HTTP.CORSOrigins = nil
		for _, origin := range strings.Split(origins, ",") {
			if value := strings.TrimSpace(origin); value != "" {
				cfg.HTTP.CORSOrigins = append(cfg.HTTP.CORSOrigins, value)
			}
		}
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	if !filepath.IsAbs(cfg.Migrations.Directory) {
		cfg.Migrations.Directory = filepath.Join(filepath.Dir(path), cfg.Migrations.Directory)
	}
	return cfg, nil
}

func (cfg Config) validate() error {
	if cfg.Database.User == "" || cfg.Database.Password == "" {
		return errors.New("DB_USER и DB_PASSWORD обязательны")
	}
	if cfg.Database.Host == "" || cfg.Database.Name == "" || cfg.Database.Port < 1 || cfg.Database.Port > 65535 {
		return errors.New("database: укажите host, name и port от 1 до 65535")
	}
	switch cfg.Database.SSLMode {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return errors.New("database.ssl_mode: неизвестный режим SSL")
	}
	_, port, err := net.SplitHostPort(cfg.HTTP.Address)
	if err != nil {
		return errors.New("http.address должен иметь формат host:port или :port")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return errors.New("http.address: порт должен быть от 1 до 65535")
	}
	if cfg.Migrations.Directory == "" {
		return errors.New("migrations.directory обязателен")
	}
	return nil
}
