// api 是路边联盟的后端 HTTP 服务（Go + Gin）。
package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"roadside-league/backend/internal/config"
	"roadside-league/backend/internal/database"
	"roadside-league/backend/internal/ddragon"
	"roadside-league/backend/internal/handler"
	"roadside-league/backend/internal/middleware"
	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
	"roadside-league/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	// 数据库
	db, err := database.Open(cfg.DBDSN)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	// 存储
	store := storage.NewLocalStore(cfg.MediaRoot, cfg.MaxImageMB, cfg.MaxVideoMB)

	// Valkey（Redis 协议）客户端，用于验证码等临时数据
	valkey := redis.NewClient(&redis.Options{Addr: cfg.ValkeyAddr})
	if err := valkey.Ping(context.Background()).Err(); err != nil {
		log.Printf("WARN valkey ping failed: %v", err)
	} else {
		log.Printf("valkey connected: %s", cfg.ValkeyAddr)
	}

	// 依赖装配
	userRepo := repository.NewUserRepository(db)
	heroRepo := repository.NewHeroRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	runeRepo := repository.NewRuneRepository(db)
	contentRepo := repository.NewContentRepository(db)
	settingRepo := repository.NewSettingRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	mailSvc := service.NewMailService(cfg)
	authService := service.NewAuthService(userRepo, valkey, mailSvc, cfg.JWTSecret, cfg.JWTExpireHours, cfg.CodeTTLSeconds, cfg.CodeResendSeconds)
	adminService := service.NewAdminService(heroRepo, assetRepo, store)
	adminUserService := service.NewAdminUserService(userRepo)
	contentService := service.NewContentService(contentRepo, heroRepo, userRepo, assetRepo, runeRepo, settingRepo, store)
	syncService := service.NewSyncService(ddragon.New(), heroRepo, assetRepo, runeRepo, settingRepo, store)
	messageService := service.NewMessageService(messageRepo, userRepo, valkey)

	authHandler := handler.NewAuthHandler(authService)
	heroHandler := handler.NewHeroHandler(heroRepo)
	assetHandler := handler.NewAssetHandler(assetRepo)
	runeHandler := handler.NewRuneHandler(runeRepo, store)
	contentHandler := handler.NewContentHandler(contentService)
	adminHandler := handler.NewAdminHandler(adminService)
	adminUserHandler := handler.NewAdminUserHandler(adminUserService)
	syncHandler := handler.NewSyncHandler(syncService)
	messageHandler := handler.NewMessageHandler(messageService)

	// 确保管理员账号存在（账号密码来自环境变量）
	if err := authService.EnsureAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Printf("WARN ensure admin failed: %v", err)
	}

	// 启动时异步回填英雄元数据：别名（内置表）+ 官方定位（Data Dragon tags）
	syncService.BackfillMeta()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 上传文件静态访问（生产经 Nginx 直挂卷；本地 go run 开发时由 Go 提供）
	r.Static("/media", cfg.MediaRoot)

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			response.OK(c, gin.H{"service": "lol-api", "status": "up"})
		})

		// 公开：注册 / 登录 / 发送邮箱验证码
		api.POST("/auth/send-code", authHandler.SendCode)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// 公开：英雄只读数据（首页头像墙、英雄列表）
		api.GET("/heroes", heroHandler.List)

		// 公开：游戏素材（装备/召唤师技能）、结构化符文树+属性碎片
		api.GET("/assets", assetHandler.List)
		api.GET("/rune-trees", runeHandler.Trees)

		// 公开：内容（视频/图文）。详情挂可选鉴权以识别作者本人/管理员看未过审内容
		api.GET("/contents", contentHandler.List)
		api.GET("/contents/:id", middleware.OptionalAuth(authService), contentHandler.Detail)
		// 从某英雄的已通过内容提取推荐装备符文
		api.GET("/heroes/:heroId/builds", contentHandler.HeroBuilds)

		// 需登录
		api.GET("/me", middleware.RequiredAuth(authService), authHandler.Me)

		// 我的内容（普通用户可查看自己的投稿记录，结果通常为空）+ 留言
		auth := api.Group("", middleware.RequiredAuth(authService))
		{
			auth.GET("/my/contents", contentHandler.Mine)
			// 普通用户给管理员留言
			auth.POST("/messages", messageHandler.Send)
			auth.GET("/messages/mine", messageHandler.Mine)
		}

		// 内容投稿（仅管理员，路径保持 /api/contents 不变）
		adminPost := api.Group("", middleware.RequiredAuth(authService), middleware.RequireAdmin())
		{
			adminPost.POST("/contents", contentHandler.Create)
			adminPost.PUT("/contents/:id", contentHandler.Update)
			adminPost.DELETE("/contents/:id", contentHandler.Remove)
		}

		// 管理后台：登录 + 管理员
		admin := api.Group("/admin", middleware.RequiredAuth(authService), middleware.RequireAdmin())
		{
			admin.GET("/heroes", adminHandler.ListHeroes)
			admin.POST("/heroes", adminHandler.CreateHero)
			admin.POST("/heroes/:id/avatar", adminHandler.UploadHeroAvatar)

			admin.GET("/assets", adminHandler.ListAssets)
			admin.POST("/assets", adminHandler.CreateAsset)
			admin.POST("/assets/:id/icon", adminHandler.UploadAssetIcon)
			admin.DELETE("/assets/:id", adminHandler.DeleteAsset)

			// 内容审批
			admin.GET("/contents", contentHandler.AdminList)
			admin.POST("/contents/:id/approve", contentHandler.Approve)
			admin.POST("/contents/:id/reject", contentHandler.Reject)

			// 用户留言管理
			admin.GET("/messages", messageHandler.AdminList)

			// 用户管理
			admin.GET("/users", adminUserHandler.List)
			admin.POST("/users", adminUserHandler.Add)
			admin.PATCH("/users/:id/status", adminUserHandler.SetStatus)
			admin.PATCH("/users/:id/role", adminUserHandler.SetRole)
			admin.POST("/users/:id/password", adminUserHandler.ResetPassword)
			admin.DELETE("/users/:id", adminUserHandler.Delete)

			// Data Dragon 数据同步（英雄/装备/召唤师技能/符文图标）
			admin.POST("/sync/ddragon", syncHandler.Trigger)
			admin.GET("/sync/status", syncHandler.Status)
		}
	}

	log.Printf("api listening on %s (media: %s)", cfg.Addr, cfg.MediaRoot)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
