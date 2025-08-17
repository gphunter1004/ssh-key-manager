package router

import (
	"ssh-key-manager/internal/config"
	"ssh-key-manager/internal/handler"
	"ssh-key-manager/internal/middleware"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// Setup은 모든 라우팅을 설정합니다.
func Setup(e *echo.Echo, cfg *config.Config) {
	// JWT 미들웨어 설정
	jwtConfig := echojwt.Config{
		SigningKey:  []byte(cfg.JWTSecret),
		ContextKey:  "user",
		TokenLookup: "header:Authorization:Bearer ",
		ErrorHandler: func(c echo.Context, err error) error {
			return handler.ErrorResponse(c, 401, "invalid or expired jwt")
		},
	}

	// API 그룹
	api := e.Group("/api")

	// 공개 라우트 설정
	setupPublicRoutes(api)

	// 인증된 사용자 라우트 설정
	setupAuthenticatedRoutes(api, jwtConfig)

	// 관리자 라우트 설정
	setupAdminRoutes(api, jwtConfig)
}

// setupPublicRoutes는 인증이 필요없는 공개 라우트를 설정합니다.
func setupPublicRoutes(api *echo.Group) {
	// 인증 관련
	api.POST("/register", handler.Register)
	api.POST("/login", handler.Login)

	// 헬스체크
	api.GET("/health", func(c echo.Context) error {
		return handler.SuccessResponse(c, map[string]string{
			"status":  "ok",
			"service": "SSH Key Manager",
		})
	})
}

// setupAuthenticatedRoutes는 인증이 필요한 라우트를 설정합니다.
func setupAuthenticatedRoutes(api *echo.Group, jwtConfig echojwt.Config) {
	auth := api.Group("")
	auth.Use(echojwt.WithConfig(jwtConfig))

	// 인증 관련
	auth.GET("/validate", handler.ValidateToken)
	auth.POST("/refresh", handler.RefreshToken)
	auth.POST("/logout", handler.Logout)

	// SSH 키 관리
	keys := auth.Group("/keys")
	keys.POST("", handler.CreateKey)
	keys.GET("", handler.GetKey)
	keys.DELETE("", handler.DeleteKey)

	// 사용자 관리
	users := auth.Group("/users")
	users.GET("/me", handler.GetCurrentUser)
	users.PUT("/me", handler.UpdateUserProfile)

	// 서버 관리 (아직 구현되지 않은 핸들러들)
	servers := auth.Group("/servers")
	servers.POST("", handler.CreateServer)
	servers.GET("", handler.GetServers)
	servers.GET("/:id", handler.GetServer)
	servers.PUT("/:id", handler.UpdateServer)
	servers.DELETE("/:id", handler.DeleteServer)
	servers.POST("/:id/test", handler.TestServerConnection)
	servers.POST("/deploy", handler.DeployKeyToServers)
	servers.GET("/deployments", handler.GetDeploymentHistory)
}

// setupAdminRoutes는 관리자 전용 라우트를 설정합니다.
func setupAdminRoutes(api *echo.Group, jwtConfig echojwt.Config) {
	admin := api.Group("/admin")
	admin.Use(echojwt.WithConfig(jwtConfig))
	admin.Use(middleware.AdminRequired)

	// 사용자 관리 (관리자용)
	admin.GET("/users", handler.GetAllUsers)
	admin.GET("/users/:id", handler.GetUserDetail)
	admin.PUT("/users/:id/role", handler.UpdateUserRole)
	admin.DELETE("/users/:id", handler.DeleteUser)
}
