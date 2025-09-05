// SSH Key Manager - 최종 완성본 (AuthManager 연동)
document.addEventListener('DOMContentLoaded', async () => {
    console.log('🚀 SSH Key Manager 시작');
    
    // 전역 상태
    window.AppState = {
        jwtToken: localStorage.getItem('jwtToken') || null,
        currentUser: null,
        currentView: 'keys'
    };
    
    window.API_BASE_URL = '/api';
    
    // DOM 요소 초기화
    initializeDOM();
    
    // 매니저들 초기화
    initializeManagers();
    
    // 이벤트 설정
    setupEvents();
    
    // 자동 로그인 확인
    await checkAutoLogin();
    
    // UI 업데이트
    updateUI();
    
    console.log('✅ 초기화 완료');
});

// DOM 요소 초기화
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

// 매니저 초기화
function initializeManagers() {
    const managers = [
        { name: 'Utils', manager: Utils },
        { name: 'CopyManager', manager: CopyManager },
        { name: 'KeyManager', manager: KeyManager },
        { name: 'ModalManager', manager: ModalManager },
        { name: 'ViewManager', manager: ViewManager },
        { name: 'AuthManager', manager: AuthManager }
    ];
    
    managers.forEach(({ name, manager }) => {
        try {
            if (manager && manager.init) {
                manager.init();
                console.log(`✅ ${name} 초기화 완료`);
            }
        } catch (error) {
            console.warn(`⚠️ ${name} 초기화 실패:`, error.message);
        }
    });
}

// 이벤트 설정
function setupEvents() {
    console.log('🎯 이벤트 설정 시작');
    
    // 네비게이션 이벤트
    setupNavigationEvents();
    
    // 매니저 이벤트 설정 (AuthManager가 인증 이벤트 담당)
    const managers = [
        { name: 'AuthManager', manager: AuthManager },
        { name: 'UserManager', manager: UserManager },
        { name: 'ProfileManager', manager: ProfileManager },
        { name: 'ModalManager', manager: ModalManager },
        { name: 'CopyManager', manager: CopyManager }
    ];
    
    managers.forEach(({ name, manager }) => {
        try {
            if (manager && manager.setupEventListeners) {
                manager.setupEventListeners();
                console.log(`✅ ${name} 이벤트 설정 완료`);
            }
        } catch (error) {
            console.warn(`⚠️ ${name} 이벤트 설정 실패:`, error.message);
        }
    });
    
    console.log('✅ 모든 이벤트 설정 완료');
}

// 네비게이션 이벤트 설정
function setupNavigationEvents() {
    const navButtons = [
        { element: DOM.navKeys, view: 'keys', name: '키 관리' },
        { element: DOM.navUsers, view: 'users', name: '사용자 목록' },
        { element: DOM.navProfile, view: 'profile', name: '프로필' },
        { element: DOM.navServers, view: 'servers', name: '서버 관리' },
        { element: DOM.navDepartments, view: 'departments', name: '부서 관리' }
    ];
    
    navButtons.forEach(({ element, view, name }) => {
        if (element) {
            element.onclick = function(e) {
                e.preventDefault();
                console.log(`🔘 ${name} 버튼 클릭`);
                showView(view);
            };
            console.log(`✅ ${name} 이벤트 설정`);
        }
    });
    
    // 로그아웃 버튼
    if (DOM.logoutBtn) {
        DOM.logoutBtn.onclick = function(e) {
            e.preventDefault();
            console.log('🚪 로그아웃 버튼 클릭');
            if (AuthManager && AuthManager.handleLogout) {
                AuthManager.handleLogout();
            } else {
                handleLogout();
            }
        };
        console.log('✅ 로그아웃 버튼 이벤트 설정');
    }
}

