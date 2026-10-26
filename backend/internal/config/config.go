// Package config загружает настройки сервиса из YAML и переменных окружения.
package config

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Config содержит настройки HTTP, PostgreSQL и миграций.
type Config struct {
	HTTP       HTTP       `yaml:"http"`
	Database   Database   `yaml:"database"`
	Migrations Migrations `yaml:"migrations"`
	Public     Public     `yaml:"public"`
	SMTP       SMTP       `yaml:"smtp"`
	Worker     Worker     `yaml:"worker"`
	Bootstrap  Bootstrap  `yaml:"-"`
}

// HTTP задаёт адрес сервера и разрешённые источники запросов браузера.
type HTTP struct {
	Address              string        `yaml:"address"`
	CORSOrigins          []string      `yaml:"cors_origins"`
	ReadHeaderTimeout    time.Duration `yaml:"read_header_timeout"`
	ReadTimeout          time.Duration `yaml:"read_timeout"`
	IdleTimeout          time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout      time.Duration `yaml:"shutdown_timeout"`
	SSEPollInterval      time.Duration `yaml:"sse_poll_interval"`
	SSEHeartbeatInterval time.Duration `yaml:"sse_heartbeat_interval"`
	BodyLimit            string        `yaml:"body_limit"`
}

// Public задаёт адрес веб-интерфейса для ссылок из писем и приглашений.
type Public struct {
	WebURL string `yaml:"web_url"`
}

// SMTP содержит параметры доставки почты. Секреты доступны только через окружение.
type SMTP struct {
	Address       string        `yaml:"address"`
	From          string        `yaml:"from"`
	User          string        `yaml:"-"`
	Password      string        `yaml:"-"`
	AllowInsecure bool          `yaml:"allow_insecure"`
	DialTimeout   time.Duration `yaml:"dial_timeout"`
	Timeout       time.Duration `yaml:"timeout"`
}

// Worker задаёт период обработки очереди и срок replay событий.
type Worker struct {
	Interval       time.Duration `yaml:"interval"`
	EventRetention time.Duration `yaml:"event_retention"`
}

// Bootstrap содержит конфигурацию создания первого администратора.
type Bootstrap struct {
	Email    string
	Nickname string
	Password string
}

// Default возвращает только необязательные настройки. Подключение к БД,
// HTTP-адрес и каталог миграций должны быть явно заданы в YAML и окружении.
func Default() Config {
	return Config{
		Public: Public{WebURL: "http://localhost:5173"},
		SMTP:   SMTP{From: "noreply@tourneyhub.local", DialTimeout: 10 * time.Second, Timeout: 15 * time.Second},
		Worker: Worker{Interval: time.Minute, EventRetention: 7 * 24 * time.Hour},
		HTTP:   HTTP{ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, ShutdownTimeout: 10 * time.Second, SSEPollInterval: time.Second, SSEHeartbeatInterval: 15 * time.Second, BodyLimit: "6M"},
	}
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
	cfg := Default()
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("прочитать конфигурацию: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Config{}, errors.New("конфигурация должна содержать один YAML-документ")
	}
	if err := cfg.applyEnvironment(); err != nil {
		return Config{}, err
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
	return cfg.validateRuntime()
}

// applyEnvironment является единственной точкой получения runtime-параметров из окружения.
func (cfg *Config) applyEnvironment() error {
	cfg.Database.User = os.Getenv("DB_USER")
	cfg.Database.Password = os.Getenv("DB_PASSWORD")
	cfg.SMTP.User = os.Getenv("SMTP_USER")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")
	for _, field := range []struct {
		key    string
		target *string
	}{
		{"DB_HOST", &cfg.Database.Host}, {"HTTP_ADDRESS", &cfg.HTTP.Address},
		{"PUBLIC_WEB_URL", &cfg.Public.WebURL}, {"SMTP_ADDRESS", &cfg.SMTP.Address},
		{"SMTP_FROM", &cfg.SMTP.From}, {"HTTP_BODY_LIMIT", &cfg.HTTP.BodyLimit},
	} {
		if value := os.Getenv(field.key); value != "" {
			*field.target = value
		}
	}
	if value := os.Getenv("DB_PORT"); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("DB_PORT должен быть целым числом")
		}
		cfg.Database.Port = port
	}
	if value := os.Getenv("SMTP_ALLOW_INSECURE"); value != "" {
		allow, err := strconv.ParseBool(value)
		if err != nil {
			return errors.New("SMTP_ALLOW_INSECURE должен быть true или false")
		}
		cfg.SMTP.AllowInsecure = allow
	}
	for _, field := range []struct {
		key    string
		target *time.Duration
	}{
		{"SMTP_DIAL_TIMEOUT", &cfg.SMTP.DialTimeout}, {"SMTP_TIMEOUT", &cfg.SMTP.Timeout},
		{"WORKER_INTERVAL", &cfg.Worker.Interval}, {"EVENT_RETENTION", &cfg.Worker.EventRetention},
		{"HTTP_READ_HEADER_TIMEOUT", &cfg.HTTP.ReadHeaderTimeout}, {"HTTP_READ_TIMEOUT", &cfg.HTTP.ReadTimeout},
		{"HTTP_IDLE_TIMEOUT", &cfg.HTTP.IdleTimeout}, {"HTTP_SHUTDOWN_TIMEOUT", &cfg.HTTP.ShutdownTimeout},
		{"SSE_POLL_INTERVAL", &cfg.HTTP.SSEPollInterval}, {"SSE_HEARTBEAT_INTERVAL", &cfg.HTTP.SSEHeartbeatInterval},
	} {
		if value := os.Getenv(field.key); value != "" {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("%s должен быть длительностью, например 15s: %w", field.key, err)
			}
			*field.target = duration
		}
	}
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		cfg.HTTP.CORSOrigins = nil
		for _, origin := range strings.Split(origins, ",") {
			if value := strings.TrimSpace(origin); value != "" {
				cfg.HTTP.CORSOrigins = append(cfg.HTTP.CORSOrigins, value)
			}
		}
	}
	return nil
}

