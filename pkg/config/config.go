package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/google/wire"
)

// ProviderSet centralizes extraction of Bootstrap's top-level fields
// and the nested Data sub-fields (Database / Redis / Broker). Other packages
// only provide their constructors; they must NOT re-extract these fields.
var ProviderSet = wire.NewSet(
	wire.FieldsOf(new(Bootstrap), "Environment"),
	wire.FieldsOf(new(Bootstrap), "Server"),
	wire.FieldsOf(new(Bootstrap), "Auth"),
	wire.FieldsOf(new(Bootstrap), "Trace"),
	wire.FieldsOf(new(Bootstrap), "Registry"),
	wire.FieldsOf(new(Bootstrap), "Data"),
	wire.FieldsOf(new(Data), "Database"),
	wire.FieldsOf(new(Data), "Cache"),
	wire.FieldsOf(new(Data), "Broker"),
)

// Environment represents the application runtime environment.
// Valid values: dev / prod.
type Environment string

const (
	EnvDevelopment Environment = "dev"
	EnvProduction  Environment = "prod"
)

// IsValid checks whether the environment value is one of the known constants.
func (e Environment) IsValid() bool {
	switch e {
	case EnvDevelopment, EnvProduction:
		return true
	}
	return false
}

// IsDev returns true when the environment is dev.
func (e Environment) IsDev() bool { return e == EnvDevelopment }

// IsProd returns true when the environment is production.
func (e Environment) IsProd() bool { return e == EnvProduction }

// Duration is a custom type that supports parsing time.Duration from strings.
type Duration time.Duration

func (d Duration) AsDuration() time.Duration { return time.Duration(d) }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Duration) UnmarshalJSON(b []byte) error {
	// Handle explicit null case -> set to nil (i.e., not set).
	if string(b) == "null" {
		*d = Duration(0)
		return nil
	}

	// Try parsing as a string (e.g., "7200s").
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		dur, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*d = Duration(dur)
		return nil
	}

	// Try parsing as a number (nanoseconds).
	var num int64
	if err := json.Unmarshal(b, &num); err == nil {
		*d = Duration(time.Duration(num))
		return nil
	}

	return fmt.Errorf("invalid duration format: %s", string(b))
}

// Bootstrap holds configuration fields shared by all services.
type Bootstrap struct {
	Environment Environment `json:"environment,omitempty"`
	Server      Server      `json:"server,omitempty"`
	Data        Data        `json:"data,omitempty"`
	Auth        Auth        `json:"auth,omitempty"`
	Trace       Trace       `json:"trace,omitempty"`
	Registry    Registry    `json:"registry,omitempty"`
}

type Server struct {
	Http HTTP `json:"http,omitempty"`
	Grpc GRPC `json:"grpc,omitempty"`
}

type Data struct {
	Database Database `json:"database,omitempty"`
	Cache    Cache    `json:"cache,omitempty"`
	Broker   Broker   `json:"broker,omitempty"`
}

type Auth struct {
	Secret             string    `json:"secret,omitempty"`
	AccessTokenExpire  *Duration `json:"access_token_expire,omitempty"`
	RefreshTokenExpire *Duration `json:"refresh_token_expire,omitempty"`
}

type Trace struct {
	// Endpoint is the OTLP gRPC collector target as a bare "host:port"
	// (e.g. "127.0.0.1:4317").
	Endpoint string `json:"endpoint,omitempty"`
}

type HTTP struct {
	Network string    `json:"network,omitempty"`
	Addr    string    `json:"addr,omitempty"`
	Timeout *Duration `json:"timeout,omitempty"`
}

type GRPC struct {
	Network string    `json:"network,omitempty"`
	Addr    string    `json:"addr,omitempty"`
	Timeout *Duration `json:"timeout,omitempty"`
}

type Database struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	DBName   string `json:"db_name,omitempty"`
	SSLMode  string `json:"ssl_mode,omitempty"`
}

type Cache struct {
	Network      string    `json:"network,omitempty"`
	Addr         string    `json:"addr,omitempty"`
	ReadTimeout  *Duration `json:"read_timeout,omitempty"`
	WriteTimeout *Duration `json:"write_timeout,omitempty"`
}

// Broker holds configuration for the message broker (NATS, Kafka, ...).
// Type selects the concrete broker implementation ("nats", "kafka", ...);
// the broker package's factory picks the right adapter.
type Broker struct {
	Type string `json:"type,omitempty"`
	Addr string `json:"addr,omitempty"`
	// Kafka-specific fields (extensible).
	Brokers []string `json:"brokers,omitempty"`
	// Additional fields can be added as new broker implementations appear.
}

// ----------------------------------------------------------------

type Registry struct {
	Consul Consul `json:"consul,omitempty"`
}

type Consul struct {
	Address string `json:"address,omitempty"`
	Scheme  string `json:"scheme,omitempty"`
}

// MinIO configuration
type Minio struct {
	Endpoint  string   `json:"endpoint"`
	AccessKey string   `json:"access_key"`
	SecretKey string   `json:"secret_key"`
	UseSSL    bool     `json:"use_ssl"`
	Buckets   []string `json:"buckets"`
}

type Verify struct {
	Code      Code      `json:"code"`
	RateLimit RateLimit `json:"rate_limit"`
	SMS       SMS       `json:"sms"`
	Email     Email     `json:"email"`
}

type Code struct {
	Length         int    `json:"length"`
	ExpireSeconds  int    `json:"expire_seconds"`
	MaxAttempts    int    `json:"max_attempts"`
	HashSecret     string `json:"hash_secret"`
	DebugFixedCode string `json:"debug_fixed_code"`
	AllowDevBypass bool   `json:"allow_dev_bypass"`
}

type RateLimit struct {
	TargetPerMinute int `json:"target_per_minute"`
	TargetPerDay    int `json:"target_per_day"`
	IPPerHour       int `json:"ip_per_hour"`
	DevicePerDay    int `json:"device_per_day"`
}

type SMS struct {
	Provider string  `json:"provider"`
	Aliyun   Aliyun  `json:"aliyun"`
	Tencent  Tencent `json:"tencent"`
}

type Aliyun struct {
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	SignName     string `json:"sign_name"`
	TemplateCode string `json:"template_code"`
}

type Tencent struct {
	SecretID   string `json:"secret_id"`
	SecretKey  string `json:"secret_key"`
	AppID      string `json:"app_id"`
	SignName   string `json:"sign_name"`
	TemplateID string `json:"template_id"`
}

type Email struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromName    string `json:"from_name"`
	FromAddress string `json:"from_address"`
}

type Livekit struct {
	URL       string `json:"url"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

type Typing struct {
	DefaultTTL   *Duration `json:"default_ttl"`
	MinTTL       *Duration `json:"min_ttl"`
	MaxTTL       *Duration `json:"max_ttl"`
	EmitDebounce *Duration `json:"emit_debounce"`
}

type JPush struct {
	AppKey       string `json:"app_key"`
	MasterSecret string `json:"master_secret"`
}

// LoadConfig loads the configuration sources (env vars then file) once and
// scans the result into all provided targets. Pass as many target pointers as
// you need; each is scanned in order. This is the shared entry point for every
// service main() so config is built and resolved a single time per process.
func LoadConfig(flagconf string, targets ...interface{}) error {
	c := config.New(
		config.WithSource(
			env.NewSource(""),
			file.NewSource(flagconf),
		),
		config.WithResolveActualTypes(true),
	)
	if err := c.Load(); err != nil {
		return err
	}
	for _, t := range targets {
		if err := c.Scan(t); err != nil {
			return err
		}
	}
	return nil
}
