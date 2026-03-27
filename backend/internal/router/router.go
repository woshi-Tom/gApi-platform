package router

import (
	"gapi-platform/internal/config"
	"gapi-platform/internal/handler"
	"gapi-platform/internal/middleware"
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

	authService := service.NewAuthService(userRepo, tokenRepo, &cfg.JWT)
	userService := service.NewUserService(userRepo)
	tokenService := service.NewTokenService(tokenRepo)
	channelService := service.NewChannelService(channelRepo)
	emailVerificationService := service.NewEmailVerificationService(db.GetDB(), redisClient)
	captchaService := service.NewSliderCaptchaService(redisClient.Client)

	userHandler := handler.NewUserHandler(authService, userService, loginLogRepo)
	tokenHandler := handler.NewTokenHandler(tokenService)
	orderHandler := handler.NewOrderHandler(orderRepo, userRepo, paymentRepo)
	paymentHandler := handler.NewPaymentHandler(cfg, orderRepo, paymentRepo, userRepo)
	productHandler := handler.NewProductHandler(vipRepo, rechargeRepo)
	apiHandler := handler.NewAPIHandler(tokenService, channelService, userRepo)
	emailHandler := handler.NewEmailVerificationHandler(emailVerificationService)
	captchaHandler := handler.NewCaptchaHandler(captchaService)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Use(corsMiddleware(cfg.Server.Frontend))

	// 审计日志中间件
	r.Use(middleware.AuditLog(auditRepo))

	v1 := r.Group("/api/v1")
	{
		user := v1.Group("/user")
		{
			user.POST("/register", userHandler.Register)
			user.POST("/login", userHandler.Login)
		}

		email := v1.Group("/email")
		{
			email.POST("/send-code", emailHandler.SendCode)
			email.POST("/verify-code", emailHandler.VerifyCode)
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
		}

		payment := v1.Group("/payment")
		payment.Use(middleware.JWTAuth(authService))
		{
			payment.POST("/alipay", paymentHandler.CreateAlipay)
			payment.POST("/wechat", paymentHandler.CreateWechat)
		}

		v1.POST("/payment/callback/alipay", paymentHandler.AlipayCallback)
		v1.POST("/payment/callback/wechat", paymentHandler.WechatCallback)

		v1.POST("/chat/completions", middleware.TokenAuth(tokenService), apiHandler.ChatCompletions)
		v1.GET("/models", middleware.TokenAuth(tokenService), apiHandler.ListModels)
		v1.POST("/embeddings", middleware.TokenAuth(tokenService), apiHandler.Embeddings)

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
	auditRepo := repository.NewAuditRepository(db.GetDB())
	loginLogRepo := repository.NewLoginLogRepository(db.GetDB())
	vipRepo := repository.NewVIPPackageRepository(db.GetDB())
	rechargeRepo := repository.NewRechargePackageRepository(db.GetDB())

	authService := service.NewAuthService(userRepo, tokenRepo, &cfg.JWT)
	channelService := service.NewChannelService(channelRepo)

	adminHandler := handler.NewAdminHandler(authService, userRepo, channelService, orderRepo, auditRepo, loginLogRepo, cfg.AdminUsers)
	productHandler := handler.NewProductHandler(vipRepo, rechargeRepo)

	r.Use(corsMiddleware(cfg.Server.Frontend))

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

			adminAuth.GET("/products", productHandler.List)
			adminAuth.POST("/products/:id/enable", productHandler.Enable)
			adminAuth.POST("/products/:id/disable", productHandler.Disable)

			adminAuth.GET("/stats/overview", adminHandler.GetDashboardStats)

			adminAuth.POST("/change-password", adminHandler.ChangePassword)
		}
	}
}

func corsMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Admin-Secret")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