// 뷰 전환
function showView(viewName) {
    console.log('🔄 뷰 전환:', viewName);
    
    // 권한 확인
    if ((viewName === 'users' || viewName === 'departments') && 
        AppState.currentUser?.role !== 'admin') {
        showMessage('관리자만 접근 가능합니다', 'warning');
        return;
    }
    
    // 모든 뷰 숨기기
    const views = ['keys', 'users', 'profile', 'servers', 'departments'];
    views.forEach(view => {
        const element = DOM[view + 'View'];
        if (element) {
            element.classList.add('hidden');
        }
    });
    
    // 네비게이션 초기화
    const navs = ['navKeys', 'navUsers', 'navProfile', 'navServers', 'navDepartments'];
    navs.forEach(nav => {
        const element = DOM[nav];
        if (element) {
            element.classList.remove('active');
        }
    });
    
    // 선택된 뷰 표시
    const targetView = DOM[viewName + 'View'];
    const targetNav = DOM['nav' + viewName.charAt(0).toUpperCase() + viewName.slice(1)];
    
    if (targetView) {
        targetView.classList.remove('hidden');
        console.log(`✅ ${viewName}-view 활성화`);
    }
    if (targetNav) {
        targetNav.classList.add('active');
        console.log(`✅ nav-${viewName} 활성화`);
    }
    
    AppState.currentView = viewName;
    
    // 뷰별 초기화
    switch(viewName) {
        case 'keys':
            if (KeyManager && KeyManager.autoLoadKeys) {
                KeyManager.autoLoadKeys();
            }
            break;
        case 'users':
            if (UserManager && UserManager.loadUsersList) {
                UserManager.loadUsersList();
            }
            break;
        case 'profile':
            if (ProfileManager && ProfileManager.loadCurrentUserProfile) {
                ProfileManager.loadCurrentUserProfile();
            }
            break;
        case 'servers':
            showMessage('서버 관리 기능은 준비 중입니다', 'info');
            break;
        case 'departments':
            showMessage('부서 관리 기능은 준비 중입니다', 'info');
            break;
    }
}

// UI 업데이트 (AuthManager에서 호출)
function updateUI() {
    console.log('🎨 UI 업데이트 실행');
    const isAuthenticated = !!(AppState.jwtToken && AppState.currentUser);
    console.log('🔐 인증 상태:', isAuthenticated, AppState.currentUser);
    
    if (isAuthenticated) {
        // 로그인 상태
        console.log('✅ 로그인 상태 - 메인 화면으로 전환');
        
        if (DOM.authSection) {
            DOM.authSection.classList.add('hidden');
            console.log('  - 인증 섹션 숨김');
        }
        if (DOM.keySection) {
            DOM.keySection.classList.remove('hidden');
            console.log('  - 메인 섹션 표시');
        }
        if (DOM.container) {
            DOM.container.classList.add('container-wide');
            console.log('  - 컨테이너 확장');
        }
        
        // 네비게이션 업데이트
        updateNavigation();
        
        // 키 뷰로 이동
        showView('keys');
        
    } else {
        // 로그아웃 상태
        console.log('❌ 로그아웃 상태 - 인증 화면으로 전환');
        
        if (DOM.keySection) {
            DOM.keySection.classList.add('hidden');
            console.log('  - 메인 섹션 숨김');
        }
        if (DOM.authSection) {
            DOM.authSection.classList.remove('hidden');
            console.log('  - 인증 섹션 표시');
        }
        if (DOM.container) {
            DOM.container.classList.remove('container-wide');
            console.log('  - 컨테이너 축소');
        }
        
        showLoginView();
    }
    
    console.log('✅ UI 업데이트 완료');
}

// 네비게이션 업데이트
function updateNavigation() {
    const isAdmin = AppState.currentUser?.role === 'admin';
    console.log('👤 사용자 권한:', AppState.currentUser?.role, '/ 관리자:', isAdmin);
    
    // 관리자 전용 버튼 표시/숨김
    if (DOM.navUsers) {
        DOM.navUsers.style.display = isAdmin ? 'inline-block' : 'none';
    }
    if (DOM.navDepartments) {
        DOM.navDepartments.style.display = isAdmin ? 'inline-block' : 'none';
    }
    
    console.log('✅ 네비게이션 권한 업데이트 완료');
}

// 로그아웃 처리 (AuthManager에서 호출하거나 직접 호출)
function handleLogout() {
    console.log('🚪 로그아웃 처리 (script.js)');
    
    const username = AppState.currentUser?.username || '사용자';
    
    // 상태 초기화
    AppState.jwtToken = null;
    AppState.currentUser = null;
    localStorage.removeItem('jwtToken');
    
    // UI 업데이트
    updateUI();
    
    // 폼 초기화
    resetForms();
    
    showMessage(`${username}님, 로그아웃되었습니다`, 'info');
}

