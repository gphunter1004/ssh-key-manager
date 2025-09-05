// 메인 애플리케이션 스크립트 - 긴급 수정 버전
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

    // DOM 요소 초기화 - 더 안전한 방식
    const domInitialized = initializeDOMElements();
    if (!domInitialized) {
        console.error('❌ DOM 초기화 실패');
        // DOM 초기화 실패해도 계속 진행
    }
    
    // 매니저들 초기화
    initializeManagers();
    
    // 이벤트 리스너 설정
    setupEventListeners();
    
    // 자동 로그인 확인
    await checkAutoLogin();
    
    // 초기 UI 업데이트 - 강제로 실행
    forceUpdateUI();
    
    // 추가 설정들
    setupTokenRefresh();
    setupOfflineDetection();
    
    console.log('✅ 애플리케이션 초기화 완료');
});

function initializeDOMElements() {
    console.log('📋 DOM 요소 초기화 중...');
    
    try {
        window.DOM = {};
        
        // 필수 요소들을 하나씩 확인하고 설정
        const elements = {
            container: '.container',
            authSection: '#auth-section',
            keySection: '#key-section',
            loginView: '#login-view',
            registerView: '#register-view',
            errorDisplay: '#error-display',
            
            // 뷰들
            keysView: '#keys-view',
            usersView: '#users-view', 
            profileView: '#profile-view',
            
            // 네비게이션
            navKeys: '#nav-keys',
            navUsers: '#nav-users',
            navProfile: '#nav-profile',
            
            // 키 관련
            keyDisplayArea: '#key-display-area',
            keyInfo: '#key-info',
            keyPublicPre: '#key-public',
            keyPemPre: '#key-pem',
            keyPpkPre: '#key-ppk',
            
            // 폼들
            loginForm: '#login-form',
            registerForm: '#register-form',
            profileForm: '#profile-form',
            
            // 버튼들
            showRegisterLink: '#show-register',
            showLoginLink: '#show-login',
            logoutBtn: '#logout-btn',
            
            // 기타
            usersList: '#users-list',
            currentUserInfo: '#current-user-info',
            userDetailModal: '#user-detail-modal',
            userDetailContent: '#user-detail-content',
            closeModalBtn: '.close'
        };
        
        Object.entries(elements).forEach(([key, selector]) => {
            const element = document.querySelector(selector);
            DOM[key] = element;
            if (!element) {
                console.warn(`⚠️ ${key} (${selector}) 요소를 찾을 수 없습니다`);
            }
        });
        
        // 최소한 필요한 요소들 확인
        const criticalElements = ['authSection', 'keySection', 'loginForm'];
        const missingCritical = criticalElements.filter(key => !DOM[key]);
        
        if (missingCritical.length > 0) {
            console.error('❌ 필수 DOM 요소 누락:', missingCritical);
            return false;
        }
        
        console.log('✅ DOM 요소 초기화 완료');
        return true;
        
    } catch (error) {
        console.error('❌ DOM 초기화 중 오류:', error);
        return false;
    }
}

function initializeManagers() {
    console.log('🔧 매니저들 초기화 중...');
    
    try {
        // 매니저 초기화 - 안전하게 실행
        [
            () => Utils?.init?.(),
            () => CopyManager?.init?.(),
            () => KeyManager?.init?.(),
            () => ModalManager?.init?.(),
            () => ViewManager?.init?.(),
            () => AuthManager?.init?.()
        ].forEach((initFn, index) => {
            try {
                initFn();
            } catch (error) {
                console.warn(`⚠️ 매니저 ${index} 초기화 실패:`, error.message);
            }
        });
        
        console.log('✅ 매니저들 초기화 완료');
        
    } catch (error) {
        console.error('❌ 매니저 초기화 중 오류:', error);
    }
}

function setupEventListeners() {
    console.log('🎯 이벤트 리스너 설정 중...');
    
    try {
        // 네비게이션 이벤트 - 직접 이벤트 리스너 추가
        if (DOM.navKeys) {
            DOM.navKeys.addEventListener('click', (e) => {
                e.preventDefault();
                console.log('🔘 키 관리 버튼 클릭');
                showKeysView();
            });
        }
        
        if (DOM.navUsers) {
            DOM.navUsers.addEventListener('click', (e) => {
                e.preventDefault();
                console.log('🔘 사용자 목록 버튼 클릭');
                showUsersView();
            });
        }
        
        if (DOM.navProfile) {
            DOM.navProfile.addEventListener('click', (e) => {
                e.preventDefault();
                console.log('🔘 프로필 버튼 클릭');
                showProfileView();
            });
        }

        // 매니저 이벤트 리스너들 설정
        [
            () => AuthManager?.setupEventListeners?.(),
            () => UserManager?.setupEventListeners?.(),
            () => ProfileManager?.setupEventListeners?.(),
            () => ModalManager?.setupEventListeners?.(),
            () => CopyManager?.setupEventListeners?.()
        ].forEach((setupFn, index) => {
            try {
                setupFn();
            } catch (error) {
                console.warn(`⚠️ 매니저 ${index} 이벤트 설정 실패:`, error.message);
            }
        });
        
        console.log('✅ 이벤트 리스너 설정 완료');
        
    } catch (error) {
        console.error('❌ 이벤트 리스너 설정 중 오류:', error);
    }
}

