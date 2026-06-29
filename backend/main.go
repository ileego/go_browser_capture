package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/crypto/bcrypt"

	"github.com/ileego/go_browser_capture/backend/application"
	"github.com/ileego/go_browser_capture/backend/config"
	"github.com/ileego/go_browser_capture/backend/database"
	_ "github.com/ileego/go_browser_capture/backend/docs"
	"github.com/ileego/go_browser_capture/backend/domain"
	"github.com/ileego/go_browser_capture/backend/handlers"
	"github.com/ileego/go_browser_capture/backend/infrastructure"
	"github.com/ileego/go_browser_capture/backend/middleware"
	"github.com/ileego/go_browser_capture/backend/pool"
	"github.com/ileego/go_browser_capture/backend/scheduler"
	"github.com/ileego/go_browser_capture/backend/websocket"
)

var _ = pgxpool.Pool{}

func main() {
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Connect(&cfg.Database); err != nil {
		log.Printf("Warning: Failed to connect to PostgreSQL: %v", err)
	}

	if err := database.ConnectRedis(&cfg.Redis); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	}

	if database.Pool != nil {
		if err := runMigrations(); err != nil {
			log.Printf("Warning: Failed to run migrations: %v", err)
		}
		if err := initAdminUser(); err != nil {
			log.Printf("Warning: Failed to init admin user: %v", err)
		}
	}

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/admin/*path", func(c *gin.Context) {
		path := c.Param("path")
		if path == "" || path == "/" {
			c.File("./admin/index.html")
			return
		}
		if strings.HasPrefix(path, "/css/") || strings.HasPrefix(path, "css/") {
			c.File("./admin/" + strings.TrimPrefix(path, "/"))
			return
		}
		if strings.HasPrefix(path, "/js/") || strings.HasPrefix(path, "js/") {
			c.File("./admin/" + strings.TrimPrefix(path, "/"))
			return
		}
		if path == "/login.html" || path == "login.html" {
			c.File("./admin/login.html")
			return
		}
		if path == "/index.html" || path == "index.html" {
			c.File("./admin/index.html")
			return
		}
		c.File("./admin/index.html")
	})

	pluginPool := pool.NewPluginPool(time.Second*30, 5)

	wsServer := websocket.NewServer()
	wsServer.SetPluginHandlers(
		func(client *websocket.Client, clientID string) string {
			var pluginInfo *pool.PluginInfo
			if clientID != "" {
				pluginInfo = pluginPool.RegisterWithID(client, clientID)
			} else {
				pluginInfo = pluginPool.Register(client)
			}
			return pluginInfo.ClientID
		},
		pluginPool.Unregister,
		pluginPool.UpdateHeartbeat,
	)

	handler := handlers.NewHandler(wsServer)
	wsServer.SetOnRequestHandler(handler.HandleCaptureRequest)
	go wsServer.Run()

	r.GET("/ws", func(c *gin.Context) {
		wsServer.HandleWebSocket(c.Writer, c.Request)
	})

	taskScheduler := scheduler.NewTaskScheduler(pluginPool, 100, 50, time.Second*30, 3)

	wsServer.SetOnResponseHandler(func(resp *websocket.CaptureResponse) {
		taskScheduler.HandleResponse(resp)
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", handler.HealthCheck)
		v1.POST("/capture", handler.CaptureHandler)
		v1.GET("/status", handler.StatusHandler)

		if database.Pool != nil {
			userRepo := infrastructure.NewPostgresUserRepository(database.Pool)
			roleRepo := infrastructure.NewPostgresRoleRepository(database.Pool)
			userService := domain.NewUserService(userRepo, roleRepo)

			authHandler := handlers.NewAuthHandler(userService)
			v1.POST("/auth/login", authHandler.Login)
			v1.POST("/auth/register", authHandler.Register)

			userHandler := handlers.NewUserHandler(userService)
			authGroup := v1.Group("")
			authGroup.Use(middleware.AuthMiddleware(userRepo))
			{
				authGroup.GET("/auth/me", authHandler.Me)

				authGroup.GET("/users", middleware.PermissionMiddleware("user:read"), userHandler.GetAllUsers)
				authGroup.GET("/users/:id", middleware.PermissionMiddleware("user:read"), userHandler.GetUserByID)
				authGroup.PUT("/users/:id", middleware.PermissionMiddleware("user:write"), userHandler.UpdateUser)
				authGroup.DELETE("/users/:id", middleware.PermissionMiddleware("user:delete"), userHandler.DeleteUser)

				authGroup.GET("/roles", middleware.PermissionMiddleware("user:read"), func(c *gin.Context) {
					roles, err := roleRepo.FindAll(c)
					if err != nil {
						c.JSON(500, gin.H{"success": false, "message": err.Error()})
						return
					}
					c.JSON(200, gin.H{"success": true, "data": roles})
				})

				selectorRepo := infrastructure.NewPostgresSelectorConfigRepository(database.Pool)
				selectorDomainService := domain.NewSelectorConfigService(selectorRepo)
				selectorAppService := application.NewSelectorConfigAppService(selectorDomainService)
				selectorConfigHandler := handlers.NewSelectorConfigHandler(selectorAppService)

				authGroup.POST("/selector-configs", middleware.PermissionMiddleware("config:write"), selectorConfigHandler.Create)
				authGroup.GET("/selector-configs", middleware.PermissionMiddleware("config:read"), selectorConfigHandler.GetAll)
				authGroup.GET("/selector-configs/active", middleware.PermissionMiddleware("config:read"), selectorConfigHandler.GetActive)
				authGroup.GET("/selector-configs/:id", middleware.PermissionMiddleware("config:read"), selectorConfigHandler.GetByID)
				authGroup.GET("/selector-configs/name/:name", middleware.PermissionMiddleware("config:read"), selectorConfigHandler.GetByName)
				authGroup.GET("/selector-configs/domain/:domain", middleware.PermissionMiddleware("config:read"), selectorConfigHandler.GetByDomain)
				authGroup.PUT("/selector-configs/:id", middleware.PermissionMiddleware("config:write"), selectorConfigHandler.Update)
				authGroup.DELETE("/selector-configs/:id", middleware.PermissionMiddleware("config:delete"), selectorConfigHandler.Delete)
				authGroup.POST("/selector-configs/:id/activate", middleware.PermissionMiddleware("config:write"), selectorConfigHandler.Activate)
				authGroup.POST("/selector-configs/:id/deactivate", middleware.PermissionMiddleware("config:write"), selectorConfigHandler.Deactivate)

				batchCaptureAppService := application.NewBatchCaptureAppService(selectorDomainService, taskScheduler)
				batchCaptureHandler := handlers.NewBatchCaptureHandler(batchCaptureAppService)
				authGroup.POST("/batch-capture", batchCaptureHandler.BatchCapture)

				statusHandler := handlers.NewStatusHandler(pluginPool, taskScheduler)
				authGroup.GET("/status/plugins", middleware.PermissionMiddleware("plugin:read"), statusHandler.GetPlugins)
				authGroup.GET("/status/plugins/:id", middleware.PermissionMiddleware("plugin:read"), statusHandler.GetPluginByID)
				authGroup.GET("/status/tasks", middleware.PermissionMiddleware("task:read"), statusHandler.GetTasks)
				authGroup.GET("/status/tasks/:id", middleware.PermissionMiddleware("task:read"), statusHandler.GetTaskByID)
				authGroup.GET("/status/system", middleware.PermissionMiddleware("plugin:read"), statusHandler.GetSystemStatus)
			}
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		database.Close()
		database.CloseRedis()
		os.Exit(0)
	}()

	log.Printf("Server starting on :%s", cfg.Server.Port)
	log.Println("WebSocket server ready at ws://localhost:8080/ws")
	log.Println("Admin panel available at http://localhost:8080/admin")
	log.Println("API documentation available at http://localhost:8080/swagger/index.html")
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS selector_configs (
			id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			domain VARCHAR(255) NOT NULL,
			selector VARCHAR(500) NOT NULL,
			regex TEXT,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_selector_configs_name ON selector_configs(name)`,
		`CREATE INDEX IF NOT EXISTS idx_selector_configs_domain ON selector_configs(domain)`,
		`CREATE INDEX IF NOT EXISTS idx_selector_configs_is_active ON selector_configs(is_active)`,

		`CREATE TABLE IF NOT EXISTS roles (
			id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			permissions TEXT[],
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name)`,

		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(50) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			email VARCHAR(100),
			role_id VARCHAR(36) REFERENCES roles(id),
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id)`,

		`INSERT INTO roles (id, name, description, permissions) VALUES ('admin-role', 'admin', '管理员', ARRAY['user:read', 'user:write', 'user:delete', 'config:read', 'config:write', 'config:delete', 'plugin:read', 'plugin:manage', 'task:read', 'task:manage']) ON CONFLICT (name) DO NOTHING`,
		`INSERT INTO roles (id, name, description, permissions) VALUES ('editor-role', 'editor', '编辑', ARRAY['user:read', 'config:read', 'config:write', 'plugin:read', 'task:read']) ON CONFLICT (name) DO NOTHING`,
		`INSERT INTO roles (id, name, description, permissions) VALUES ('viewer-role', 'viewer', '查看者', ARRAY['user:read', 'config:read', 'plugin:read', 'task:read']) ON CONFLICT (name) DO NOTHING`,
	}

	for _, query := range migrations {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if _, err := database.Pool.Exec(ctx, query); err != nil {
			cancel()
			return err
		}
		cancel()
	}

	log.Println("Migrations executed successfully")
	return nil
}

func initAdminUser() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var count int
	err := database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE username = 'admin'`).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Admin user already exists")
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = database.Pool.Exec(ctx, `INSERT INTO users (username, password, email, role_id, is_active) VALUES ('admin', $1, 'admin@example.com', 'admin-role', true)`, string(hashedPassword))
	if err != nil {
		return err
	}

	log.Println("Admin user created: username=admin, password=admin123")
	return nil
}