// 인증 화면 전환
function showLoginView() {
    if (DOM.registerView) {
        DOM.registerView.classList.add('hidden');
    }
    if (DOM.loginView) {
        DOM.loginView.classList.remove('hidden');
    }
    console.log('🔐 로그인 화면 표시');
}

function showRegisterView() {
    if (DOM.loginView) {
        DOM.loginView.classList.add('hidden');
    }
    if (DOM.registerView) {
        DOM.registerView.classList.remove('hidden');
    }
    console.log('📝 회원가입 화면 표시');
}

// 폼 초기화
function resetForms() {
    const forms = [DOM.loginForm, DOM.registerForm, DOM.profileForm];
    forms.forEach(form => {
        if (form) {
            try {
                form.reset();
            } catch (error) {
                console.warn('폼 초기화 실패:', error);
            }
        }
    });
}

// 자동 로그인 확인
async function checkAutoLogin() {
    if (AppState.jwtToken && AuthManager && AuthManager.validateToken) {
        console.log('🔍 자동 로그인 확인 중...');
        const isValid = await AuthManager.validateToken();
        if (isValid) {
            console.log('✅ 자동 로그인 성공');
            if (Utils && Utils.showToast) {
                Utils.showToast(`안녕하세요, ${AppState.currentUser?.username || '사용자'}님!`, 'success');
            }
            return true;
        } else {
            console.log('❌ 토큰 무효, 로그아웃 처리');
        }
    }
    return false;
}

// 메시지 표시
function showMessage(message, type = 'info') {
    if (Utils && Utils.showToast) {
        Utils.showToast(message, type);
    } else {
        console.log(`[${type.toUpperCase()}] ${message}`);
        // 중요한 메시지는 alert으로도 표시
        if (type === 'error' || type === 'warning') {
            alert(message);
        }
    }
}

// AppUtils - API 통신 및 유틸리티
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
                const error = new Error(result.error || result.message || `HTTP ${response.status}`);
                error.status = response.status;
                throw error;
            }
            
            // 🔥 API 응답 구조 정규화
            if (result.success && result.data) {
                // {success: true, data: {...}} 형태면 data 부분 반환
                console.log('📦 정규화된 데이터:', result.data);
                return result.data;
            } else {
                // 직접 데이터가 온 경우 그대로 반환
                return result;
            }
            
        } catch (error) {
            console.error(`❌ API 오류: ${method} ${endpoint}`, error);
            
            // Utils가 있으면 에러 처리 위임
            if (Utils && Utils.handleError) {
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
    },
    
    showError: function(message) {
        if (DOM.errorDisplay) {
            DOM.errorDisplay.textContent = message;
            DOM.errorDisplay.style.display = 'block';
        }
        console.error('앱 에러:', message);
    }
};

// 전역 함수 노출 (AuthManager 및 다른 매니저에서 사용)
window.updateUI = updateUI;
window.showView = showView;
window.handleLogout = handleLogout;
window.showLoginView = showLoginView;
window.showRegisterView = showRegisterView;

// 디버깅 도구 (간소화)
window.DEBUG = {
    state: function() {
        console.log('=== 🔍 현재 상태 ===');
        console.log('AppState:', AppState);
        console.log('인증 상태:', !!(AppState.jwtToken && AppState.currentUser));
        console.log('현재 뷰:', AppState.currentView);
        console.log('===============');
    },
    
    view: function(viewName) {
        console.log('🧪 뷰 전환 테스트:', viewName);
        showView(viewName);
    },
    
    logout: function() {
        console.log('🧪 로그아웃 테스트');
        handleLogout();
    },
    
    login: function() {
        console.log('🧪 로그인 시뮬레이션');
        AppState.jwtToken = 'test-token';
        AppState.currentUser = { id: 1, username: 'testuser', role: 'admin' };
        updateUI();
    }
};

console.log('🔧 === 디버깅 명령어 ===');
console.log('🔧 DEBUG.state() - 현재 상태 확인');
console.log('🔧 DEBUG.view("users") - 뷰 전환 테스트');
console.log('🔧 DEBUG.logout() - 로그아웃 테스트');
console.log('🔧 DEBUG.login() - 로그인 시뮬레이션');