// 직접 뷰 전환 함수들 (ViewManager가 작동하지 않을 때 사용)
function showKeysView() {
    console.log('🔑 키 뷰 표시');
    hideAllViews();
    resetNavigation();
    
    if (DOM.keysView) {
        DOM.keysView.classList.remove('hidden');
    }
    if (DOM.navKeys) {
        DOM.navKeys.classList.add('active');
    }
    
    // KeyManager 자동 로드
    if (KeyManager?.autoLoadKeys) {
        KeyManager.autoLoadKeys();
    }
}

function showUsersView() {
    console.log('👥 사용자 뷰 표시');
    
    // 권한 확인
    if (AppState.currentUser?.role !== 'admin') {
        Utils?.showToast?.('관리자만 접근 가능합니다', 'warning');
        return;
    }
    
    hideAllViews();
    resetNavigation();
    
    if (DOM.usersView) {
        DOM.usersView.classList.remove('hidden');
    }
    if (DOM.navUsers) {
        DOM.navUsers.classList.add('active');
    }
    
    // UserManager 로드
    if (UserManager?.loadUsersList) {
        UserManager.loadUsersList();
    }
}

function showProfileView() {
    console.log('👤 프로필 뷰 표시');
    hideAllViews();
    resetNavigation();
    
    if (DOM.profileView) {
        DOM.profileView.classList.remove('hidden');
    }
    if (DOM.navProfile) {
        DOM.navProfile.classList.add('active');
    }
    
    // ProfileManager 로드
    if (ProfileManager?.loadCurrentUserProfile) {
        ProfileManager.loadCurrentUserProfile();
    }
}

function hideAllViews() {
    const views = ['keysView', 'usersView', 'profileView'];
    views.forEach(view => {
        if (DOM[view]) {
            DOM[view].classList.add('hidden');
        }
    });
}

function resetNavigation() {
    const navs = ['navKeys', 'navUsers', 'navProfile'];
    navs.forEach(nav => {
        if (DOM[nav]) {
            DOM[nav].classList.remove('active');
        }
    });
}

// 강제 UI 업데이트 함수
function forceUpdateUI() {
    console.log('🎨 강제 UI 업데이트 실행');
    
    const isAuthenticated = !!(AppState.jwtToken && AppState.currentUser);
    console.log('🔐 인증 상태:', isAuthenticated);
    
    if (isAuthenticated) {
        console.log('✅ 로그인 상태 - 메인 섹션으로 전환');
        
        // 인증 섹션 숨기기
        if (DOM.authSection) {
            DOM.authSection.classList.add('hidden');
            console.log('  - 인증 섹션 숨김');
        }
        
        // 메인 섹션 표시
        if (DOM.keySection) {
            DOM.keySection.classList.remove('hidden');
            console.log('  - 메인 섹션 표시');
        }
        
        // 컨테이너 확장
        if (DOM.container) {
            DOM.container.classList.add('container-wide');
            console.log('  - 컨테이너 확장');
        }
        
        // 네비게이션 업데이트
        updateNavigation();
        
        // 키 뷰 표시
        showKeysView();
        
    } else {
        console.log('❌ 로그아웃 상태 - 인증 섹션으로 전환');
        
        // 메인 섹션 숨기기
        if (DOM.keySection) {
            DOM.keySection.classList.add('hidden');
            console.log('  - 메인 섹션 숨김');
        }
        
        // 인증 섹션 표시
        if (DOM.authSection) {
            DOM.authSection.classList.remove('hidden');
            console.log('  - 인증 섹션 표시');
        }
        
        // 컨테이너 축소
        if (DOM.container) {
            DOM.container.classList.remove('container-wide');
            console.log('  - 컨테이너 축소');
        }
        
        // 로그인 폼 표시
        if (DOM.registerView) {
            DOM.registerView.classList.add('hidden');
        }
        if (DOM.loginView) {
            DOM.loginView.classList.remove('hidden');
        }
    }
    
    console.log('✅ 강제 UI 업데이트 완료');
}

function updateNavigation() {
    const isAdmin = AppState.currentUser?.role === 'admin';
    console.log('📋 네비게이션 업데이트 - 관리자:', isAdmin);
    
    // 사용자 목록 버튼
    if (DOM.navUsers) {
        if (isAdmin) {
            DOM.navUsers.style.display = 'inline-block';
            console.log('  - 사용자 관리 버튼 표시');
        } else {
            DOM.navUsers.style.display = 'none';
            console.log('  - 사용자 관리 버튼 숨김');
        }
    }
}

