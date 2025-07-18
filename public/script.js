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
    
    console.log('✅ 이벤트 리스너 설정 완료');
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
                    setTimeout(() => AuthManager.handleLogout(), 2000);
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
    
    if (AppState.jwtToken) {
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

// 개선된 전역 유틸리티 함수들
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
            const errorType = Utils.handleError(error, `API ${method} ${endpoint}`);
            
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
    
    Utils.removeToast();
    Utils.hideLoadingIndicator();
});