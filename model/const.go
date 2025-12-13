package model

import (
	"time"
)

const (
	REDIS_PREFIX                         = "YTKIDDAPI"         //
	OTP_VERIFICATION_EXPIRY              = 1 * time.Minute     //
	UPDATE_PHONE_OTP_VERIFICATION_EXPIRY = 1 * time.Minute     //
	INIT_PASS_VERIFICATION_EXPIRY        = 30 * time.Minute    //
	RESET_PASS_VERIFICATION_EXPIRY       = 30 * time.Minute    //
	MAX_FILE_SIZE_MB                     = 5 * (1 << 20)       // 5 mb
	USER_AUTH_EXPIRY                     = 30 * 24 * time.Hour // 30 days
	FREE_TIER_PLAN_ID                    = 1                   //
	CHAT_LOG_EXPIRY                      = 30 * 24 * time.Hour // 30 days
	CALL_EXPIRY                          = 90 * time.Second    //
	PDF_MAX_FILE_SIZE_MB                 = 50 * (1 << 20)      // 50 mb
	DEFAULT_IMAGE_CACHE_SECONDS          = 2592000             // 30 days in seconds = 30 * 24 * 60 * 60 = 2592000
	ALLOWED_PREMIUM_FREE_PAGE_COUNT      = 5
	THUMBNAIL_WIDTH                      = 240
	THUMBNAIL_HEIGHT                     = 320
)

const (
	YOUTUBE_MAX_PAGE = int64(1)

	SCOPE_ALL        = "all"
	SCOPE_GIFTS      = "gifts"
	SCOPE_GIFTERS    = "gifters"
	SCOPE_GIFT_STATS = "gift_stats"

	OS_ANDROID = "android"
	OS_IOS     = "ios"
	OS_WINDOWS = "windows"
	OS_UBUNTU  = "ubuntu"
	OS_UNKNOWN = "unknown"

	PAYMENT_PLATFORM_PLAYSTORE = "playstore"
	PAYMENT_PLATFORM_APPSTORE  = "appstore"

	MESSAGE_WEBSOCKET = "websocket"
	MESSAGE_REDIS     = "redis"

	CALL_TYPE_VIDEO = "video"
	CALL_TYPE_VOICE = "voice"
)

var (
	OS_TO_PAYMENT_PLATFORM_MAP = map[string]string{
		OS_ANDROID: PAYMENT_PLATFORM_PLAYSTORE,
		OS_IOS:     PAYMENT_PLATFORM_APPSTORE,
	}
)
