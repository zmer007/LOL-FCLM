// api 是路边联盟的后端 HTTP 服务（Go + Gin）。
package main

import (
	"log"

	"github.com/gin-gonic/gin"

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
	store := storage.NewLocalStore(cfg.MediaRoot, cfg.MaxImageMB)

	// 依赖装配
	userRepo := repository.NewUserRepository(db)
	heroRepo := repository.NewHeroRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	settingRepo := repository.NewSettingRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpireHours)
	adminService := service.NewAdminService(heroRepo, assetRepo, store)
	adminUserService := service.NewAdminUserService(userRepo)
	syncService := service.NewSyncService(ddragon.New(), heroRepo, assetRepo, settingRepo, store)

	authHandler := handler.NewAuthHandler(authService)
	adminHandler := handler.NewAdminHandler(adminService)
	adminUserHandler := handler.NewAdminUserHandler(adminUserService)
	syncHandler := handler.NewSyncHandler(syncService)

	// 确保管理员账号存在（账号密码来自环境变量）
	if err := authService.EnsureAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Printf("WARN ensure admin failed: %v", err)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 上传文件静态访问（生产经 Nginx 直挂卷；本地 go run 开发时由 Go 提供）
	r.Static("/media", cfg.MediaRoot)

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			response.OK(c, gin.H{"service": "lol-api", "status": "up"})
		})

		// 公开：注册 / 登录
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// 需登录
		api.GET("/me", middleware.RequiredAuth(authService), authHandler.Me)

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
