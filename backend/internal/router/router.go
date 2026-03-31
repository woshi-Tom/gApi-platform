package router

import (
	"fmt"

	"gapi-platform/internal/config"
	"gapi-platform/internal/handler"
	"gapi-platform/internal/middleware"
	"gapi-platform/internal/mq"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(
	r *gin.Engine,
	cfg *config.Config,
	db *repository.Database,
	redisClient *repository.RedisClient,
) {
	userRepo := repository.NewUserRepository(db.GetDB())
	tokenRepo := repository.NewTokenRepository(db.GetDB())
	channelRepo := repository.NewChannelRepository(db.GetDB())
	orderRepo := repository.NewOrderRepository(db.GetDB())
	paymentRepo := repository.NewPaymentRepository(db.GetDB())
	auditRepo := repository.NewAuditRepository(db.GetDB())
	vipRepo := repository.NewVIPPackageRepository(db.GetDB())
	rechargeRepo := repository.NewRechargePackageRepository(db.GetDB())
	loginLogRepo := repository.NewLoginLogRepository(db.GetDB())
	apiAccessLogRepo := repository.NewAPIAccessLogRepository(db.GetDB())
	idempRepo := repository.NewIdempotencyRepository(db.GetDB())
	paymentLogRepo := repository.NewPaymentLogRepository(db.GetDB())
	_ = paymentLogRepo

	authService := service.NewAuthService(userRepo, tokenRepo, &cfg.JWT)
	userService := service.NewUserService(userRepo)
	tokenService := service.NewTokenService(tokenRepo)
	tokenService.SetUserRepo(userRepo, vipRepo)
	channelService := service.NewChannelService(channelRepo)
	settingsService := service.NewSettingsService(db.GetDB())
	emailVerificationService := service.NewEmailVerificationService(db.GetDB(), redisClient, &cfg.Email, settingsService)
	captchaService := service.NewSliderCaptchaService(redisClient.Client)

	alipayService := service.NewAlipayService(
		settingsService,
		orderRepo,
		paymentRepo,
		userRepo,
		cfg.Server.Mode,
		fmt.Sprintf("%s/api/v1/payment/callback/alipay", cfg.Server.Frontend),
	)

	userHandler := handler.NewUserHandler(authService, userService, loginLogRepo)
	tokenHandler := handler.NewTokenHandler(tokenService)
	orderHandler := handler.NewOrderHandler(orderRepo, userRepo, paymentRepo, vipRepo, rechargeRepo, idempRepo)
	paymentHandler := handler.NewPaymentHandler(orderRepo, paymentRepo, userRepo, vipRepo, auditRepo, alipayService)
	productHandler := handler.NewProductHandler(vipRepo, rechargeRepo)
	apiHandler := handler.NewAPIHandler(tokenService, channelService, userRepo)
	emailHandler := handler.NewEmailVerificationHandler(emailVerificationService)
	captchaHandler := handler.NewCaptchaHandler(captchaService)
	apiAccessLogHandler := handler.NewAPIAccessLogHandler(apiAccessLogRepo)

	// Init handler for setup wizard
	var mqClient *mq.Client
	if mq.DefaultClient != nil {
		mqClient = mq.DefaultClient()
	}
	initHandler := handler.NewInitHandler(db.GetDB(), redisClient, mqClient, cfg.AdminUsers)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Use(corsMiddleware([]string{cfg.Server.Frontend, cfg.Server.AdminFrontend}))

	// 审计日志中间件
	r.Use(middleware.AuditLog(auditRepo))

	v1 := r.Group("/api/v1")
	{
		user := v1.Group("/user")
		{
			user.POST("/register", userHandler.Register)
			user.POST("/login", userHandler.Login)
		}

		init := v1.Group("/init")
		{
			init.GET("/status", initHandler.GetStatus)
			init.POST("/test-db", initHandler.TestDatabase)
			init.POST("/test-db-with-config", initHandler.TestDatabaseWithConfig)
			init.POST("/init-db", initHandler.InitializeDatabase)
			init.POST("/test-redis", initHandler.TestRedis)
			init.POST("/create-admin", initHandler.CreateAdmin)
		}

		email := v1.Group("/email")
		{
			email.POST("/send-code", emailHandler.SendCode)
			email.POST("/verify-code", emailHandler.VerifyCode)
		}

		auth := v1.Group("/auth")
		{
			auth.POST("/forgot-password", emailHandler.ForgotPassword)
			auth.GET("/reset-password", emailHandler.VerifyResetToken)
			auth.POST("/reset-password", emailHandler.ResetPassword)
		}

		captcha := v1.Group("/captcha")
		{
			captcha.GET("/generate", captchaHandler.Generate)
			captcha.POST("/verify", captchaHandler.Verify)
			captcha.GET("/validate", captchaHandler.ValidateToken)
		}

		userAuth := v1.Group("/user")
		userAuth.Use(middleware.JWTAuth(authService))
		{
			userAuth.GET("/profile", userHandler.GetProfile)
			userAuth.PUT("/profile", userHandler.UpdateProfile)
			userAuth.POST("/change-password", userHandler.ChangePassword)
			userAuth.GET("/quota", userHandler.GetQuota)
			userAuth.GET("/vip/status", userHandler.GetVIPStatus)
			userAuth.GET("/stats/usage", userHandler.GetUsageStats)
		}

		tokens := v1.Group("/tokens")
		tokens.Use(middleware.JWTAuth(authService))
		{
			tokens.GET("", tokenHandler.List)
			tokens.POST("", tokenHandler.Create)
			tokens.DELETE("/:id", tokenHandler.Delete)
		}

		products := v1.Group("/products")
		{
			products.GET("", productHandler.List)
			products.GET("/:id", productHandler.GetByID)
		}

		orders := v1.Group("/orders")
		orders.Use(middleware.JWTAuth(authService))
		{
			orders.GET("", orderHandler.List)
			orders.POST("", orderHandler.Create)
			orders.GET("/:id", orderHandler.GetByID)
			orders.GET("/no/:order_no", orderHandler.GetByOrderNo)
		}

		apiLogs := v1.Group("/logs")
		apiLogs.Use(middleware.JWTAuth(authService))
		{
			apiLogs.GET("", apiAccessLogHandler.List)
		}

		payment := v1.Group("/payment")
		payment.Use(middleware.JWTAuth(authService))
		{
			payment.POST("/alipay", paymentHandler.CreateAlipay)
			payment.GET("/alipay/query/:order_no", paymentHandler.QueryAlipayOrder)
			payment.POST("/alipay/cancel/:order_no", paymentHandler.CancelAlipayOrder)
			payment.GET("/config", paymentHandler.GetPaymentConfig)
		}

		v1.POST("/payment/callback/alipay", paymentHandler.AlipayNotify)

		v1.POST("/chat/completions", middleware.TokenAuth(tokenService), middleware.APIAccessLog(apiAccessLogRepo), apiHandler.ChatCompletions)
		v1.GET("/models", middleware.TokenAuth(tokenService), middleware.APIAccessLog(apiAccessLogRepo), apiHandler.ListModels)
		v1.POST("/embeddings", middleware.TokenAuth(tokenService), middleware.APIAccessLog(apiAccessLogRepo), apiHandler.Embeddings)

		internal := v1.Group("/internal")
		{
			internal.GET("/health", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})

			channels := internal.Group("/channels")
			{
				channels.GET("", handler.NewChannelHandler(channelService, auditRepo).List)
				channels.POST("", handler.NewChannelHandler(channelService, auditRepo).Create)
				channels.PUT("/:id", handler.NewChannelHandler(channelService, auditRepo).Update)
				channels.DELETE("/:id", handler.NewChannelHandler(channelService, auditRepo).Delete)
				channels.POST("/:id/test", handler.NewChannelHandler(channelService, auditRepo).Test)
			}
		}
	}
}

