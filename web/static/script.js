// SSH Key Manager - 애플리케이션 시작점 (최종 수정본)
document.addEventListener('DOMContentLoaded', async () => {
    console.log('🚀 SSH Key Manager 시작');
    
    // 전역 상태 초기화
    window.AppState = {
        jwtToken: localStorage.getItem('jwtToken') || null,
        currentUser: null,
        currentView: 'keys'
    };
    
    window.API_BASE_URL = '/api';
    
    // DOM 요소 초기화
    initializeDOM();
    
    // 모든 매니저 초기화 및 이벤트 설정
    initializeManagers();
    
    // 인증 상태 확인 및 UI 업데이트
    await AuthManager.checkLoginStatus();
    
    console.log('✅ 초기화 완료');
});

/**
 * 전역 DOM 요소들을 초기화합니다.
 */
function initializeDOM() {
    window.DOM = {
        // 컨테이너
        container: document.querySelector('.container'),
        authSection: document.getElementById('auth-section'),
        keySection: document.getElementById('key-section'),
        
        // 인증 관련
        loginView: document.getElementById('login-view'),
        registerView: document.getElementById('register-view'),
        loginForm: document.getElementById('login-form'),
        registerForm: document.getElementById('register-form'),
        showRegisterLink: document.getElementById('show-register'),
        showLoginLink: document.getElementById('show-login'),
        
        // 네비게이션
        navKeys: document.getElementById('nav-keys'),
        navUsers: document.getElementById('nav-users'),
        navProfile: document.getElementById('nav-profile'),
        navServers: document.getElementById('nav-servers'),
        navDepartments: document.getElementById('nav-departments'),
        logoutBtn: document.getElementById('logout-btn'),
        
        // 뷰
        keysView: document.getElementById('keys-view'),
        usersView: document.getElementById('users-view'),
        profileView: document.getElementById('profile-view'),
        serversView: document.getElementById('servers-view'),
        departmentsView: document.getElementById('departments-view'),
        
        // 모달 (모든 모달을 등록)
        userDetailModal: document.getElementById('user-detail-modal'),
        deployModal: document.getElementById('deploy-modal'),

        // 키 관리
        keyInfo: document.getElementById('key-info'),
        keyDisplayArea: document.getElementById('key-display-area'),

        // 사용자 관리
        usersList: document.getElementById('users-list'),
        totalUsersSpan: document.getElementById('total-users'),
        usersWithKeysSpan: document.getElementById('users-with-keys'),

        // 프로필
        profileForm: document.getElementById('profile-form'),
        currentUserInfo: document.getElementById('current-user-info'),

        // 기타
        errorDisplay: document.getElementById('error-display'),
    };
    
    console.log('✅ DOM 요소 초기화 완료');
}

/**
 * 모든 매니저를 순차적으로 초기화합니다.
 */
function initializeManagers() {
    const managers = [
        // 의존성 없는 유틸리티 먼저 초기화
        { name: 'Utils', manager: Utils },
        { name: 'ModalManager', manager: ModalManager },
        { name: 'CopyManager', manager: CopyManager },
        
        // 나머지 매니저들 초기화
        { name: 'ViewManager', manager: ViewManager },
        { name: 'AuthManager', manager: AuthManager },
        { name: 'KeyManager', manager: KeyManager },
        { name: 'UserManager', manager: UserManager },
        { name: 'ProfileManager', manager: ProfileManager }
    ];
    
    managers.forEach(({ name, manager }) => {
        try {
            if (manager && typeof manager.init === 'function') {
                manager.init();
            }
            console.log(`✅ ${name} 초기화 완료`);
        } catch (error) {
            console.error(`⚠️ ${name} 초기화 실패:`, error);
        }
    });
}

// API 통신을 위한 래퍼 함수
window.AppUtils = {
    apiFetch: async function(endpoint, method = 'GET', body = null) {
        const headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
        
        if (AppState.jwtToken) {
            headers['Authorization'] = `Bearer ${AppState.jwtToken}`;
        }
        
        const options = { method, headers };
        if (body) {
            options.body = JSON.stringify(body);
        }
        
        try {
            const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
            const result = await response.json();
            
            if (!response.ok) {
                const errorMessage = result.error?.message || result.message || `HTTP 에러: ${response.status}`;
                const error = new Error(errorMessage);
                error.status = response.status;
                error.data = result;
                throw error;
            }
            
            return result.data || result;
            
        } catch (error) {
            console.error(`❌ API 오류 [${method} ${endpoint}]:`, error);
            Utils.showToast(error.message, 'error');
            throw error;
        }
    }
};