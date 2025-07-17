// 메인 애플리케이션 스크립트
document.addEventListener('DOMContentLoaded', async () => {
    console.log('🚀 SSH Key Manager 애플리케이션 시작');
    
    // 전역 상태 관리
    window.AppState = {
        jwtToken: localStorage.getItem('jwtToken') || null,
        currentUser: null,
        currentView: 'keys'
    };

    // API 기본 설정
    window.API_BASE_URL = '/api';

    // DOM 요소 초기화
    initializeDOMElements();
    
    // 각 관리자 초기화
    await initializeManagers();
    
    // 이벤트 리스너 설정
    setupEventListeners();
    
    // 자동 로그인 확인
    await checkAutoLogin();
    
    // 초기 UI 업데이트
    updateUI();
    
    console.log('✅ 애플리케이션 초기화 완료');
});

function initializeDOMElements() {
    console.log('📋 DOM 요소 초기화 중...');
    
    // 컨테이너 및 섹션 요소들
    window.DOM = {
        container: document.querySelector('.container'),
        authSection: document.getElementById('auth-section'),
        keySection: document.getElementById('key-section'),
        loginView: document.getElementById('login-view'),
        registerView: document.getElementById('register-view'),
        errorDisplay: document.getElementById('error-display'),
        
        // 네비게이션 관련
        keysView: document.getElementById('keys-view'),
        usersView: document.getElementById('users-view'),
        profileView: document.getElementById('profile-view'),
        navKeys: document.getElementById('nav-keys'),
        navUsers: document.getElementById('nav-users'),
        navProfile: document.getElementById('nav-profile'),
        
        // 키 관련 요소들
        keyDisplayArea: document.getElementById('key-display-area'),
        keyInfo: document.getElementById('key-info'),
        keyPublicPre: document.getElementById('key-public'),
        keyPemPre: document.getElementById('key-pem'),
        keyPpkPre: document.getElementById('key-ppk'),
        cmdPublicPre: document.getElementById('cmd-public'),
        cmdAuthorizedKeysPre: document.getElementById('cmd-authorized-keys'),
        cmdPemPre: document.getElementById('cmd-pem'),
        cmdPpkPre: document.getElementById('cmd-ppk'),
        
        // 사용자 관련 요소들
        usersList: document.getElementById('users-list'),
        totalUsersSpan: document.getElementById('total-users'),
        usersWithKeysSpan: document.getElementById('users-with-keys'),
        
        // 프로필 관련 요소들
        profileForm: document.getElementById('profile-form'),
        currentUserInfo: document.getElementById('current-user-info'),
        
        // 모달 관련 요소들
        userDetailModal: document.getElementById('user-detail-modal'),
        userDetailContent: document.getElementById('user-detail-content'),
        closeModalBtn: document.querySelector('.close'),
        
        // 인증 관련
        showRegisterLink: document.getElementById('show-register'),
        showLoginLink: document.getElementById('show-login'),
        loginForm: document.getElementById('login-form'),
        registerForm: document.getElementById('register-form'),
        logoutBtn: document.getElementById('logout-btn')
    };
    
    console.log('✅ DOM 요소 초기화 완료');
}

async function initializeManagers() {
    console.log('🔧 관리자 모듈 초기화 중...');
    
    try {
        // ViewManager 초기화
        ViewManager.init();
        
        // ModalManager 초기화
        ModalManager.init();
        
        // CopyManager 초기화
        CopyManager.init();
        
        console.log('✅ 모든 관리자 초기화 완료');
    } catch (error) {
        console.error('❌ 관리자 초기화 실패:', error);
    }
}