func (cfg Config) validateRuntime() error {
	parsed, err := url.Parse(cfg.Public.WebURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("public.web_url должен быть абсолютным HTTP(S)-адресом веб-интерфейса")
	}
	if cfg.SMTP.Address != "" {
		if err := validateHostPort(cfg.SMTP.Address); err != nil {
			return fmt.Errorf("smtp.address: %w", err)
		}
	}
	from, err := mail.ParseAddress(cfg.SMTP.From)
	if err != nil || from.Address != cfg.SMTP.From || strings.ContainsAny(cfg.SMTP.From, "\r\n") {
		return errors.New("smtp.from должен содержать один e-mail без имени и переносов строки")
	}
	for _, field := range []struct {
		name  string
		value time.Duration
	}{
		{"http.read_header_timeout", cfg.HTTP.ReadHeaderTimeout}, {"http.read_timeout", cfg.HTTP.ReadTimeout},
		{"http.idle_timeout", cfg.HTTP.IdleTimeout}, {"http.shutdown_timeout", cfg.HTTP.ShutdownTimeout},
		{"http.sse_poll_interval", cfg.HTTP.SSEPollInterval}, {"http.sse_heartbeat_interval", cfg.HTTP.SSEHeartbeatInterval},
		{"smtp.dial_timeout", cfg.SMTP.DialTimeout}, {"smtp.timeout", cfg.SMTP.Timeout},
		{"worker.interval", cfg.Worker.Interval}, {"worker.event_retention", cfg.Worker.EventRetention},
	} {
		if field.value <= 0 {
			return fmt.Errorf("%s должен быть положительной длительностью", field.name)
		}
	}
	if !regexp.MustCompile(`(?i)^[1-9][0-9]*(B|K|KB|M|MB|G|GB)?$`).MatchString(cfg.HTTP.BodyLimit) {
		return errors.New("http.body_limit должен задавать положительный размер, например 6M")
	}
	return nil
}

func validateHostPort(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil || host == "" || strings.ContainsAny(host, " \t\r\n") {
		return errors.New("ожидается host:port")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return errors.New("порт должен быть от 1 до 65535")
	}
	return nil
}

// LoadBootstrap загружает обычную конфигурацию и проверяет ADMIN_* для CLI.
func LoadBootstrap() (Config, error) {
	cfg, err := Load()
	if err != nil {
		return Config{}, err
	}
	cfg.Bootstrap = Bootstrap{Email: strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))), Nickname: os.Getenv("ADMIN_NICKNAME"), Password: os.Getenv("ADMIN_PASSWORD")}
	parsed, err := mail.ParseAddress(cfg.Bootstrap.Email)
	if err != nil || parsed.Address != cfg.Bootstrap.Email || len(cfg.Bootstrap.Email) > 254 {
		return Config{}, errors.New("ADMIN_EMAIL must be an email address")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`).MatchString(cfg.Bootstrap.Nickname) {
		return Config{}, errors.New("ADMIN_NICKNAME must contain 3..32 Latin letters, digits or underscores")
	}
	if utf8.RuneCountInString(cfg.Bootstrap.Password) < 12 || len(cfg.Bootstrap.Password) > 72 {
		return Config{}, errors.New("ADMIN_PASSWORD must contain at least 12 characters and at most 72 bytes")
	}
	return cfg, nil
}
