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
    const domInitialized = initializeDOMElements();
    if (!domInitialized) {
        console.error('❌ DOM 초기화 실패, 애플리케이션을 시작할 수 없습니다');
        return;
    }
    
    // 매니저들 초기화
    const managersInitialized = initializeManagers();
    if (!managersInitialized) {
        console.error('❌ 매니저 초기화 실패');
        return;
    }
    
    // 이벤트 리스너 설정
    setupEventListeners();
    
    // 자동 로그인 확인
    await checkAutoLogin();
    
    // 초기 UI 업데이트
    updateUI();
    
    // 자동 토큰 갱신 설정
    setupTokenRefresh();
    
    // 오프라인 감지 설정
    setupOfflineDetection();
    
    console.log('✅ 애플리케이션 초기화 완료');
});

function initializeDOMElements() {
    console.log('📋 DOM 요소 초기화 중...');
    
    try {
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
        
        // 필수 요소들 확인
        const criticalElements = [
            'container', 'authSection', 'keySection', 'loginView', 'registerView',
            'navKeys', 'navUsers', 'navProfile', 'loginForm', 'registerForm', 'logoutBtn'
        ];
        
        const missingCritical = criticalElements.filter(key => !DOM[key]);
        
        if (missingCritical.length > 0) {
            console.error('❌ 필수 DOM 요소가 누락됨:', missingCritical);
            return false;
        }
        
        // 키 관련 요소들 확인 (경고만 출력)
        const keyElements = [
            'keyDisplayArea', 'keyInfo', 'keyPublicPre', 'keyPemPre', 'keyPpkPre',
            'cmdPublicPre', 'cmdAuthorizedKeysPre', 'cmdPemPre', 'cmdPpkPre'
        ];
        
        const missingKeyElements = keyElements.filter(key => !DOM[key]);
        if (missingKeyElements.length > 0) {
            console.warn('⚠️ 키 관련 DOM 요소가 누락됨:', missingKeyElements);
        }
        
        console.log('✅ DOM 요소 초기화 완료');
        return true;
        
    } catch (error) {
        console.error('❌ DOM 초기화 중 오류 발생:', error);
        return false;
    }
}

function initializeManagers() {
    console.log('🔧 매니저들 초기화 중...');
    
    try {
        // Utils 초기화
        if (typeof Utils !== 'undefined' && Utils.init) {
            Utils.init();
        }
        
        // CopyManager 초기화
        if (typeof CopyManager !== 'undefined' && CopyManager.init) {
            CopyManager.init();
        }
        
        // KeyManager 초기화 (DOM 요소 체크 포함)
        if (typeof KeyManager !== 'undefined') {
            const keyManagerInitialized = KeyManager.init();
            if (!keyManagerInitialized) {
                console.error('❌ KeyManager 초기화 실패');
                // KeyManager는 치명적이지 않으므로 계속 진행
            }
        }
        
        // ModalManager 초기화
        if (typeof ModalManager !== 'undefined' && ModalManager.init) {
            ModalManager.init();
        }
        
        // ViewManager 초기화
        if (typeof ViewManager !== 'undefined' && ViewManager.init) {
            ViewManager.init();
        }
        
        console.log('✅ 매니저들 초기화 완료');
        return true;
        
    } catch (error) {
        console.error('❌ 매니저 초기화 중 오류 발생:', error);
        return false;
    }
}

function setupEventListeners() {
    console.log('🎯 이벤트 리스너 설정 중...');
    
    try {
        // 네비게이션 이벤트 (null 체크 추가)
        if (DOM.navKeys) {
            DOM.navKeys.addEventListener('click', () => ViewManager?.showView('keys'));
        }
        if (DOM.navUsers) {
            DOM.navUsers.addEventListener('click', () => ViewManager?.showView('users'));
        }
        if (DOM.navProfile) {
            DOM.navProfile.addEventListener('click', () => ViewManager?.showView('profile'));
        }

        // 각 관리자의 이벤트 리스너 설정 (존재 확인 후)
        if (typeof AuthManager !== 'undefined') {
            AuthManager.setupEventListeners();
        }
        if (typeof UserManager !== 'undefined') {
            UserManager.setupEventListeners();
        }
        if (typeof ProfileManager !== 'undefined') {
            ProfileManager.setupEventListeners();
        }
        if (typeof ModalManager !== 'undefined') {
            ModalManager.setupEventListeners();
        }
        if (typeof CopyManager !== 'undefined') {
            CopyManager.setupEventListeners();
        }
        
        console.log('✅ 이벤트 리스너 설정 완료');
        
    } catch (error) {
        console.error('❌ 이벤트 리스너 설정 중 오류 발생:', error);
    }
}

// 자동 토큰 갱신
function setupTokenRefresh() {
    setInterval(async () => {
        if (AppState.jwtToken) {
            try {
                await AppUtils.apiFetch('/users/me');
                console.log('🔄 토큰 유효성 확인 완료');
            } catch (error) {
                if (error.message.includes('401') || error.message.includes('expired')) {
                    console.warn('🔐 토큰이 만료되어 자동 로그아웃됩니다');
                    Utils.showToast('세션이 만료되어 로그아웃됩니다', 'warning');
                    setTimeout(() => AuthManager?.handleLogout(), 2000);
                }
            }
        }
    }, 5 * 60 * 1000); // 5분마다 확인
    
    console.log('🔄 자동 토큰 갱신 설정 완료 (5분 간격)');
}

