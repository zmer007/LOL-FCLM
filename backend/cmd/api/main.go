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
	runeRepo := repository.NewRuneRepository(db)
	buildPlanRepo := repository.NewBuildPlanRepository(db)
	settingRepo := repository.NewSettingRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpireHours)
	adminService := service.NewAdminService(heroRepo, assetRepo, store)
	adminUserService := service.NewAdminUserService(userRepo)
	buildPlanService := service.NewBuildPlanService(heroRepo, assetRepo, runeRepo, buildPlanRepo)
	syncService := service.NewSyncService(ddragon.New(), heroRepo, assetRepo, runeRepo, settingRepo, store)

	authHandler := handler.NewAuthHandler(authService)
	heroHandler := handler.NewHeroHandler(heroRepo)
	assetHandler := handler.NewAssetHandler(assetRepo)
	runeHandler := handler.NewRuneHandler(runeRepo, store)
	buildHandler := handler.NewBuildHandler(buildPlanService)
	adminHandler := handler.NewAdminHandler(adminService)
	adminUserHandler := handler.NewAdminUserHandler(adminUserService)
	syncHandler := handler.NewSyncHandler(syncService)

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

		// 公开：注册 / 登录
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// 公开：英雄只读数据（首页头像墙、英雄列表）
		api.GET("/heroes", heroHandler.List)

		// 公开：游戏素材（装备/召唤师技能）、结构化符文树+属性碎片、英雄装备符文方案
		api.GET("/assets", assetHandler.List)
		api.GET("/rune-trees", runeHandler.Trees)
		api.GET("/heroes/:heroId/plans", buildHandler.ListByHero)

		// 需登录
		api.GET("/me", middleware.RequiredAuth(authService), authHandler.Me)

		// 管理后台：登录 + 管理员
		admin := api.Group("/admin", middleware.RequiredAuth(authService), middleware.RequireAdmin())
		{
			admin.GET("/heroes", adminHandler.ListHeroes)
			admin.POST("/heroes", adminHandler.CreateHero)
			admin.POST("/heroes/:id/avatar", adminHandler.UploadHeroAvatar)
			// 整英雄批量替换装备符文方案（空数组=清空）
			admin.PUT("/heroes/:id/plans", buildHandler.Replace)

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
