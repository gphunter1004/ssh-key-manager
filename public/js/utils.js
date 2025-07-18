// 공통 유틸리티 함수들
window.Utils = {
    // 사용자 친화적 에러 메시지
    FRIENDLY_ERRORS: {
        'connection': '인터넷 연결을 확인해주세요 🌐',
        'timeout': '요청 시간이 초과되었습니다. 다시 시도해주세요 ⏱️',
        'server': '서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요 🔧',
        'auth': '로그인이 필요합니다 🔐',
        'expired': '세션이 만료되었습니다. 다시 로그인해주세요 🕐',
        'invalid': '입력 정보를 확인해주세요 ✏️',
        'notfound': '요청한 정보를 찾을 수 없습니다 🔍',
        'permission': '권한이 없습니다 🚫'
    },

    // 에러 타입 감지
    getErrorType: function(error) {
        const message = error.message?.toLowerCase() || '';
        
        if (message.includes('network') || message.includes('fetch')) return 'connection';
        if (message.includes('timeout')) return 'timeout';
        if (message.includes('500') || message.includes('internal')) return 'server';
        if (message.includes('401') || message.includes('unauthorized')) return 'auth';
        if (message.includes('403') || message.includes('forbidden')) return 'permission';
        if (message.includes('404') || message.includes('not found')) return 'notfound';
        if (message.includes('expired') || message.includes('invalid token')) return 'expired';
        if (message.includes('validation') || message.includes('invalid')) return 'invalid';
        
        return 'unknown';
    },

    // 개선된 에러 처리
    handleError: function(error, context = '') {
        console.error(`[${context}] 에러 발생:`, error);
        
        const errorType = this.getErrorType(error);
        const friendlyMessage = this.FRIENDLY_ERRORS[errorType] || 
                               `알 수 없는 오류가 발생했습니다: ${error.message}`;
        
        // 자동 로그아웃 처리
        if (errorType === 'auth' || errorType === 'expired') {
            setTimeout(() => {
                if (window.AuthManager && AppState.jwtToken) {
                    AuthManager.handleLogout();
                }
            }, 1500);
        }
        
        AppUtils.showError(friendlyMessage);
        return errorType;
    },

    // 토스트 알림 (alert 대신)
    showToast: function(message, type = 'success', duration = 3000) {
        // 기존 토스트 제거
        this.removeToast();
        
        const toast = document.createElement('div');
        toast.id = 'app-toast';
        toast.className = `toast toast-${type}`;
        
        // 아이콘 추가
        const icons = {
            success: '✅',
            error: '❌',
            warning: '⚠️',
            info: 'ℹ️'
        };
        
        toast.innerHTML = `
            <span class="toast-icon">${icons[type] || icons.info}</span>
            <span class="toast-message">${message}</span>
            <button class="toast-close" onclick="Utils.removeToast()">×</button>
        `;
        
        // 스타일 설정
        Object.assign(toast.style, {
            position: 'fixed',
            top: '20px',
            right: '20px',
            padding: '12px 16px',
            borderRadius: '8px',
            color: 'white',
            fontWeight: '500',
            fontSize: '14px',
            zIndex: '10000',
            maxWidth: '350px',
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)',
            transform: 'translateX(100%)',
            transition: 'transform 0.3s ease-in-out, opacity 0.3s',
            opacity: '0',
            display: 'flex',
            alignItems: 'center',
            gap: '8px'
        });
        
        // 타입별 색상
        const colors = {
            success: '#27ae60',
            error: '#e74c3c',
            warning: '#f39c12',
            info: '#3498db'
        };
        toast.style.backgroundColor = colors[type] || colors.info;
        
        document.body.appendChild(toast);
        
        // 애니메이션
        setTimeout(() => {
            toast.style.transform = 'translateX(0)';
            toast.style.opacity = '1';
        }, 10);
        
        // 자동 제거
        setTimeout(() => {
            this.removeToast();
        }, duration);
    },

    removeToast: function() {
        const toast = document.getElementById('app-toast');
        if (toast) {
            toast.style.transform = 'translateX(100%)';
            toast.style.opacity = '0';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        }
    },

    // 로딩 상태 관리
    setLoading: function(isLoading, message = '처리 중...') {
        if (isLoading) {
            document.body.style.cursor = 'wait';
            
            // 모든 버튼 비활성화
            const buttons = document.querySelectorAll('button:not(.toast-close)');
            buttons.forEach(btn => {
                btn.disabled = true;
                btn.dataset.wasDisabled = btn.disabled;
                btn.style.opacity = '0.6';
            });
            
            // 로딩 인디케이터 표시
            this.showLoadingIndicator(message);
        } else {
            document.body.style.cursor = '';
            
            // 버튼 활성화
            const buttons = document.querySelectorAll('button:not(.toast-close)');
            buttons.forEach(btn => {
                if (btn.dataset.wasDisabled !== 'true') {
                    btn.disabled = false;
                }
                btn.style.opacity = '1';
            });
            
            this.hideLoadingIndicator();
        }
    },

    showLoadingIndicator: function(message) {
        this.hideLoadingIndicator(); // 기존 제거
        
        const indicator = document.createElement('div');
        indicator.id = 'loading-indicator';
        indicator.innerHTML = `
            <div class="loading-overlay">
                <div class="loading-spinner"></div>
                <div class="loading-message">${message}</div>
            </div>
        `;
        
        // 스타일
        indicator.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.3);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 9999;
        `;
        
        document.body.appendChild(indicator);
    },

    hideLoadingIndicator: function() {
        const indicator = document.getElementById('loading-indicator');
        if (indicator) {
            indicator.remove();
        }
    },

    // 폼 검증
    validateInput: function(input, rules) {
        const value = input.value.trim();
        let errorMessage = '';
        
        for (const rule of rules) {
            if (!rule.test(value)) {
                errorMessage = rule.message;
                break;
            }
        }
        
        // 에러 메시지 표시/숨김
        let errorEl = input.parentNode.querySelector('.validation-error');
        if (errorMessage) {
            if (!errorEl) {
                errorEl = document.createElement('div');
                errorEl.className = 'validation-error';
                errorEl.style.cssText = 'color: #e74c3c; font-size: 12px; margin-top: 4px;';
                input.parentNode.appendChild(errorEl);
            }
            errorEl.textContent = errorMessage;
            input.classList.add('error');
            return false;
        } else {
            if (errorEl) {
                errorEl.remove();
            }
            input.classList.remove('error');
            return true;
        }
    },

    // 설정 관리
    Settings: {
        get: function(key, defaultValue = null) {
            try {
                const value = localStorage.getItem(`ssh_setting_${key}`);
                return value !== null ? JSON.parse(value) : defaultValue;
            } catch (error) {
                console.warn('설정 로드 실패:', key, error);
                return defaultValue;
            }
        },
        
        set: function(key, value) {
            try {
                localStorage.setItem(`ssh_setting_${key}`, JSON.stringify(value));
                return true;
            } catch (error) {
                console.warn('설정 저장 실패:', key, error);
                return false;
            }
        },
        
        remove: function(key) {
            localStorage.removeItem(`ssh_setting_${key}`);
        }
    },

    // 디바운스 (검색 등에 유용)
    debounce: function(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    },

    // 날짜 포맷팅
    formatDate: function(dateString, options = {}) {
        const defaultOptions = {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            ...options
        };
        
        try {
            return new Date(dateString).toLocaleString('ko-KR', defaultOptions);
        } catch (error) {
            return dateString;
        }
    },

    // 상대 시간 (방금 전, 3분 전 등)
    getRelativeTime: function(dateString) {
        try {
            const date = new Date(dateString);
            const now = new Date();
            const diffMs = now - date;
            const diffSec = Math.floor(diffMs / 1000);
            const diffMin = Math.floor(diffSec / 60);
            const diffHour = Math.floor(diffMin / 60);
            const diffDay = Math.floor(diffHour / 24);
            
            if (diffSec < 60) return '방금 전';
            if (diffMin < 60) return `${diffMin}분 전`;
            if (diffHour < 24) return `${diffHour}시간 전`;
            if (diffDay < 7) return `${diffDay}일 전`;
            if (diffDay < 30) return `${Math.floor(diffDay / 7)}주 전`;
            if (diffDay < 365) return `${Math.floor(diffDay / 30)}개월 전`;
            return `${Math.floor(diffDay / 365)}년 전`;
        } catch (error) {
            return dateString;
        }
    },

    // HTML 이스케이프
    escapeHtml: function(unsafe) {
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }
};