// 오프라인 감지
function setupOfflineDetection() {
    window.addEventListener('offline', () => {
        console.log('📡 네트워크 연결 끊김');
        Utils.showToast('인터넷 연결이 끊어졌습니다', 'warning', 5000);
    });

    window.addEventListener('online', () => {
        console.log('🌐 네트워크 연결됨');
        Utils.showToast('인터넷 연결이 복원되었습니다', 'success');
        AppUtils.clearError();
    });
    
    console.log('📡 오프라인 감지 설정 완료');
}

async function checkAutoLogin() {
    console.log('🔐 자동 로그인 확인 중...');
    
    if (AppState.jwtToken && typeof AuthManager !== 'undefined') {
        const isValid = await AuthManager.validateToken();
        if (isValid) {
            console.log('✅ 자동 로그인 성공');
            Utils.showToast(`안녕하세요, ${AppState.currentUser?.username || '사용자'}님!`, 'success');
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
        DOM.authSection?.classList.add('hidden');
        DOM.keySection?.classList.remove('hidden');
        DOM.container?.classList.add('container-wide');
        ViewManager?.showView(AppState.currentView);
    } else {
        DOM.authSection?.classList.remove('hidden');
        DOM.keySection?.classList.add('hidden');
        DOM.container?.classList.remove('container-wide');
        DOM.registerView?.classList.add('hidden');
        DOM.loginView?.classList.remove('hidden');
    }
}

// 개선된 전역 유틸리티 함수들
window.AppUtils = {
    showError: function(message) {
        if (DOM.errorDisplay) {
            DOM.errorDisplay.textContent = message;
            DOM.errorDisplay.style.display = 'block';
        }
        console.error('앱 에러:', message);
        
        // 에러 자동 숨김 (10초 후)
        setTimeout(() => {
            AppUtils.clearError();
        }, 10000);
    },
    
    clearError: function() {
        if (DOM.errorDisplay) {
            DOM.errorDisplay.textContent = '';
            DOM.errorDisplay.style.display = 'none';
        }
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
            credentials: 'same-origin',
            timeout: 30000 // 30초 타임아웃
        };
        
        if (body) {
            options.body = JSON.stringify(body);
        }
        
        try {
            console.log(`🌐 API 요청: ${method} ${endpoint}`);
            
            const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
            const data = await response.json();
            
            if (!response.ok) {
                const error = new Error(data.error || data.message || `HTTP ${response.status}`);
                error.status = response.status;
                throw error;
            }
            
            console.log(`✅ API 응답: ${method} ${endpoint} - 성공`);
            return data;
            
        } catch (error) {
            console.error(`❌ API 오류: ${method} ${endpoint}`, error);
            
            // 개선된 에러 처리
            const errorType = Utils?.handleError(error, `API ${method} ${endpoint}`);
            
            // 에러 타입에 따른 추가 처리
            if (errorType === 'auth' || errorType === 'expired') {
                // 인증 에러는 Utils.handleError에서 처리됨
            }
            
            throw error;
        }
    }
};

// 전역 함수로 노출
window.updateUI = updateUI;

// 애플리케이션 종료 시 정리
window.addEventListener('beforeunload', () => {
    console.log('🔄 애플리케이션 종료 중...');
    
    // 정리 작업
    if (ModalManager && ModalManager.isOpen) {
        ModalManager.closeModal();
    }
    
    if (Utils) {
        Utils.removeToast();
        Utils.hideLoadingIndicator();
    }
});

// 전역 에러 핸들러
window.addEventListener('error', (event) => {
    console.error('🚨 전역 에러 발생:', event.error);
    if (event.error && event.error.message) {
        Utils?.showToast('예상치 못한 오류가 발생했습니다', 'error');
    }
});

// 디버깅용 전역 객체 노출 (개발 환경에서만)
if (location.hostname === 'localhost' || location.hostname === '127.0.0.1') {
    window.DEBUG = {
        AppState,
        DOM,
        AuthManager: typeof AuthManager !== 'undefined' ? AuthManager : null,
        KeyManager: typeof KeyManager !== 'undefined' ? KeyManager : null,
        UserManager: typeof UserManager !== 'undefined' ? UserManager : null,
        ProfileManager: typeof ProfileManager !== 'undefined' ? ProfileManager : null,
        ModalManager: typeof ModalManager !== 'undefined' ? ModalManager : null,
        ViewManager: typeof ViewManager !== 'undefined' ? ViewManager : null,
        CopyManager: typeof CopyManager !== 'undefined' ? CopyManager : null,
        Utils: typeof Utils !== 'undefined' ? Utils : null
    };
    console.log('🔧 디버깅 모드: window.DEBUG 객체를 통해 모든 매니저에 접근 가능');
}