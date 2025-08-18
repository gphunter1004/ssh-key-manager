package router

import (
	"net/http"
	"ssh-key-manager/internal/config"
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/handler"
	"ssh-key-manager/internal/middleware"
	"ssh-key-manager/internal/model"

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
			return c.JSON(http.StatusUnauthorized, dto.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    model.ErrInvalidJWT,
					Message: "유효하지 않거나 만료된 JWT 토큰입니다",
					Details: err.Error(),
				},
			})
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

	// 부서 관리 (기본 조회는 인증된 사용자도 가능)
	departments := auth.Group("/departments")
	departments.GET("", handler.GetDepartments)
	// departments.GET("/tree", handler.GetDepartmentTree)  // 복잡한 트리 기능 제거
	departments.GET("/:id", handler.GetDepartment)

	// 서버 관리
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
	// JWT 검증 + 관리자 권한 확인을 한번에 처리 (통합 미들웨어 사용)
	admin.Use(echojwt.WithConfig(jwtConfig))
	admin.Use(middleware.RequireAdmin()) // ✅ 관리자 미들웨어 활성화

	// 사용자 관리 (관리자용)
	admin.GET("/users", handler.GetAllUsers)
	admin.GET("/users/:id", handler.GetUserDetail)
	admin.PUT("/users/:id/role", handler.UpdateUserRole)
	admin.DELETE("/users/:id", handler.DeleteUser)

	// 부서 관리 (관리자용)
	admin.POST("/departments", handler.CreateDepartment)
	admin.PUT("/departments/:id", handler.UpdateDepartment)
	admin.DELETE("/departments/:id", handler.DeleteDepartment)
	admin.GET("/departments/:id/users", handler.GetDepartmentUsers)
}
