// SSH Key Manager - 간소화된 메인 스크립트
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
        userDetailContent: document.getElementById('user-detail-content')
    };
}

// 매니저 초기화
function initializeManagers() {
    const managers = [Utils, CopyManager, KeyManager, ModalManager, ViewManager, AuthManager];
    
    managers.forEach(manager => {
        try {
            if (manager && manager.init) {
                manager.init();
            }
        } catch (error) {
            console.warn(`매니저 초기화 실패:`, error.message);
        }
    });
}

// 이벤트 설정
function setupEvents() {
    // 인증 이벤트
    setupAuthEvents();
    
    // 네비게이션 이벤트
    setupNavigationEvents();
    
    // 매니저 이벤트
    const managers = [AuthManager, UserManager, ProfileManager, ModalManager, CopyManager];
    managers.forEach(manager => {
        try {
            if (manager && manager.setupEventListeners) {
                manager.setupEventListeners();
            }
        } catch (error) {
            console.warn(`매니저 이벤트 설정 실패:`, error.message);
        }
    });
}

// 인증 이벤트 설정
function setupAuthEvents() {
    // 로그인 폼
    if (DOM.loginForm) {
        DOM.loginForm.onsubmit = async function(e) {
            e.preventDefault();
            const username = this.elements['username']?.value?.trim();
            const password = this.elements['password']?.value;
            
            if (!username || !password) {
                showMessage('사용자명과 비밀번호를 입력해주세요', 'warning');
                return;
            }
            
            await handleLogin(username, password);
        };
    }
    
    // 회원가입 폼
    if (DOM.registerForm) {
        DOM.registerForm.onsubmit = async function(e) {
            e.preventDefault();
            const username = this.elements['username']?.value?.trim();
            const password = this.elements['password']?.value;
            
            if (!username || !password) {
                showMessage('사용자명과 비밀번호를 입력해주세요', 'warning');
                return;
            }
            
            if (username.length < 2 || password.length < 4) {
                showMessage('사용자명 2자 이상, 비밀번호 4자 이상 입력해주세요', 'warning');
                return;
            }
            
            await handleRegister(username, password);
        };
    }
    
    // 화면 전환 링크
    if (DOM.showRegisterLink) {
        DOM.showRegisterLink.onclick = function(e) {
            e.preventDefault();
            showRegisterView();
        };
    }
    
    if (DOM.showLoginLink) {
        DOM.showLoginLink.onclick = function(e) {
            e.preventDefault();
            showLoginView();
        };
    }
}

// 네비게이션 이벤트 설정
function setupNavigationEvents() {
    const navButtons = [
        { element: DOM.navKeys, view: 'keys' },
        { element: DOM.navUsers, view: 'users' },
        { element: DOM.navProfile, view: 'profile' },
        { element: DOM.navServers, view: 'servers' },
        { element: DOM.navDepartments, view: 'departments' }
    ];
    
    navButtons.forEach(({ element, view }) => {
        if (element) {
            element.onclick = function(e) {
                e.preventDefault();
                showView(view);
            };
        }
    });
    
    // 로그아웃 버튼
    if (DOM.logoutBtn) {
        DOM.logoutBtn.onclick = function(e) {
            e.preventDefault();
            handleLogout();
        };
    }
}

// 로그인 처리
async function handleLogin(username, password) {
    try {
        showMessage('로그인 중...', 'info');
        
        const data = await apiFetch('/login', 'POST', { username, password });
        
        AppState.jwtToken = data.token;
        AppState.currentUser = {
            id: data.user_id || data.id,
            username: data.username,
            role: data.role
        };
        
        localStorage.setItem('jwtToken', AppState.jwtToken);
        
        showMessage(`환영합니다, ${username}님!`, 'success');
        DOM.loginForm.reset();
        
        updateUI();
        
    } catch (error) {
        showMessage('로그인 실패: ' + error.message, 'error');
    }
}

// 회원가입 처리
async function handleRegister(username, password) {
    try {
        showMessage('회원가입 중...', 'info');
        
        await apiFetch('/register', 'POST', { username, password });
        
        showMessage(`${username}님, 회원가입 완료!`, 'success');
        DOM.registerForm.reset();
        
        showLoginView();
        
        // 로그인 폼에 사용자명 자동 입력
        if (DOM.loginForm.elements['username']) {
            DOM.loginForm.elements['username'].value = username;
        }
        
    } catch (error) {
        showMessage('회원가입 실패: ' + error.message, 'error');
    }
}

// 로그아웃 처리
function handleLogout() {
    const username = AppState.currentUser?.username || '사용자';
    
    AppState.jwtToken = null;
    AppState.currentUser = null;
    localStorage.removeItem('jwtToken');
    
    updateUI();
    
    // 폼 초기화
    if (DOM.loginForm) DOM.loginForm.reset();
    if (DOM.registerForm) DOM.registerForm.reset();
    
    showMessage(`${username}님, 로그아웃되었습니다`, 'info');
}

