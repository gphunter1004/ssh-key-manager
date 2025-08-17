package controllers

import (
	"fmt"
	"ssh-key-manager/helpers"
	"ssh-key-manager/services"
	"ssh-key-manager/types"
	"ssh-key-manager/utils"

	"github.com/labstack/echo/v4"
)

// Register godoc
// @Summary Register a new user
// @Description Register a new user with username and password
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   user  body   types.AuthRequest  true  "User Info"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /register [post]
func Register(c echo.Context) error {
	var req types.AuthRequest
	if err := c.Bind(&req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	// 입력값 검증
	if err := validateAuthRequest(req); err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	if err := services.RegisterUser(req.Username, req.Password); err != nil {
		// 사용자 친화적 에러 메시지 처리
		if isUserFriendlyError(err) {
			return helpers.BadRequestResponse(c, err.Error())
		}
		return helpers.InternalServerErrorResponse(c, "사용자 등록 중 오류가 발생했습니다")
	}

	// 성공 응답 - 보안상 사용자 정보는 포함하지 않음
	responseData := map[string]interface{}{
		"message": "사용자가 성공적으로 등록되었습니다",
		"username": req.Username,
	}

	return helpers.CreatedResponse(c, "회원가입이 완료되었습니다", responseData)
}

// Login godoc
// @Summary Log in a user
// @Description Log in with username and password to get a JWT
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   user  body   types.AuthRequest  true  "Credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /login [post]
func Login(c echo.Context) error {
	var req types.AuthRequest
	if err := c.Bind(&req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	// 입력값 검증
	if err := validateAuthRequest(req); err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	token, err := services.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		// 보안상 구체적인 실패 이유는 노출하지 않음
		return helpers.UnauthorizedResponse(c, "사용자명 또는 비밀번호가 올바르지 않습니다")
	}

	// 성공 응답 데이터
	responseData := map[string]interface{}{
		"token": token,
		"username": req.Username,
		"message": "로그인이 성공했습니다",
	}

	return helpers.SuccessWithMessageResponse(c, "로그인 성공", responseData)
}

// Logout godoc
// @Summary Log out a user
// @Description Log out the current user (client-side token removal)
// @Tags auth
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /logout [post]
func Logout(c echo.Context) error {
	// JWT는 stateless이므로 서버에서 실제로 할 일은 없음
	// 클라이언트에서 토큰을 제거하라는 메시지만 전송
	
	// 로그인한 사용자 확인 (선택사항)
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	// 로그아웃 로깅 (보안/감사 목적)
	// 실제 프로덕션에서는 로그아웃 시간을 기록할 수 있음
	_ = userID // 사용자 ID를 로깅에 활용 가능

	responseData := map[string]interface{}{
		"message": "로그아웃이 완료되었습니다",
		"action": "클라이언트에서 토큰을 삭제해주세요",
	}

	return helpers.SuccessWithMessageResponse(c, "로그아웃 완료", responseData)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Get a new JWT token using the current valid token
// @Tags auth
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /refresh [post]
func RefreshToken(c echo.Context) error {
	// 현재 토큰에서 사용자 ID 추출
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	// 새로운 토큰 생성 (utils.GenerateJWT 사용)
	newToken, err := utils.GenerateJWT(userID)
	if err != nil {
		return helpers.InternalServerErrorResponse(c, "토큰 갱신 중 오류가 발생했습니다")
	}

	responseData := map[string]interface{}{
		"token": newToken,
		"message": "토큰이 성공적으로 갱신되었습니다",
	}

	return helpers.SuccessWithMessageResponse(c, "토큰 갱신 완료", responseData)
}

// validateAuthRequest는 인증 요청의 입력값을 검증합니다.
func validateAuthRequest(req types.AuthRequest) error {
	var errors []string

	// 사용자명 검증
	if len(req.Username) == 0 {
		errors = append(errors, "사용자명을 입력해주세요")
	} else if len(req.Username) < 3 {
		errors = append(errors, "사용자명은 최소 3자 이상이어야 합니다")
	} else if len(req.Username) > 50 {
		errors = append(errors, "사용자명은 최대 50자까지 가능합니다")
	}

	// 비밀번호 검증
	if len(req.Password) == 0 {
		errors = append(errors, "비밀번호를 입력해주세요")
	} else if len(req.Password) < 4 {
		errors = append(errors, "비밀번호는 최소 4자 이상이어야 합니다")
	} else if len(req.Password) > 100 {
		errors = append(errors, "비밀번호는 최대 100자까지 가능합니다")
	}

	// 에러가 있으면 첫 번째 에러 반환
	if len(errors) > 0 {
		return fmt.Errorf(errors[0])
	}

	return nil
}

// isUserFriendlyError는 사용자에게 직접 보여줄 수 있는 에러인지 확인합니다.
func isUserFriendlyError(err error) bool {
	userFriendlyMessages := []string{
		"이미 사용 중인 사용자명입니다",
		"사용자명을 입력해주세요",
		"비밀번호는 최소 4자 이상이어야 합니다",
		"사용자명 또는 비밀번호가 올바르지 않습니다",
	}

	errMsg := err.Error()
	for _, msg := range userFriendlyMessages {
		if errMsg == msg {
			return true
		}
	}
	return false
}