// 기존 updateUI 함수를 forceUpdateUI로 대체
function updateUI() {
    forceUpdateUI();
}

async function checkAutoLogin() {
    console.log('🔐 자동 로그인 확인 중...');
    
    if (AppState.jwtToken && typeof AuthManager !== 'undefined') {
        const isValid = await AuthManager.validateToken();
        if (isValid) {
            console.log('✅ 자동 로그인 성공');
            Utils?.showToast?.(`안녕하세요, ${AppState.currentUser?.username || '사용자'}님!`, 'success');
        } else {
            console.log('❌ 자동 로그인 실패');
        }
    } else {
        console.log('📝 저장된 토큰 없음');
    }
}

function setupTokenRefresh() {
    setInterval(async () => {
        if (AppState.jwtToken) {
            try {
                await AppUtils.apiFetch('/users/me');
                console.log('🔄 토큰 유효성 확인 완료');
            } catch (error) {
                if (error.message.includes('401') || error.message.includes('expired')) {
                    console.warn('🔐 토큰이 만료되어 자동 로그아웃됩니다');
                    Utils?.showToast?.('세션이 만료되어 로그아웃됩니다', 'warning');
                    setTimeout(() => AuthManager?.handleLogout?.(), 2000);
                }
            }
        }
    }, 5 * 60 * 1000);
    
    console.log('🔄 자동 토큰 갱신 설정 완료');
}

function setupOfflineDetection() {
    window.addEventListener('offline', () => {
        console.log('📡 네트워크 연결 끊김');
        Utils?.showToast?.('인터넷 연결이 끊어졌습니다', 'warning', 5000);
    });

    window.addEventListener('online', () => {
        console.log('🌐 네트워크 연결됨');
        Utils?.showToast?.('인터넷 연결이 복원되었습니다', 'success');
        AppUtils?.clearError?.();
    });
    
    console.log('📡 오프라인 감지 설정 완료');
}

// AppUtils - 더 안전한 버전
window.AppUtils = {
    showError: function(message) {
        if (DOM.errorDisplay) {
            DOM.errorDisplay.textContent = message;
            DOM.errorDisplay.style.display = 'block';
        }
        console.error('앱 에러:', message);
        
        setTimeout(() => {
            this.clearError();
        }, 10000);
    },
    
    clearError: function() {
        if (DOM.errorDisplay) {
            DOM.errorDisplay.textContent = '';
            DOM.errorDisplay.style.display = 'none';
        }
    },
    
    apiFetch: async function(endpoint, method = 'GET', body = null) {
        this.clearError();
        
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
            credentials: 'same-origin'
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
            
            if (Utils?.handleError) {
                Utils.handleError(error, `API ${method} ${endpoint}`);
            } else {
                console.error('Utils.handleError를 사용할 수 없음');
            }
            
            throw error;
        }
    }
};

// 전역 함수로 노출
window.updateUI = updateUI;
window.forceUpdateUI = forceUpdateUI;
window.showKeysView = showKeysView;
window.showUsersView = showUsersView;
window.showProfileView = showProfileView;

// 전역 에러 핸들러
window.addEventListener('error', (event) => {
    console.error('🚨 전역 에러:', event.error);
});

// 디버깅 헬퍼
window.DEBUG_HELPERS = {
    diagnose: function() {
        console.log('=== 🔍 전체 진단 ===');
        console.log('AppState:', AppState);
        console.log('DOM 요소들:');
        Object.entries(DOM).forEach(([key, element]) => {
            console.log(`  ${key}:`, element ? '✅' : '❌');
        });
        console.log('CSS 상태:');
        if (DOM.authSection) {
            console.log(`  auth-section: display=${getComputedStyle(DOM.authSection).display}, hidden=${DOM.authSection.classList.contains('hidden')}`);
        }
        if (DOM.keySection) {
            console.log(`  key-section: display=${getComputedStyle(DOM.keySection).display}, hidden=${DOM.keySection.classList.contains('hidden')}`);
        }
        console.log('================');
    },
    
    forceLogin: function() {
        console.log('🧪 강제 로그인 테스트');
        AppState.jwtToken = 'test-token';
        AppState.currentUser = { id: 1, username: 'testuser', role: 'user' };
        forceUpdateUI();
    },
    
    forceLogout: function() {
        console.log('🧪 강제 로그아웃 테스트');
        AppState.jwtToken = null;
        AppState.currentUser = null;
        forceUpdateUI();
    }
};

console.log('🔧 DEBUG_HELPERS.diagnose() - 진단');
console.log('🔧 DEBUG_HELPERS.forceLogin() - 강제 로그인');
console.log('🔧 forceUpdateUI() - 강제 UI 업데이트');