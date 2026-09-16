package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/tidwall/pretty"
)

type Config struct {
	PublicURL string `env:"PUBLIC_URL" env-description:"single public http(s) URL used by the controller and managed components"`
	App       struct {
		UseGvisorNet   bool   `env:"USE_GVISOR_NET" env-default:"false" env-description:"use gvisor netstack for TUN device"`
		GlobalSecret   string `env:"GLOBAL_SECRET" env-description:"at least 32 random characters; used to derive signing keys"`
		CookieAge      int    `env:"COOKIE_AGE" env-default:"86400" env-description:"cookie age in second, default is 1 day"`
		CookieName     string `env:"COOKIE_NAME" env-default:"frp-panel-cookie" env-description:"cookie name"`
		CookiePath     string `env:"COOKIE_PATH" env-default:"/" env-description:"cookie path"`
		CookieDomain   string `env:"COOKIE_DOMAIN" env-default:"" env-description:"cookie domain"`
		CookieSecure   bool   `env:"COOKIE_SECURE" env-default:"true" env-description:"cookie secure"`
		CookieHTTPOnly bool   `env:"COOKIE_HTTP_ONLY" env-default:"true" env-description:"cookie http only"`
		AllowedOrigins string `env:"ALLOWED_ORIGINS" env-description:"comma-separated browser origins allowed for websocket upgrades"`
		EnableRegister bool   `env:"ENABLE_REGISTER" env-default:"false" env-description:"enable register, only allow the first admin to register"`
		GithubProxyUrl string `env:"GITHUB_PROXY_URL" env-description:"optional explicitly trusted github proxy url"`
	} `env-prefix:"APP_"`
	Master struct {
		APIPort   int    `env:"API_PORT" env-default:"9000" env-description:"master api port"`
		APIHost   string `env:"API_HOST" env-description:"master host, can behind proxy like cdn"`
		APIScheme string `env:"API_SCHEME" env-default:"http" env-description:"master api scheme"`
		CacheSize int    `env:"CACHE_SIZE" env-default:"10" env-description:"cache size in MB"`
		RPCHost   string `env:"RPC_HOST" env-default:"127.0.0.1" env-description:"master host, is a public ip or domain"`
		RPCPort   int    `env:"RPC_PORT" env-default:"9001" env-description:"master rpc port"`
	} `env-prefix:"MASTER_"`
	Server struct {
		APIPort int `env:"API_PORT" env-default:"8999" env-description:"server api port"`
	} `env-prefix:"SERVER_"`
	DB struct {
		Type string `env:"TYPE" env-default:"sqlite3" env-description:"database type: sqlite3 or postgres"`
		DSN  string `env:"DSN" env-default:"/data/data.db?_pragma=journal_mode(WAL)" env-description:"SQLite path or PostgreSQL DSN"`
	} `env-prefix:"DB_"`
	Client struct {
		ID                    string `env:"ID" env-description:"client id"`
		Secret                string `env:"SECRET" env-description:"client secret"`
		TLSRpc                bool   `env:"TLS_RPC" env-default:"true" env-description:"use tls for rpc connection"`
		RPCUrl                string `env:"RPC_URL" env-description:"rpc url, support ws or wss or grpc scheme, eg: ws://127.0.0.1:9000"`
		APIUrl                string `env:"API_URL" env-description:"api url, support http or https scheme, eg: http://127.0.0.1:9000"`
		TLSInsecureSkipVerify bool   `env:"TLS_INSECURE_SKIP_VERIFY" env-default:"false" env-description:"skip tls verification (unsafe)"`
		Worker                struct {
			WorkerdBinaryPath  string `env:"WORKERD_BINARY_PATH" env-description:"workerd binary path"`
			WorkerdWorkDir     string `env:"WORKERD_WORK_DIR" env-default:"/tmp/frpp/workerd" env-description:"workerd work dir"`
			WorkerdDownloadURL struct {
				UseProxy   bool   `env:"USE_PROXY" env-default:"false" env-description:"use an explicitly configured proxy"`
				LinuxArm64 string `env:"LINUX_ARM64" env-default:"https://github.com/cloudflare/workerd/releases/download/v1.20250505.0/workerd-linux-arm64.gz"`
				LinuxX8664 string `env:"LINUX_X86_64" env-default:"https://github.com/cloudflare/workerd/releases/download/v1.20250505.0/workerd-linux-64.gz"`
			} `env-prefix:"WORKERD_DOWNLOAD_URL_" env-description:"workerd download url"`
		} `env-prefix:"WORKER_" env-description:"worker's config"`
		Features struct {
			EnableFunctions   bool `env:"ENABLE_FUNCTIONS" env-default:"false" env-description:"enable functions"`
			EnableRemoteShell bool `env:"ENABLE_REMOTE_SHELL" env-default:"false" env-description:"enable remote shell"`
			EnableWireGuard   bool `env:"ENABLE_WIREGUARD" env-default:"false" env-description:"enable WireGuard"`
		} `env-prefix:"FEATURES_" env-description:"features config"`
	} `env-prefix:"CLIENT_"`
	IsDebug bool `env:"IS_DEBUG" env-default:"false" env-description:"is debug mode"`
	Debug   struct {
		ProfilerEnabled bool `env:"PROFILER_ENABLED" env-default:"false" env-description:"enable profiler"`
		ProfilerPort    int  `env:"PROFILER_PORT" env-default:"6961" env-description:"profiler port"`
	} `env-prefix:"DEBUG_"`
	Logger struct {
		DefaultLoggerLevel string `env:"DEFAULT_LOGGER_LEVEL" env-default:"info" env-description:"frp-panel internal default logger level"`
		FRPLoggerLevel     string `env:"FRP_LOGGER_LEVEL" env-default:"info" env-description:"frp logger level"`
	} `env-prefix:"LOGGER_"`
	HTTP_PROXY string `env:"HTTP_PROXY" env-description:"http proxy"`
}

