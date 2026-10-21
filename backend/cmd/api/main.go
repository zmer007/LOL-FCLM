// api 是路边小站的后端 HTTP 服务（Go + Gin）。
package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"roadside-league/backend/internal/config"
	"roadside-league/backend/internal/database"
	"roadside-league/backend/internal/ddragon"
	"roadside-league/backend/internal/handler"
	"roadside-league/backend/internal/middleware"
	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
	"roadside-league/backend/internal/storage"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// loadDotEnv 简易加载 .env 文件（仅本地 go run 开发用；Docker 由 compose 注入环境变量）。
// 不覆盖已存在的环境变量。
func loadDotEnv() {
	for _, p := range []string{".env", "../.env", "../../.env"} {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			idx := strings.Index(line, "=")
			if idx < 0 {
				continue
			}
			key := strings.TrimSpace(line[:idx])
			val := strings.Trim(strings.TrimSpace(line[idx+1:]), `"'`)
			if key != "" && os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
		f.Close()
		log.Printf("loaded env from %s", p)
		return
	}
}

func main() {
	loadDotEnv()

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
	favoriteRepo := repository.NewFavoriteRepository(db)

	authService := service.NewAuthService(userRepo, valkey, cfg.JWTSecret, cfg.JWTExpireHours)
	adminService := service.NewAdminService(heroRepo, assetRepo, store)
	adminUserService := service.NewAdminUserService(userRepo)
	contentService := service.NewContentService(contentRepo, heroRepo, userRepo, assetRepo, runeRepo, settingRepo, favoriteRepo, store)
	// 本地 Data Dragon 数据包（离线部署，从 assets 目录读取）
	assetsDir := getenv("ASSETS_DIR", getenv("ASSETS_PATH", "/app/assets"))
	localReader, err := ddragon.NewLocalReader(assetsDir)
	if err != nil {
		log.Fatalf("加载本地 Data Dragon 数据失败: %v", err)
	}
	log.Printf("Data Dragon 数据版本: %s", localReader.Version())

	syncService := service.NewSyncService(localReader, heroRepo, assetRepo, runeRepo, settingRepo, store)
	favoriteService := service.NewFavoriteService(favoriteRepo, contentService)

	authHandler := handler.NewAuthHandler(authService)
	heroHandler := handler.NewHeroHandler(heroRepo)
	assetHandler := handler.NewAssetHandler(assetRepo)
	runeHandler := handler.NewRuneHandler(runeRepo)
	contentHandler := handler.NewContentHandler(contentService)
	adminHandler := handler.NewAdminHandler(adminService)
	adminUserHandler := handler.NewAdminUserHandler(adminUserService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)

	// 确保管理员账号存在（账号密码来自环境变量）
	if err := authService.EnsureAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Printf("WARN ensure admin failed: %v", err)
	}

	// 启动时从本地 assets 离线同步 Data Dragon 数据到数据库与媒体目录
	if err := syncService.Run(); err != nil {
		log.Printf("WARN Data Dragon 离线同步失败: %v", err)
	} else {
		log.Printf("Data Dragon 离线同步完成: %s", syncService.State().Message)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 上传文件静态访问（生产经 Nginx 直挂卷；本地 go run 开发时由 Go 提供）
	r.Static("/media", cfg.MediaRoot)
	// Data Dragon 官方图标直接从 assets 提供（不复制到 media，部署即生效）
	r.Static("/ddragon", filepath.Join(assetsDir, localReader.Meta().DirName, "img"))

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			response.OK(c, gin.H{"service": "lol-api", "status": "up"})
		})

		// 公开：当前 Data Dragon 数据版本（离线部署）
		api.GET("/ddragon-version", func(c *gin.Context) {
			v, _ := settingRepo.Get(model.SettingDDragonVersion)
			response.OK(c, gin.H{"version": v})
		})

		// 公开：登录（注册采用邀请制，由管理员在后台创建账号）
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

		// 我的内容（普通用户可查看自己的整理记录，结果通常为空）
		auth := api.Group("", middleware.RequiredAuth(authService))
		{
			auth.GET("/my/contents", contentHandler.Mine)
			// 收藏
			auth.POST("/favorites/:contentId", favoriteHandler.Toggle)
			auth.GET("/favorites/mine", favoriteHandler.ListMine)
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
			admin.POST("/contents/export", contentHandler.Export)
			admin.POST("/contents/import", contentHandler.Import)
			admin.GET("/exports/:dir/:file", contentHandler.DownloadExport)

			// 用户管理
			admin.GET("/users", adminUserHandler.List)
			admin.POST("/users", adminUserHandler.Add)
			admin.PATCH("/users/:id/status", adminUserHandler.SetStatus)
			admin.PATCH("/users/:id/role", adminUserHandler.SetRole)
			admin.POST("/users/:id/password", adminUserHandler.ResetPassword)
			admin.DELETE("/users/:id", adminUserHandler.Delete)
		}
	}

	log.Printf("api listening on %s (media: %s)", cfg.Addr, cfg.MediaRoot)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
