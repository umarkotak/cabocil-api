package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/config"
	"github.com/umarkotak/ytkidd-api/cron"
	"github.com/umarkotak/ytkidd-api/datastore"
	"github.com/umarkotak/ytkidd-api/handlers/ai_handler"
	"github.com/umarkotak/ytkidd-api/handlers/book_handler"
	"github.com/umarkotak/ytkidd-api/handlers/comfy_ui_handler"
	"github.com/umarkotak/ytkidd-api/handlers/flash_card_handler"
	"github.com/umarkotak/ytkidd-api/handlers/gold_handler"
	"github.com/umarkotak/ytkidd-api/handlers/kemendikbud_handler"
	"github.com/umarkotak/ytkidd-api/handlers/order_handler"
	"github.com/umarkotak/ytkidd-api/handlers/ping_handler"
	"github.com/umarkotak/ytkidd-api/handlers/poki_handler"
	"github.com/umarkotak/ytkidd-api/handlers/product_handler"
	"github.com/umarkotak/ytkidd-api/handlers/user_handler"
	"github.com/umarkotak/ytkidd-api/handlers/utils_handler"
	"github.com/umarkotak/ytkidd-api/handlers/youtube_channel_handler"
	"github.com/umarkotak/ytkidd-api/handlers/youtube_handler"
	"github.com/umarkotak/ytkidd-api/handlers/youtube_video_handler"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/book_content_repo"
	"github.com/umarkotak/ytkidd-api/repos/book_repo"
	"github.com/umarkotak/ytkidd-api/repos/file_bucket_repo"
	"github.com/umarkotak/ytkidd-api/repos/flash_card_repo"
	"github.com/umarkotak/ytkidd-api/repos/google_repo"
	"github.com/umarkotak/ytkidd-api/repos/order_repo"
	"github.com/umarkotak/ytkidd-api/repos/product_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_activity_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_stroke_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_subscription_repo"
	"github.com/umarkotak/ytkidd-api/repos/youtube_channel_repo"
	"github.com/umarkotak/ytkidd-api/repos/youtube_video_repo"
	"github.com/umarkotak/ytkidd-api/utils/log_formatter"
	"github.com/umarkotak/ytkidd-api/utils/log_hook"
	"github.com/umarkotak/ytkidd-api/utils/middlewares"
	"github.com/umarkotak/ytkidd-api/utils/payment_lib"
	"github.com/umarkotak/ytkidd-api/utils/ratelimit_lib"
	"github.com/umarkotak/ytkidd-api/utils/redis_lock"
	"github.com/umarkotak/ytkidd-api/utils/user_auth"
	"github.com/umarkotak/ytkidd-api/utils/word_censor_lib"
	"github.com/umarkotak/ytkidd-api/worker"
)

func main() {
	quickExec()

	logrus.SetReportCaller(true)
	logrus.SetFormatter(&log_formatter.Formatter{})
	logrus.AddHook(&log_hook.LogrusHook{})

	config.Initialize()
	datastore.Initialize()
	initializeDependencies()

	args := os.Args[1:]

	if len(args) == 0 {
		runServer()

	} else if len(args) == 1 {
		switch args[0] {
		case "gen_base_key":
			user_auth.GenBaseKey()
		case "scrap_kemendikbud":
			kemendikbud_handler.Scrap()
		}

	} else if len(args) == 2 && args[0] == "migrate" {
		switch args[1] {
		case "up":
			datastore.MigrateUp()
		default:
			fmt.Println("Unknown migrate command. Use 'up' or 'seed'.")
		}

	} else {
		fmt.Println("Usage: go run .")
		fmt.Println("       go run . migrate up")
		fmt.Println("       go run . gen_base_key")
		fmt.Println("       go run . scrap_kemendikbud")
	}
}

func runServer() {
	logrus.Infof("starting ytkidd backend")

	cron.Start()

	go worker.Run()

	go initializeHttpServer()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan, os.Interrupt, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM,
	)

	sig := <-sigChan
	logrus.Infof("received %v signal. graceful shutdown...", sig)

	// execute gracefull shutdown

	logrus.Infof("cleanup completed. exiting...")
}

