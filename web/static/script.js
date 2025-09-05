document.addEventListener('DOMContentLoaded', async () => {
    console.log('🚀 SSH Key Manager 시작');
    window.AppState = { jwtToken: localStorage.getItem('jwtToken') || null, currentUser: null, currentView: 'keys' };
    window.API_BASE_URL = '/api';
    initializeDOM();
    initializeManagers();
    await AuthManager.checkLoginStatus();
    console.log('✅ 초기화 완료');
});

function initializeDOM() {
    window.DOM = {
        container: document.querySelector('.container'),
        authSection: document.getElementById('auth-section'),
        keySection: document.getElementById('key-section'),
        loginView: document.getElementById('login-view'),
        registerView: document.getElementById('register-view'),
        loginForm: document.getElementById('login-form'),
        registerForm: document.getElementById('register-form'),
        showRegisterLink: document.getElementById('show-register'),
        showLoginLink: document.getElementById('show-login'),
        navKeys: document.getElementById('nav-keys'),
        navUsers: document.getElementById('nav-users'),
        navProfile: document.getElementById('nav-profile'),
        navServers: document.getElementById('nav-servers'),
        navDepartments: document.getElementById('nav-departments'),
        logoutBtn: document.getElementById('logout-btn'),
        keysView: document.getElementById('keys-view'),
        usersView: document.getElementById('users-view'),
        profileView: document.getElementById('profile-view'),
        serversView: document.getElementById('servers-view'),
        departmentsView: document.getElementById('departments-view'),
        userDetailModal: document.getElementById('user-detail-modal'),
        deployModal: document.getElementById('deploy-modal'),
        keyInfo: document.getElementById('key-info'),
        keyDisplayArea: document.getElementById('key-display-area'),
        serversList: document.getElementById('servers-list'),
        usersList: document.getElementById('users-list'),
        totalUsersSpan: document.getElementById('total-users'),
        usersWithKeysSpan: document.getElementById('users-with-keys'),
        profileForm: document.getElementById('profile-form'),
        currentUserInfo: document.getElementById('current-user-info'),
        errorDisplay: document.getElementById('error-display'),
    };
}

function initializeManagers() {
    const managers = [
        { name: 'Utils', manager: Utils }, { name: 'ModalManager', manager: ModalManager },
        { name: 'CopyManager', manager: CopyManager }, { name: 'AuthManager', manager: AuthManager },
        { name: 'ViewManager', manager: ViewManager }, { name: 'KeyManager', manager: KeyManager },
        { name: 'UserManager', manager: UserManager }, { name: 'ProfileManager', manager: ProfileManager },
        { name: 'ServerManager', manager: ServerManager }
    ];
    managers.forEach(({ name, manager }) => {
        try { if (manager && typeof manager.init === 'function') manager.init(); } 
        catch (error) { console.error(`⚠️ ${name} 초기화 실패:`, error); }
    });
}

window.AppUtils = {
    apiFetch: async function(endpoint, method = 'GET', body = null) {
        const headers = { 'Content-Type': 'application/json', 'Accept': 'application/json' };
        if (AppState.jwtToken) headers['Authorization'] = `Bearer ${AppState.jwtToken}`;
        const options = { method, headers };
        if (body) options.body = JSON.stringify(body);
        try {
            const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
            const result = await response.json();
            if (!response.ok) {
                const error = new Error(result.error?.message || result.message || `HTTP 에러: ${response.status}`);
                error.status = response.status;
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