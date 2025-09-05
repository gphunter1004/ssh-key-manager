// 인증 관리자 - 완전 독립 버전 (UI 전환까지 담당)
window.AuthManager = {
    init: function() {
        console.log('🔐 AuthManager 초기화 시작');
        
        // 이벤트 리스너 설정
        this.setupEventListeners();
        
        console.log('✅ AuthManager 초기화 완료');
        return true;
    },

    setupEventListeners: function() {
        console.log('🎯 AuthManager 이벤트 리스너 설정');
        
        // 로그인 폼
        if (DOM.loginForm) {
            DOM.loginForm.onsubmit = (e) => this.handleLogin(e);
            console.log('✅ 로그인 폼 이벤트 설정');
        }

        // 회원가입 폼
        if (DOM.registerForm) {
            DOM.registerForm.onsubmit = (e) => this.handleRegister(e);
            console.log('✅ 회원가입 폼 이벤트 설정');
        }

        // 로그인/회원가입 전환 링크
        if (DOM.showRegisterLink) {
            DOM.showRegisterLink.onclick = (e) => {
                e.preventDefault();
                this.showRegisterView();
            };
            console.log('✅ 회원가입 링크 이벤트 설정');
        }

        if (DOM.showLoginLink) {
            DOM.showLoginLink.onclick = (e) => {
                e.preventDefault();
                this.showLoginView();
            };
            console.log('✅ 로그인 링크 이벤트 설정');
        }

        // 로그아웃 버튼
        if (DOM.logoutBtn) {
            DOM.logoutBtn.onclick = (e) => {
                e.preventDefault();
                this.handleLogout();
            };
            console.log('✅ 로그아웃 버튼 이벤트 설정');
        }

        console.log('✅ AuthManager 모든 이벤트 설정 완료');
    },

    handleLogin: async function(e) {
        e.preventDefault();
        
        const username = e.target.elements['username']?.value?.trim();
        const password = e.target.elements['password']?.value;

        if (!username || !password) {
            this.showMessage('사용자명과 비밀번호를 모두 입력해주세요', 'warning');
            return;
        }

        try {
            console.log('🔐 로그인 시도:', username);
            this.setLoading(true, '로그인 중...');
            
            const data = await AppUtils.apiFetch('/login', 'POST', {
                username: username,
                password: password
            });

            // AppState 업데이트
            AppState.jwtToken = data.token;
            localStorage.setItem('jwtToken', AppState.jwtToken);
            
            AppState.currentUser = {
                id: data.user_id || data.id || 1,
                username: data.username || username,
                role: data.role || 'user'
            };
            
            console.log('✅ 로그인 성공:', AppState.currentUser);
            this.showMessage(`환영합니다, ${username}님!`, 'success');
            
            // 🔥 AuthManager가 직접 UI 전환 처리
            this.switchToMainView();
            
            // 로그인 폼 초기화
            e.target.reset();
            
        } catch (error) {
            console.error('❌ 로그인 실패:', error.message);
            this.showMessage('로그인 실패: ' + error.message, 'error');
        } finally {
            this.setLoading(false);
        }
    },

    handleRegister: async function(e) {
        e.preventDefault();
        
        const username = e.target.elements['username']?.value?.trim();
        const password = e.target.elements['password']?.value;

        if (!username || !password) {
            this.showMessage('사용자명과 비밀번호를 모두 입력해주세요', 'warning');
            return;
        }

        if (username.length < 2) {
            this.showMessage('사용자명은 최소 2자 이상이어야 합니다', 'warning');
            return;
        }

        if (password.length < 4) {
            this.showMessage('비밀번호는 최소 4자 이상이어야 합니다', 'warning');
            return;
        }

        try {
            console.log('📝 회원가입 시도:', username);
            this.setLoading(true, '회원가입 중...');
            
            await AppUtils.apiFetch('/register', 'POST', {
                username: username,
                password: password
            });

            this.showMessage(`${username}님, 회원가입이 완료되었습니다!`, 'success');
            console.log('✅ 회원가입 성공:', username);
            
            // 폼 초기화 및 로그인 화면으로 전환
            e.target.reset();
            this.showLoginView();
            
            // 로그인 폼에 사용자명 자동 입력
            if (DOM.loginForm && DOM.loginForm.elements['username']) {
                DOM.loginForm.elements['username'].value = username;
                DOM.loginForm.elements['password']?.focus();
            }
            
        } catch (error) {
            console.error('❌ 회원가입 실패:', error.message);
            this.showMessage('회원가입 실패: ' + error.message, 'error');
        } finally {
            this.setLoading(false);
        }
    },

    handleLogout: function() {
        console.log('🚪 로그아웃 실행');
        
        const username = AppState.currentUser?.username || '사용자';
        
        // 상태 초기화
        AppState.jwtToken = null;
        AppState.currentUser = null;
        AppState.currentView = 'keys';
        
        // 로컬 스토리지에서 토큰 제거
        localStorage.removeItem('jwtToken');
        
        // 🔥 AuthManager가 직접 UI 전환 처리
        this.switchToAuthView();
        
        // 에러 클리어
        AppUtils.clearError();
        
        // 키 정보 숨기기
        if (typeof KeyManager !== 'undefined' && KeyManager.hideKeys) {
            KeyManager.hideKeys();
        }
        
        // 모든 폼 초기화
        this.resetAllForms();
        
        this.showMessage(`${username}님, 안전하게 로그아웃되었습니다`, 'info');
    },

    // 🔥 AuthManager가 UI 전환 담당
    switchToMainView: function() {
        console.log('🎨 메인 화면으로 전환 (AuthManager)');
        
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
        this.updateNavigation();
        
        // ViewManager에게 키 뷰 표시 요청
        if (typeof ViewManager !== 'undefined' && ViewManager.showView) {
            ViewManager.showView('keys');
        } else {
            // ViewManager가 없으면 직접 처리
            this.showKeysViewDirect();
        }
        
        console.log('✅ 메인 화면 전환 완료');
    },

    switchToAuthView: function() {
        console.log('🎨 인증 화면으로 전환 (AuthManager)');
        
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
        
        // 로그인 화면 표시
        this.showLoginView();
        
        console.log('✅ 인증 화면 전환 완료');
    },

    updateNavigation: function() {
        const isAdmin = AppState.currentUser?.role === 'admin';
        console.log('👤 네비게이션 업데이트 - 관리자:', isAdmin);
        
        // 관리자 전용 버튼 표시/숨김
        if (DOM.navUsers) {
            DOM.navUsers.style.display = isAdmin ? 'inline-block' : 'none';
        }
        if (DOM.navDepartments) {
            DOM.navDepartments.style.display = isAdmin ? 'inline-block' : 'none';
        }
        
        console.log('✅ 네비게이션 권한 업데이트 완료');
    },

    showLoginView: function() {
        if (DOM.registerView) {
            DOM.registerView.classList.add('hidden');
        }
        if (DOM.loginView) {
            DOM.loginView.classList.remove('hidden');
        }
        console.log('🔐 로그인 화면 표시');
    },

    showRegisterView: function() {
        if (DOM.loginView) {
            DOM.loginView.classList.add('hidden');
        }
        if (DOM.registerView) {
            DOM.registerView.classList.remove('hidden');
        }
        console.log('📝 회원가입 화면 표시');
    },

    // ViewManager가 없을 때 직접 키 뷰 표시
    showKeysViewDirect: function() {
        // 모든 뷰 숨기기
        const views = ['keysView', 'usersView', 'profileView', 'serversView', 'departmentsView'];
        views.forEach(view => {
            if (DOM[view]) {
                DOM[view].classList.add('hidden');
            }
        });

        // 모든 네비게이션 비활성화
        const navs = ['navKeys', 'navUsers', 'navProfile', 'navServers', 'navDepartments'];
        navs.forEach(nav => {
            if (DOM[nav]) {
                DOM[nav].classList.remove('active');
            }
        });

        // 키 뷰 활성화
        if (DOM.keysView) {
            DOM.keysView.classList.remove('hidden');
        }
        if (DOM.navKeys) {
            DOM.navKeys.classList.add('active');
        }

        AppState.currentView = 'keys';

        // KeyManager에게 키 로드 요청
        if (typeof KeyManager !== 'undefined' && KeyManager.autoLoadKeys) {
            KeyManager.autoLoadKeys();
        }

        console.log('🔑 키 뷰 직접 표시 완료');
    },

    // 토큰 유효성 검사
    validateToken: async function() {
        if (!AppState.jwtToken) {
            return false;
        }

        try {
            const data = await AppUtils.apiFetch('/validate', 'GET');
            
            if (data.valid) {
                AppState.currentUser = {
                    id: data.user_id,
                    username: data.username,
                    role: data.role
                };
                console.log('🔍 토큰 검증 성공:', AppState.currentUser);
                return true;
            }
        } catch (error) {
            console.error('❌ 토큰 검증 실패:', error.message);
            this.handleLogout();
            return false;
        }
        
        return false;
    },

    // 자동 로그인 확인
    checkAutoLogin: async function() {
        console.log('🔍 자동 로그인 확인 중...');
        
        if (AppState.jwtToken) {
            const isValid = await this.validateToken();
            if (isValid) {
                console.log('✅ 자동 로그인 성공');
                this.showMessage(`안녕하세요, ${AppState.currentUser?.username || '사용자'}님!`, 'success');
                this.switchToMainView();
                return true;
            }
        } else {
            console.log('📝 저장된 토큰 없음');
            this.switchToAuthView();
        }
        return false;
    },

    // 헬퍼 함수들
    resetAllForms: function() {
        try {
            const forms = [DOM.loginForm, DOM.registerForm, DOM.profileForm];
            forms.forEach(form => {
                if (form) {
                    form.reset();
                }
            });
            console.log('✅ 모든 폼 초기화 완료');
        } catch (error) {
            console.warn('⚠️ 폼 초기화 중 오류:', error.message);
        }
    },

    setLoading: function(isLoading, message = '처리 중...') {
        if (typeof Utils !== 'undefined' && Utils.setLoading) {
            Utils.setLoading(isLoading, message);
        } else {
            console.log(`로딩 상태: ${isLoading} - ${message}`);
        }
    },

    showMessage: function(message, type = 'info') {
        if (typeof Utils !== 'undefined' && Utils.showToast) {
            Utils.showToast(message, type);
        } else {
            console.log(`[${type.toUpperCase()}] ${message}`);
            
            // 중요한 메시지는 alert으로도 표시
            if (type === 'error' || type === 'warning') {
                alert(message);
            }
        }
    },

    // 인증 상태 확인
    isAuthenticated: function() {
        return !!(AppState.jwtToken && AppState.currentUser);
    },

    // 관리자 권한 확인
    isAdmin: function() {
        return AppState.currentUser?.role === 'admin';
    },

    // 현재 사용자 정보 가져오기
    getCurrentUser: function() {
        return AppState.currentUser;
    }
};