func initializeDependencies() {
	var err error

	user_auth.Initialize(user_auth.UserAuth{
		JwtPrivateKey: config.Get().CkAuthJwtPrivateKey,
		JwtPublicKey:  config.Get().CkAuthJwtPublicKey,
		JweSecretKey:  config.Get().CkAuthJweSecretKey,
	})

	err = ratelimit_lib.Initialize(ratelimit_lib.RateLimiter{
		Prefix:    model.REDIS_PREFIX,
		RedisAddr: config.Get().RedisUrl,
		RedisPwd:  config.Get().RedisPassword,
	})
	if err != nil {
		logrus.WithContext(context.TODO()).Error(err)
	}

	err = redis_lock.Initialize(redis_lock.RedisLock{
		Prefix:    model.REDIS_PREFIX,
		RedisAddr: config.Get().RedisUrl,
		RedisPwd:  config.Get().RedisPassword,
	})
	if err != nil {
		logrus.WithContext(context.TODO()).Error(err)
	}

	err = worker.Initialize(config.Get().RedisUrl, config.Get().RedisPassword)
	if err != nil {
		logrus.WithContext(context.TODO()).Error(err)
	}

	err = cron.Initialize(config.Get().RedisUrl, config.Get().RedisPassword)
	if err != nil {
		logrus.WithContext(context.TODO()).Error(err)
	}

	payment_lib.Initialize(datastore.Get().Db, payment_lib.PaymentConf{
		IsLive:             config.Get().MidtransLive,
		MidtransMerchantID: config.Get().MidtransMerchantID,
		MidtransClientKey:  config.Get().MidtransClientKey,
		MidtransServerKey:  config.Get().MidtransServerKey,
	})

	// Repositories
	youtube_channel_repo.Initialize()
	youtube_video_repo.Initialize()
	book_repo.Initialize()
	book_content_repo.Initialize()
	file_bucket_repo.Initialize()
	user_repo.Initialize()
	google_repo.Initialize()
	product_repo.Initialize()
	order_repo.Initialize()
	user_subscription_repo.Initialize()
	user_stroke_repo.Initialize()
	user_activity_repo.Initialize()
	flash_card_repo.Initialize()

	word_censor_lib.Initialize(word_censor_lib.WordCensorLib{
		Words: []string{"kucing", "anjing", "gajah"},
	})
}