function setupEventListeners() {
    console.log('🎯 이벤트 리스너 설정 중...');
    
    // 네비게이션 이벤트
    DOM.navKeys.addEventListener('click', () => ViewManager.showView('keys'));
    DOM.navUsers.addEventListener('click', () => ViewManager.showView('users'));
    DOM.navProfile.addEventListener('click', () => ViewManager.showView('profile'));

    // 각 관리자의 이벤트 리스너 설정
    AuthManager.setupEventListeners();
    KeyManager.setupEventListeners();
    UserManager.setupEventListeners();
    ProfileManager.setupEventListeners();
    ModalManager.setupEventListeners();
    CopyManager.setupEventListeners();
    
    // 전역 키보드 이벤트
    setupGlobalKeyboardEvents();
    
    console.log('✅ 이벤트 리스너 설정 완료');
}

function setupGlobalKeyboardEvents() {
    document.addEventListener('keydown', (e) => {
        // Ctrl + K: 키 생성 단축키
        if (e.ctrlKey && e.key === 'k' && AppState.jwtToken) {
            e.preventDefault();
            ViewManager.showView('keys');
            setTimeout(() => KeyManager.createKey(), 100);
        }
        
        // Ctrl + U: 사용자 목록 단축키
        if (e.ctrlKey && e.key === 'u' && AppState.jwtToken) {
            e.preventDefault();
            ViewManager.showView('users');
        }
        
        // Ctrl + P: 프로필 단축키
        if (e.ctrlKey && e.key === 'p' && AppState.jwtToken) {
            e.preventDefault();
            ViewManager.showView('profile');
        }
    });
}

async function checkAutoLogin() {
    console.log('🔐 자동 로그인 확인 중...');
    
    if (AppState.jwtToken) {
        const isValid = await AuthManager.validateToken();
        if (isValid) {
            console.log('✅ 자동 로그인 성공');
        } else {
            console.log('❌ 토큰이 무효하여 로그아웃 처리됨');
        }
    } else {
        console.log('📝 저장된 토큰이 없음');
    }
}

function updateUI() {
    console.log('🎨 UI 업데이트 중...');
    
    if (AppState.jwtToken) {
        DOM.authSection.classList.add('hidden');
        DOM.keySection.classList.remove('hidden');
        DOM.container.classList.add('container-wide');
        ViewManager.showView(AppState.currentView);
    } else {
        DOM.authSection.classList.remove('hidden');
        DOM.keySection.classList.add('hidden');
        DOM.container.classList.remove('container-wide');
        DOM.registerView.classList.add('hidden');
        DOM.loginView.classList.remove('hidden');
    }
}