func NewConfig() Config {
	var (
		err        error
		useEnvFile bool
		ctx        = context.Background()
	)

	// 越前面优先级越高，后面的不会覆盖前面的
	envFiles := []string{
		defs.CurEnvPath,
		defs.SysEnvPath,
	}

	for _, envFile := range envFiles {
		if err = godotenv.Load(envFile); err == nil {
			logger.Logger(ctx).Infof("load env file success: %s", envFile)
			useEnvFile = true
		}
	}

	if !useEnvFile {
		logger.Logger(ctx).Info("use runtime env variables")
	}

	cfg := DefaultConfig()
	if err = cleanenv.ReadEnv(&cfg); err != nil {
		logger.Logger(ctx).Panic(err)
	}
	if err := cfg.Complete(); err != nil {
		logger.Logger(ctx).Panic(err)
	}

	if !cfg.IsDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	return cfg
}

// DefaultConfig provides the same safe defaults without reading process
// environment or files. The Agent uses it so controller-side .env files cannot
// leak into a node runtime started from an arbitrary working directory.
func DefaultConfig() Config {
	var cfg Config
	cfg.App.CookieAge = 86400
	cfg.App.CookieName = "frp-panel-cookie"
	cfg.App.CookiePath = "/"
	cfg.App.CookieSecure = true
	cfg.App.CookieHTTPOnly = true
	cfg.Master.APIPort = 9000
	cfg.Master.APIScheme = "http"
	cfg.Master.CacheSize = 10
	cfg.Master.RPCHost = "127.0.0.1"
	cfg.Master.RPCPort = 9001
	cfg.Server.APIPort = 8999
	cfg.DB.Type = defs.DBTypeSQLite3
	cfg.DB.DSN = "/data/data.db?_pragma=journal_mode(WAL)"
	cfg.Client.TLSRpc = true
	cfg.Client.Worker.WorkerdWorkDir = "/tmp/frp-panel/workerd"
	cfg.Logger.DefaultLoggerLevel = "info"
	cfg.Logger.FRPLoggerLevel = "info"
	return cfg
}

func (cfg *Config) Complete() error {
	if err := cfg.applyPublicURL(); err != nil {
		return err
	}
	if len(cfg.Master.APIHost) == 0 {
		cfg.Master.APIHost = cfg.Master.RPCHost
	}

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(cfg.Client.ID) == 0 {
		cfg.Client.ID = hostname
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("failed to get current working directory:", err)
		os.Exit(1)
	}

	if len(cfg.Client.Worker.WorkerdBinaryPath) == 0 {
		w, _ := utils.FindExecutableNames(func(name string) bool {
			return strings.HasPrefix(name, "workerd")
		}, cwd, "/")
		if len(w) > 0 {
			cfg.Client.Worker.WorkerdBinaryPath = w[0]
		}
	}
	return nil
}

func (cfg *Config) applyPublicURL() error {
	value := strings.TrimSpace(cfg.PublicURL)
	if value == "" {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("PUBLIC_URL must be an absolute http(s) URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return fmt.Errorf("PUBLIC_URL must not contain credentials, a path, query, or fragment")
	}
	parsed.Path = ""
	cfg.PublicURL = strings.TrimRight(parsed.String(), "/")
	cfg.Master.APIHost = parsed.Hostname()
	cfg.Master.APIScheme = parsed.Scheme
	cfg.Master.RPCHost = parsed.Hostname()
	cfg.Client.APIUrl = cfg.PublicURL
	if parsed.Scheme == "https" {
		parsed.Scheme = "wss"
	} else {
		parsed.Scheme = "ws"
	}
	cfg.Client.RPCUrl = strings.TrimRight(parsed.String(), "/")
	return nil
}

func (cfg Config) PrintStr() string {
	redacted := cfg
	redacted.App.GlobalSecret = "[redacted]"
	redacted.Client.Secret = "[redacted]"
	redacted.HTTP_PROXY = redactURL(redacted.HTTP_PROXY)
	raw, _ := json.Marshal(redacted)
	return string(pretty.Pretty(raw))
}

func redactURL(value string) string {
	if value == "" {
		return ""
	}
	return "[redacted]"
}