func initializeHttpServer() {
	poki_handler.Initialize()
	gold_handler.Initialize()

	r := chi.NewRouter()

	r.Use(
		chiMiddleware.RequestID, //
		chiMiddleware.RealIP,    //
		chiMiddleware.Recoverer, //
		middlewares.Cors,        //
		// middlewares.RequestLog,  //
		middlewares.CommonCtx, // it will extract headers and put the value to common context
		// middlewares.ReqRateLimit(1000, 1*time.Second), // max 100 request per second based on X-Device-Id
	)

	r.Route("/ytkidd/api", func(ri chi.Router) {
		// rDevInternal := ri.With(middlewares.InternalDevAuth)
		rUserAuth := ri.With(middlewares.UserAuth)
		rOptionalUserAuth := ri.With(middlewares.OptionalUserAuth)
		rAdminAuth := ri.With(middlewares.UserAuthAdmin)

		ri.Get("/ping", ping_handler.Ping)

		rOptionalUserAuth.Get("/youtube_videos", youtube_video_handler.GetYoutubeVideos)
		rOptionalUserAuth.Get("/youtube_video/{id}", youtube_video_handler.GetYoutubeVideoDetail)
		rAdminAuth.Delete("/youtube_videos/{id}", youtube_video_handler.DeleteVideo)
		rOptionalUserAuth.Get("/youtube_channels", youtube_channel_handler.GetYoutubeChannels)
		rOptionalUserAuth.Get("/youtube_channel/{id}", youtube_channel_handler.GetYoutubeChannelDetail)
		rAdminAuth.Get("/youtube_channel/{id}/detailed", youtube_channel_handler.GetYoutubeChannelDetailed)
		rAdminAuth.Patch("/youtube_channel/{id}", youtube_channel_handler.UpdateYoutubeChannel)
		rAdminAuth.Patch("/youtube_channel/{id}/active", youtube_channel_handler.UpdateYoutubeChannelActive)
		rAdminAuth.Delete("/youtube_channel/{id}", youtube_channel_handler.DeleteYoutubeChannel)

		rAdminAuth.Post("/youtube/scrap_videos", youtube_handler.ScrapVideos)

		rAdminAuth.Get("/books/upload_status", book_handler.GetBooksUploadStatus)
		rAdminAuth.Post("/books/insert_from_pdf", book_handler.InsertFromPdf)
		rAdminAuth.Post("/books/insert_from_pdf_urls", book_handler.InsertFromPdfUrls)
		rOptionalUserAuth.Get("/books", book_handler.GetBooks)
		rOptionalUserAuth.Get("/book/{slug}", book_handler.GetBookDetail)
		rAdminAuth.Patch("/book/{id}", book_handler.UpdateBook)
		rAdminAuth.Patch("/book/{id}/cover", book_handler.UpdateBookCover)
		rAdminAuth.Delete("/book/{id}", book_handler.DeleteBook)
		rAdminAuth.Post("/book/{id}/page/remove", book_handler.RemoveBookPage)

		ri.Get("/comfy_ui/output", comfy_ui_handler.Gallery)
		ri.Post("/ai/chat", ai_handler.Chat)

		ri.Post("/user/sign_in", user_handler.SignIn)
		rUserAuth.Get("/user/check_auth", user_handler.CheckAuth)
		rUserAuth.Get("/user/profile", user_handler.MyProfile)
		rUserAuth.Get("/user/subscription", user_handler.MySubscription)
		rAdminAuth.Get("/users", user_handler.AdminGetUsers)
		rAdminAuth.Get("/users/all_activities/recent", user_handler.AdminGetUserActivities)

		ri.Get("/products", product_handler.GetProducts)

		rUserAuth.Post("/order/create", order_handler.PostCreateOrder)
		rUserAuth.Post("/order/{order_number}/payment/check", order_handler.PostCheckOrderPayment)
		rUserAuth.Get("/order/list", order_handler.GetOrderList)
		rUserAuth.Get("/order/{order_number}", order_handler.GetOrderDetail)

		rOptionalUserAuth.Get("/book/user_stroke", book_handler.GetUserStroke)
		rOptionalUserAuth.Post("/book/user_stroke", book_handler.StoreUserStroke)

		rOptionalUserAuth.Get("/user/activities", user_handler.GetUserActivities)
		rOptionalUserAuth.Post("/user/activity", user_handler.PostUserActivity)

		ri.Post("/midtrans/callback/transaction", payment_lib.MidtransCallbackHandler)

		ri.Get("/poki/games", poki_handler.GetGameList)
		ri.Get("/poki/games/detail", poki_handler.GetGameDetail)

		ri.Get("/flash_card/{id}", flash_card_handler.GetByID)
		ri.Get("/flash_cards", flash_card_handler.GetByTags)
		rAdminAuth.Post("/flash_cards/bulk", flash_card_handler.BulkInsert)

		ri.Get("/utils/image/compress", utils_handler.CompressHandler)
		ri.Delete("/utils/image/cache", utils_handler.DeleteCacheHandler)

		ri.Get("/gold/price/today", gold_handler.GetTodayPrice)
	})

	r.Get("/file_bucket/*", func(w http.ResponseWriter, r *http.Request) {
		// logrus.Infof("Image HIT")
		w.Header().Set(model.ImgCacheHeader, model.ImgCacheDuration7Days)
		fs := http.StripPrefix("/file_bucket", http.FileServer(http.Dir(config.Get().FileBucketPath)))
		fs.ServeHTTP(w, r)
	})
	r.Get("/comfy_ui_gallery/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/comfy_ui_gallery", http.FileServer(http.Dir(config.Get().ComfyUIOutputDir))).ServeHTTP(w, r)
	})

	port := fmt.Sprintf(":%s", config.Get().AppPort)
	logrus.Infof("running http server on port %s", port)

	err := http.ListenAndServe(port, r)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatalf("fail to start http server on port %s", port)
	}
}
