// SSH Key Manager - 애플리케이션 시작점 (최소화된 부트스트랩)
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
        
        // 기타
        errorDisplay: document.getElementById('error-display'),
        usersList: document.getElementById('users-list'),
        currentUserInfo: document.getElementById('current-user-info'),
        userDetailModal: document.getElementById('user-detail-modal'),
        userDetailContent: document.getElementById('user-detail-content'),
        totalUsersSpan: document.getElementById('total-users'),
        usersWithKeysSpan: document.getElementById('users-with-keys'),
        profileForm: document.getElementById('profile-form'),
        closeModalBtn: document.querySelector('.close')
    };
    
    console.log('✅ DOM 요소 초기화 완료');
}

/**
 * 모든 매니저를 순차적으로 초기화합니다.
 */
function initializeManagers() {
    const managers = [
        { name: 'Utils', manager: Utils },
        { name: 'CopyManager', manager: CopyManager },
        { name: 'ModalManager', manager: ModalManager },
        { name: 'ViewManager', manager: ViewManager },
        { name: 'AuthManager', manager: AuthManager },
        { name: 'KeyManager', manager: KeyManager },
        { name: 'UserManager', manager: UserManager },
        { name: 'ProfileManager', manager: ProfileManager }
    ];
    
    managers.forEach(({ name, manager }) => {
        try {
            if (manager && manager.init) {
                manager.init();
            } else if (manager && manager.setupEventListeners) {
                // init이 없는 경우 eventListeners만 설정
                manager.setupEventListeners();
            }
            console.log(`✅ ${name} 초기화 완료`);
        } catch (error) {
            console.warn(`⚠️ ${name} 초기화 실패:`, error.message);
        }
    });
}

// AppUtils는 유틸리티 파일에 정의되어야 하지만, 현재 구조를 위해 여기에 정의
// 실제 운영 환경에서는 utils.js로 이동하여 모듈화하는 것이 좋습니다.
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
        if (body) options.body = JSON.stringify(body);
        
        try {
            console.log(`🌐 API 요청: ${method} ${endpoint}`);
            
            const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
            const result = await response.json();
            
            console.log(`📦 API 응답 [${method} ${endpoint}]:`, result);
            
            if (!response.ok) {
                const error = new Error(result.error?.message || result.message || `HTTP ${response.status}`);
                error.status = response.status;
                error.data = result;
                throw error;
            }
            
            return result.data || result;
            
        } catch (error) {
            console.error(`❌ API 오류: ${method} ${endpoint}`, error);
            
            if (typeof Utils !== 'undefined' && Utils.handleError) {
                Utils.handleError(error, `API ${method} ${endpoint}`);
            }
            
            throw error;
        }
    },
    clearError: function() {
        if (DOM.errorDisplay) {
            DOM.errorDisplay.textContent = '';
            DOM.errorDisplay.style.display = 'none';
        }
    }
};