func SetupAdminRoutes(
	r *gin.Engine,
	cfg *config.Config,
	db *repository.Database,
	redisClient *repository.RedisClient,
) {
	userRepo := repository.NewUserRepository(db.GetDB())
	tokenRepo := repository.NewTokenRepository(db.GetDB())
	channelRepo := repository.NewChannelRepository(db.GetDB())
	orderRepo := repository.NewOrderRepository(db.GetDB())
	paymentRepo := repository.NewPaymentRepository(db.GetDB())
	auditRepo := repository.NewAuditRepository(db.GetDB())
	loginLogRepo := repository.NewLoginLogRepository(db.GetDB())
	vipRepo := repository.NewVIPPackageRepository(db.GetDB())
	rechargeRepo := repository.NewRechargePackageRepository(db.GetDB())
	apiAccessLogRepo := repository.NewAPIAccessLogRepository(db.GetDB())

	authService := service.NewAuthService(userRepo, tokenRepo, &cfg.JWT)
	channelService := service.NewChannelService(channelRepo)
	settingsService := service.NewSettingsService(db.GetDB())
	alipayService := service.NewAlipayService(
		settingsService,
		orderRepo,
		paymentRepo,
		userRepo,
		cfg.Server.Mode,
		fmt.Sprintf("%s/api/v1/payment/callback/alipay", cfg.Server.Frontend),
	)

	adminHandler := handler.NewAdminHandler(authService, userRepo, channelService, orderRepo, auditRepo, loginLogRepo, apiAccessLogRepo, cfg.AdminUsers)
	productHandler := handler.NewProductHandler(vipRepo, rechargeRepo)
	settingsHandler := handler.NewSettingsHandler(settingsService, alipayService)

	r.Use(corsMiddleware([]string{cfg.Server.Frontend, cfg.Server.AdminFrontend}))

	v1 := r.Group("/api/v1/admin")
	{
		v1.POST("/login", adminHandler.Login)

		adminAuth := v1.Group("")
		adminAuth.Use(middleware.JWTAuth(authService), middleware.AdminAuth(cfg.Server.AdminSecret))
		{
			adminAuth.GET("/users", adminHandler.ListUsers)
			adminAuth.PUT("/users/:id", adminHandler.UpdateUser)

			adminAuth.GET("/channels", adminHandler.ListChannels)
			adminAuth.POST("/channels", adminHandler.CreateChannel)
			adminAuth.PUT("/channels/:id", adminHandler.UpdateChannel)
			adminAuth.POST("/channels/:id/test", adminHandler.TestChannel)

			adminAuth.GET("/orders", adminHandler.ListOrders)

			adminAuth.GET("/logs/operation", adminHandler.GetAuditLogs)
			adminAuth.GET("/logs/login", adminHandler.GetLoginLogs)
			adminAuth.GET("/test", func(c *gin.Context) { c.String(200, "test") })

			adminAuth.GET("/products", productHandler.ListAll)
			adminAuth.POST("/products", productHandler.Create)
			adminAuth.PUT("/products/:id", productHandler.Update)
			adminAuth.POST("/products/:id/enable", productHandler.Enable)
			adminAuth.POST("/products/:id/disable", productHandler.Disable)

			adminAuth.GET("/stats/overview", adminHandler.GetDashboardStats)
			adminAuth.GET("/stats/trends", adminHandler.GetStatsTrends)

			adminAuth.GET("/stats/user-overview", adminHandler.StatsUserOverview)
			adminAuth.GET("/stats/user-ranking", adminHandler.StatsUserRanking)
			adminAuth.GET("/stats/user-list", adminHandler.StatsUserList)
			adminAuth.GET("/stats/abnormal-users", adminHandler.StatsAbnormalUsers)
			adminAuth.GET("/stats/user/:id/detail", adminHandler.StatsUserDetail)

			adminAuth.POST("/change-password", adminHandler.ChangePassword)

			adminAuth.GET("/settings/email", settingsHandler.GetSMTPConfig)
			adminAuth.PUT("/settings/email", settingsHandler.UpdateSMTPConfig)
			adminAuth.POST("/settings/email/test", settingsHandler.TestSMTPConnection)
			adminAuth.GET("/settings/register", settingsHandler.GetRegisterSettings)
			adminAuth.PUT("/settings/register", settingsHandler.UpdateRegisterSettings)
			adminAuth.GET("/settings/payment", settingsHandler.GetPaymentConfig)
			adminAuth.PUT("/settings/payment", settingsHandler.UpdatePaymentConfig)
		}
	}
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowedOrigin := ""
		for _, o := range allowedOrigins {
			if o != "" && (o == origin || o == "*") {
				allowedOrigin = o
				break
			}
		}

		if allowedOrigin == "" && len(allowedOrigins) > 0 && allowedOrigins[0] != "" {
			allowedOrigin = allowedOrigins[0]
		}

		if allowedOrigin == "" {
			allowedOrigin = "*"
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Admin-Secret, X-Idempotency-Key")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