// 뷰 전환
function showView(viewName) {
    // 권한 확인
    if ((viewName === 'users' || viewName === 'departments') && 
        AppState.currentUser?.role !== 'admin') {
        showMessage('관리자만 접근 가능합니다', 'warning');
        return;
    }
    
    // 모든 뷰 숨기기
    ['keys', 'users', 'profile', 'servers', 'departments'].forEach(view => {
        const element = DOM[view + 'View'];
        if (element) {
            element.classList.add('hidden');
        }
    });
    
    // 네비게이션 초기화
    ['navKeys', 'navUsers', 'navProfile', 'navServers', 'navDepartments'].forEach(nav => {
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
    }
    if (targetNav) {
        targetNav.classList.add('active');
    }
    
    AppState.currentView = viewName;
    
    // 뷰별 초기화
    switch(viewName) {
        case 'keys':
            if (KeyManager?.autoLoadKeys) KeyManager.autoLoadKeys();
            break;
        case 'users':
            if (UserManager?.loadUsersList) UserManager.loadUsersList();
            break;
        case 'profile':
            if (ProfileManager?.loadCurrentUserProfile) ProfileManager.loadCurrentUserProfile();
            break;
        case 'servers':
            showMessage('서버 관리 기능은 준비 중입니다', 'info');
            break;
        case 'departments':
            showMessage('부서 관리 기능은 준비 중입니다', 'info');
            break;
    }
}

// UI 업데이트
function updateUI() {
    const isAuthenticated = !!(AppState.jwtToken && AppState.currentUser);
    
    if (isAuthenticated) {
        // 로그인 상태
        if (DOM.authSection) DOM.authSection.classList.add('hidden');
        if (DOM.keySection) DOM.keySection.classList.remove('hidden');
        if (DOM.container) DOM.container.classList.add('container-wide');
        
        // 네비게이션 업데이트
        const isAdmin = AppState.currentUser.role === 'admin';
        if (DOM.navUsers) DOM.navUsers.style.display = isAdmin ? 'inline-block' : 'none';
        if (DOM.navDepartments) DOM.navDepartments.style.display = isAdmin ? 'inline-block' : 'none';
        
        showView('keys');
        
    } else {
        // 로그아웃 상태
        if (DOM.keySection) DOM.keySection.classList.add('hidden');
        if (DOM.authSection) DOM.authSection.classList.remove('hidden');
        if (DOM.container) DOM.container.classList.remove('container-wide');
        
        showLoginView();
    }
}

// 인증 화면 전환
function showLoginView() {
    if (DOM.registerView) DOM.registerView.classList.add('hidden');
    if (DOM.loginView) DOM.loginView.classList.remove('hidden');
}

function showRegisterView() {
    if (DOM.loginView) DOM.loginView.classList.add('hidden');
    if (DOM.registerView) DOM.registerView.classList.remove('hidden');
}

// 자동 로그인 확인
async function checkAutoLogin() {
    if (AppState.jwtToken) {
        try {
            const data = await apiFetch('/validate', 'GET');
            if (data.valid) {
                AppState.currentUser = {
                    id: data.user_id,
                    username: data.username,
                    role: data.role
                };
                console.log('자동 로그인 성공');
                return true;
            }
        } catch (error) {
            console.log('토큰 무효, 로그아웃 처리');
            AppState.jwtToken = null;
            localStorage.removeItem('jwtToken');
        }
    }
    return false;
}

// API 호출
async function apiFetch(endpoint, method = 'GET', body = null) {
    const headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    };
    
    if (AppState.jwtToken) {
        headers['Authorization'] = `Bearer ${AppState.jwtToken}`;
    }
    
    const options = { method, headers };
    if (body) options.body = JSON.stringify(body);
    
    const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
    const data = await response.json();
    
    if (!response.ok) {
        const error = new Error(data.error || data.message || `HTTP ${response.status}`);
        error.status = response.status;
        throw error;
    }
    
    return data;
}

// 메시지 표시
function showMessage(message, type = 'info') {
    if (Utils && Utils.showToast) {
        Utils.showToast(message, type);
    } else {
        console.log(`[${type.toUpperCase()}] ${message}`);
        
        // 간단한 알림
        if (type === 'error' || type === 'warning') {
            alert(message);
        }
    }
}

// AppUtils
window.AppUtils = {
    apiFetch: apiFetch,
    clearError: function() {
        if (DOM.errorDisplay) {
            DOM.errorDisplay.textContent = '';
            DOM.errorDisplay.style.display = 'none';
        }
    }
};

// 전역 함수 노출
window.updateUI = updateUI;
window.showView = showView;
window.handleLogin = handleLogin;
window.handleLogout = handleLogout;

// 디버깅 도구 (간소화)
window.DEBUG = {
    login: function(username = 'admin', password = 'admin') {
        handleLogin(username, password);
    },
    logout: function() {
        handleLogout();
    },
    view: function(viewName) {
        showView(viewName);
    }
};

console.log('🔧 DEBUG.login("사용자명", "비밀번호") - 테스트 로그인');
console.log('🔧 DEBUG.logout() - 로그아웃');
console.log('🔧 DEBUG.view("keys") - 뷰 전환');