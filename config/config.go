package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/structs"
)

type Config struct {
	Version string `koanf:"version"`

	ServiceBasicConfig ServiceBasicConfig `koanf:"service_basic_config"`

	REST     ServerREST `koanf:"rest"`
	Postgres Postgres   `koanf:"postgres"`
	Redis    Redis      `koanf:"redis"`

	Carbon Carbon `koanf:"carbon"`

	CorsHosts     string `koanf:"cors_hosts"`
	NominatimAddr string `koanf:"nominatim_addr"`
	SessmsAddr    string `koanf:"sessms_addr"`

	GoroutinePoolMax int `koanf:"goroutine_pool_max"`

	Meerastorage Meerastorage `koanf:"meerastorage"`

	GoogleAnalytics GoogleAnalytics `koanf:"google_analytics"`

	SmartPay string `koanf:"smart_pay"`
	ClickPay string `koanf:"click_pay"`

	JwtSignedSecret string `koanf:"jwt_signed_secret"`

	BytesMatch BytesMatch `koanf:"bytes_match"`

	EnableIngredientAnalysis bool `koanf:"enable_ingredient_analysis"`
}

type ServerREST struct {
	Host     string `koanf:"host"`
	Port     uint   `koanf:"port"`
	CAPath   string `koanf:"ca"`
	CertPath string `koanf:"cert"`
	KeyPath  string `koanf:"key"`
}

type ServiceBasicConfig struct {
	EnableSwaggerDocs        bool   `koanf:"enable_swagger_docs"`
	DomainAddr               string `koanf:"domain_addr"`
	RedirectDomain           string `koanf:"redirect_domain"`
	SingleLogin              bool   `koanf:"single_login"`
	LoginTokenExpireDuration int    `koanf:"login_token_expire_duration"`
	EmailSender              string `koanf:"email_sender"`
	BytesEnv                 string `koanf:"bytes_env"`
}

type Postgres struct {
	Dsn          string `koanf:"dsn"`
	MaxIdleConns int    `koanf:"max_idle_conns"`
	MaxOpenConns int    `koanf:"max_open_conns"`
	LogLevel     string `koanf:"log_level"`
}

type Carbon struct {
	Timezone     string `koanf:"timezone"`
	WeekStartsAt int    `koanf:"week_starts_at"`
}

type Redis struct {
	Url string `koanf:"url"`
}

type Meerastorage struct {
	EndPoint string `koanf:"endpoint"`
	Bucket   string `koanf:"bucket"`
}

type GoogleAnalytics struct {
	CredFile string `koanf:"cred_file"`
	PropId   string `koanf:"prop_id"`
}

type BytesMatch struct {
	Address        string `koanf:"address"`
	ExpireDuration int    `koanf:"expire_duration"`
}

var DefaultConfig = &Config{
	Version: "0.0.0",
}

func LoadWithDefault(configPath string) (Config, error) {
	var cfg Config

	k := koanf.New(".")

	if err := k.Load(structs.Provider(DefaultConfig, "koanf"), nil); err != nil {
		return cfg, err
	}

	return cfg, nil
}
