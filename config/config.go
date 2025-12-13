package config

import (
	"os"
	"strconv"

	"github.com/subosito/gotenv"
)

var (
	config Config
)

func Initialize() {
	gotenv.Load()

	config = Config{
		AppEnv:            getEnvStringWithDefault("APP_ENV", "development"),
		AppPort:           getEnvStringWithDefault("APP_PORT", "33000"),
		AppHost:           getEnvStringWithDefault("APP_HOST", "http://localhost:33000"),
		DbURL:             os.Getenv("DB_URL"),
		DbTimezone:        getEnvStringWithDefault("DB_TIMEZONE", "Asia/Jakarta"),
		RedisUrl:          os.Getenv("REDIS_URL"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		ChatWebsocketHost: getEnvStringWithDefault("CHAT_WEBSOCKET_HOST", "ws://localhost:33000/jodoh-express/api/chat/websocket/connect"),
		YoutubeApiKey:     os.Getenv("YOUTUBE_API_KEY"),
		FileBucketPath:    getEnvStringWithDefault("FILE_BUCKET_PATH", "file_bucket"),
		CacheDirPath:      getEnvStringWithDefault("CACHE_DIR_PATH", "file_bucket/cache"),
		OllamaHost:        getEnvStringWithDefault("OLLAMA_HOST", "http://localhost:11434"),
		ComfyUIOutputDir:  getEnvStringWithDefault("COMFY_UI_OUTPUT_DIR", "/Users/umar/umar/personal_project/dev-notes/local_app/ComfyUI/output"),

		CkAuthJwtPrivateKey: os.Getenv("CK_AUTH_JWT_PRIVATE_KEY"),
		CkAuthJwtPublicKey:  os.Getenv("CK_AUTH_JWT_PUBLIC_KEY"),
		CkAuthJweSecretKey:  os.Getenv("CK_AUTH_JWE_SECRET_KEY"),

		ChatTokenJwtPrivateKey: os.Getenv("CHAT_TOKEN_JWT_PRIVATE_KEY"),
		ChatTokenJwtPublicKey:  os.Getenv("CHAT_TOKEN_JWT_PUBLIC_KEY"),

		InternalClientID:  os.Getenv("INTERNAL_CLIENT_ID"),
		InternalSecretKey: os.Getenv("INTERNAL_SECRET_KEY"),

		DevInternalClientID:  os.Getenv("DEV_INTERNAL_CLIENT_ID"),
		DevInternalSecretKey: os.Getenv("DEV_INTERNAL_SECRET_KEY"),

		MidtransLive:       os.Getenv("MIDTRANS_LIVE") == "true",
		MidtransMerchantID: os.Getenv("MIDTRANS_MERCHANT_ID"),
		MidtransClientKey:  os.Getenv("MIDTRANS_CLIENT_KEY"),
		MidtransServerKey:  os.Getenv("MIDTRANS_SERVER_KEY"),

		R2TokenValue:        os.Getenv("R2_TOKEN_VALUE"),
		R2AccessKeyId:       os.Getenv("R2_ACCESS_KEY_ID"),
		R2AccessKeySecret:   os.Getenv("R2_ACCESS_KEY_SECRET"),
		R2StorageEndpoint:   os.Getenv("R2_STORAGE_ENDPOINT"),
		R2BucketName:        getEnvStringWithDefault("R2_BUCKET_NAME", "cabocil-bucket"),
		R2PublicDomain:      getEnvStringWithDefault("R2_PUBLIC_DOMAIN", "https://data.cabocil.com"),
		R2UserApiToken:      os.Getenv("R2_USER_API_TOKEN"),
		CloudflareZoneId:    os.Getenv("CLOUDFLARE_ZONE_ID"),
		CloudflareAccountId: os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
	}
}

func Get() Config {
	return config
}

func getEnvStringWithDefault(env, def string) string {
	if os.Getenv(env) == "" {
		return def
	}
	return os.Getenv(env)
}

func getEnvIntWithDefault(env string, def int) int {
	i, err := strconv.ParseInt(os.Getenv(env), 10, 64)
	if err != nil {
		return def
	}
	return int(i)
}
