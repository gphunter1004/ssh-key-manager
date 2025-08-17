package utils

import (
	"fmt"
	"log"
	"time"
)

// LogOperation 작업 로그 (시작/완료 표준화)
func LogOperation(operation string, fn func() error) error {
	log.Printf("🚀 %s 시작", operation)
	start := time.Now()

	err := fn()
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ %s 실패 (%v): %v", operation, duration, err)
		return err
	}

	log.Printf("✅ %s 완료 (%v)", operation, duration)
	return nil
}

// LogUserAction 사용자 행동 로그 표준화
func LogUserAction(userID uint, action, resource string, success bool, details ...string) {
	icon := "✅"
	status := "성공"

	if !success {
		icon = "❌"
		status = "실패"
	}

	detail := ""
	if len(details) > 0 {
		detail = fmt.Sprintf(" - %s", details[0])
	}

	log.Printf("%s 사용자 행동: ID=%d, 행동=%s, 리소스=%s, 상태=%s%s",
		icon, userID, action, resource, status, detail)
}

// LogServiceCall 서비스 호출 로그 표준화
func LogServiceCall(serviceName, methodName string, userID uint, params ...interface{}) {
	paramStr := ""
	if len(params) > 0 {
		paramStr = fmt.Sprintf(" 파라미터=%+v", params)
	}

	log.Printf("🔧 서비스 호출: %s.%s (사용자 ID: %d)%s",
		serviceName, methodName, userID, paramStr)
}

// LogAPIRequest API 요청 로그 표준화
func LogAPIRequest(method, path string, userID uint, success bool, duration time.Duration) {
	icon := "📡"
	if !success {
		icon = "📛"
	}

	log.Printf("%s API 요청: %s %s (사용자 ID: %d) - %v",
		icon, method, path, userID, duration)
}

// LogDatabaseOperation 데이터베이스 작업 로그
func LogDatabaseOperation(operation, table string, recordID uint, success bool, err error) {
	icon := "💾"
	status := "성공"

	if !success {
		icon = "💥"
		status = "실패"
	}

	errDetail := ""
	if err != nil {
		errDetail = fmt.Sprintf(" - 오류: %v", err)
	}

	log.Printf("%s DB 작업: %s %s (ID: %d) - %s%s",
		icon, operation, table, recordID, status, errDetail)
}

// LogServerOperation 서버 작업 로그
func LogServerOperation(operation, serverInfo string, success bool, duration time.Duration, err error) {
	icon := "🖥️"
	status := "성공"

	if !success {
		icon = "💻"
		status = "실패"
	}

	errDetail := ""
	if err != nil {
		errDetail = fmt.Sprintf(" - 오류: %v", err)
	}

	log.Printf("%s 서버 작업: %s %s (%v) - %s%s",
		icon, operation, serverInfo, duration, status, errDetail)
}

// LogKeyOperation SSH 키 작업 로그
func LogKeyOperation(operation string, userID uint, keyType string, success bool, details string) {
	icon := "🔑"
	status := "성공"

	if !success {
		icon = "🔒"
		status = "실패"
	}

	log.Printf("%s SSH 키 작업: %s (사용자 ID: %d, 타입: %s) - %s - %s",
		icon, operation, userID, keyType, status, details)
}

// LogConfigOperation 설정 작업 로그
func LogConfigOperation(operation string, config string, success bool) {
	icon := "⚙️"
	status := "성공"

	if !success {
		icon = "🔧"
		status = "실패"
	}

	log.Printf("%s 설정 작업: %s (%s) - %s",
		icon, operation, config, status)
}

// LogSecurityEvent 보안 이벤트 로그
func LogSecurityEvent(eventType string, userID uint, details string, severity string) {
	var icon string
	switch severity {
	case "high":
		icon = "🚨"
	case "medium":
		icon = "⚠️"
	case "low":
		icon = "ℹ️"
	default:
		icon = "🔐"
	}

	log.Printf("%s 보안 이벤트 [%s]: %s (사용자 ID: %d) - %s",
		icon, severity, eventType, userID, details)
}