// 전역 유틸리티 함수들
window.AppUtils = {
    showError: function(message) {
        DOM.errorDisplay.textContent = message;
        DOM.errorDisplay.style.display = 'block';
        console.error('앱 에러:', message);
        
        // 에러 자동 숨김 (10초 후)
        setTimeout(() => {
            AppUtils.clearError();
        }, 10000);
    },
    
    clearError: function() {
        DOM.errorDisplay.textContent = '';
        DOM.errorDisplay.style.display = 'none';
    },
    
    apiFetch: async function(endpoint, method = 'GET', body = null) {
        AppUtils.clearError();
        
        const headers = { 
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
        
        if (AppState.jwtToken) {
            headers['Authorization'] = `Bearer ${AppState.jwtToken}`;
        }
        
        const options = { 
            method, 
            headers,
            credentials: 'same-origin' // CSRF 보호
        };
        
        if (body) {
            options.body = JSON.stringify(body);
        }
        
        try {
            console.log(`🌐 API 요청: ${method} ${endpoint}`);
            
            const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
            const data = await response.json();
            
            if (!response.ok) {
                // 401 Unauthorized인 경우 자동 로그아웃
                if (response.status === 401 && AppState.jwtToken) {
                    console.warn('토큰이 만료되어 자동 로그아웃됩니다.');
                    AuthManager.handleLogout();
                    throw new Error('세션이 만료되었습니다. 다시 로그인해주세요.');
                }
                
                throw new Error(data.error || data.message || `HTTP ${response.status}: 요청이 실패했습니다.`);
            }
            
            console.log(`✅ API 응답: ${method} ${endpoint} - 성공`);
            return data;
            
        } catch (error) {
            console.error(`❌ API 오류: ${method} ${endpoint}`, error);
            AppUtils.showError(error.message);
            throw error;
        }
    },
    
    // 날짜 포맷팅 유틸리티
    formatDate: function(dateString, options = {}) {
        const defaultOptions = {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            ...options
        };
        
        return new Date(dateString).toLocaleString('ko-KR', defaultOptions);
    },
    
    // 상대 시간 계산
    getRelativeTime: function(dateString) {
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
    },
    
    // 로딩 상태 관리
    setGlobalLoading: function(isLoading, message = '처리 중...') {
        const loadingEl = document.getElementById('global-loading');
        
        if (isLoading) {
            if (!loadingEl) {
                const loading = document.createElement('div');
                loading.id = 'global-loading';
                loading.innerHTML = `
                    <div class="loading-overlay">
                        <div class="loading-spinner"></div>
                        <div class="loading-message">${message}</div>
                    </div>
                `;
                document.body.appendChild(loading);
            }
        } else {
            if (loadingEl) {
                loadingEl.remove();
            }
        }
    },
    
    // 디바운스 유틸리티
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
    
    // 쓰로틀 유틸리티
    throttle: function(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }
};

// 전역 함수로 노출
window.updateUI = updateUI;

// 애플리케이션 상태 변경 감지
window.addEventListener('beforeunload', () => {
    console.log('🔄 애플리케이션 종료 중...');
    
    // 필요한 정리 작업 수행
    if (ModalManager.isOpen) {
        ModalManager.closeModal();
    }
});

// 네트워크 상태 모니터링
window.addEventListener('online', () => {
    console.log('🌐 네트워크 연결됨');
    AppUtils.clearError();
});

window.addEventListener('offline', () => {
    console.log('📡 네트워크 연결 끊김');
    AppUtils.showError('네트워크 연결이 끊어졌습니다. 인터넷 연결을 확인해주세요.');
});

// 개발자 도구용 전역 함수들 (프로덕션에서는 제거)
if (typeof window !== 'undefined' && window.location.hostname === 'localhost') {
    window.debug = {
        getState: () => AppState,
        getDOM: () => DOM,
        clearStorage: () => {
            localStorage.clear();
            sessionStorage.clear();
            console.log('스토리지 클리어됨');
        },
        testAPI: async (endpoint) => {
            try {
                const result = await AppUtils.apiFetch(endpoint);
                console.log('API 테스트 결과:', result);
                return result;
            } catch (error) {
                console.error('API 테스트 실패:', error);
            }
        }
    };
    
    console.log('🛠️ 개발자 모드: window.debug 객체 사용 가능');
}// 메인 애플리케이션 스크립트
document.addEventListener('DOMContentLoaded', () => {
    // 전역 상태 관리
    window.AppState = {
        jwtToken: localStorage.getItem('jwtToken') || null,
        currentUser: null,
        currentView: 'keys'
    };

    // API 기본 설정
    window.API_BASE_URL = '/api';

    // DOM 요소 초기화
    initializeDOMElements();
    
    // 이벤트 리스너 설정
    setupEventListeners();
    
    // 초기 UI 업데이트
    updateUI();
});

function initializeDOMElements() {
    // 컨테이너 및 섹션 요소들
    window.DOM = {
        container: document.querySelector('.container'),
        authSection: document.getElementById('auth-section'),
        keySection: document.getElementById('key-section'),
        loginView: document.getElementById('login-view'),
        registerView: document.getElementById('register-view'),
        errorDisplay: document.getElementById('error-display'),
        
        // 네비게이션 관련
        keysView: document.getElementById('keys-view'),
        usersView: document.getElementById('users-view'),
        profileView: document.getElementById('profile-view'),
        navKeys: document.getElementById('nav-keys'),
        navUsers: document.getElementById('nav-users'),
        navProfile: document.getElementById('nav-profile'),
        
        // 키 관련 요소들
        keyDisplayArea: document.getElementById('key-display-area'),
        keyInfo: document.getElementById('key-info'),
        keyPublicPre: document.getElementById('key-public'),
        keyPemPre: document.getElementById('key-pem'),
        keyPpkPre: document.getElementById('key-ppk'),
        cmdPublicPre: document.getElementById('cmd-public'),
        cmdAuthorizedKeysPre: document.getElementById('cmd-authorized-keys'),
        cmdPemPre: document.getElementById('cmd-pem'),
        cmdPpkPre: document.getElementById('cmd-ppk'),
        
        // 사용자 관련 요소들
        usersList: document.getElementById('users-list'),
        totalUsersSpan: document.getElementById('total-users'),
        usersWithKeysSpan: document.getElementById('users-with-keys'),
        
        // 프로필 관련 요소들
        profileForm: document.getElementById('profile-form'),
        currentUserInfo: document.getElementById('current-user-info'),
        
        // 모달 관련 요소들
        userDetailModal: document.getElementById('user-detail-modal'),
        userDetailContent: document.getElementById('user-detail-content'),
        closeModalBtn: document.querySelector('.close'),
        
        // 인증 관련
        showRegisterLink: document.getElementById('show-register'),
        showLoginLink: document.getElementById('show-login'),
        loginForm: document.getElementById('login-form'),
        registerForm: document.getElementById('register-form'),
        logoutBtn: document.getElementById('logout-btn')
    };
}

function setupEventListeners() {
    // 네비게이션 이벤트
    DOM.navKeys.addEventListener('click', () => ViewManager.showView('keys'));
    DOM.navUsers.addEventListener('click', () => ViewManager.showView('users'));
    DOM.navProfile.addEventListener('click', () => ViewManager.showView('profile'));

    // 인증 관련 이벤트
    AuthManager.setupEventListeners();
    
    // 키 관리 이벤트
    KeyManager.setupEventListeners();
    
    // 사용자 관리 이벤트
    UserManager.setupEventListeners();
    
    // 프로필 관리 이벤트
    ProfileManager.setupEventListeners();
    
    // 모달 이벤트
    ModalManager.setupEventListeners();
    
    // 복사 기능 이벤트
    CopyManager.setupEventListeners();
}

function updateUI() {
    if (AppState.jwtToken) {
        DOM.authSection.classList.add('hidden');
        DOM.keySection.classList.remove('hidden');
        DOM.container.classList.add('container-wide');
        ViewManager.showView(AppState.currentView);
    } else {
        DOM.authSection.classList.remove('hidden');
        DOM.keySection.classList.add('hidden');
        DOM.container.classList.remove('container-wide');
        DOM.registerView.classList.add('hidden');
        DOM.loginView.classList.remove('hidden');
    }
}

// 전역 유틸리티 함수들
window.AppUtils = {
    showError: function(message) {
        DOM.errorDisplay.textContent = message;
        DOM.errorDisplay.style.display = 'block';
    },
    
    clearError: function() {
        DOM.errorDisplay.textContent = '';
        DOM.errorDisplay.style.display = 'none';
    },
    
    apiFetch: async function(endpoint, method = 'GET', body = null) {
        AppUtils.clearError();
        const headers = { 'Content-Type': 'application/json' };
        if (AppState.jwtToken) {
            headers['Authorization'] = `Bearer ${AppState.jwtToken}`;
        }
        const options = { method, headers };
        if (body) {
            options.body = JSON.stringify(body);
        }
        
        try {
            const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || data.message || 'An unknown error occurred.');
            }
            return data;
        } catch (error) {
            AppUtils.showError(error.message);
            throw error;
        }
    }
};

// 전역 함수로 노출
window.updateUI = updateUI;