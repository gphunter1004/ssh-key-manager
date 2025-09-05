// 인증 관리자 - 완전 독립 버전
window.AuthManager = {
    /**
     * AuthManager를 초기화하고 이벤트 리스너를 설정합니다.
     */
    init: function() {
        console.log('🔐 AuthManager 초기화 시작');
        
        this.setupEventListeners();
        
        console.log('✅ AuthManager 초기화 완료');
        return true;
    },

    /**
     * 인증 관련 DOM 요소에 이벤트 리스너를 설정합니다.
     */
    setupEventListeners: function() {
        console.log('🎯 AuthManager 이벤트 리스너 설정');
        
        if (DOM.loginForm) {
            DOM.loginForm.onsubmit = (e) => this.handleLogin(e);
            console.log('✅ 로그인 폼 이벤트 설정');
        }

        if (DOM.registerForm) {
            DOM.registerForm.onsubmit = (e) => this.handleRegister(e);
            console.log('✅ 회원가입 폼 이벤트 설정');
        }

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

        if (DOM.logoutBtn) {
            DOM.logoutBtn.onclick = (e) => {
                e.preventDefault();
                this.handleLogout();
            };
            console.log('✅ 로그아웃 버튼 이벤트 설정');
        }

        console.log('✅ AuthManager 모든 이벤트 설정 완료');
    },

    /**
     * 로그인 폼 제출을 처리합니다.
     * @param {Event} e - 폼 제출 이벤트 객체
     */
    handleLogin: async function(e) {
        e.preventDefault();
        
        const username = e.target.elements['username']?.value?.trim();
        const password = e.target.elements['password']?.value;

        if (!username || !password) {
            Utils.showToast('사용자명과 비밀번호를 모두 입력해주세요', 'warning');
            return;
        }

        try {
            console.log('🔐 로그인 시도:', username);
            Utils.setLoading(true, '로그인 중...');
            
            const data = await AppUtils.apiFetch('/login', 'POST', {
                username: username,
                password: password
            });

            AppState.jwtToken = data.token;
            localStorage.setItem('jwtToken', AppState.jwtToken);
            
            AppState.currentUser = {
                id: data.user_id || data.id,
                username: data.username,
                role: data.role
            };
            
            console.log('✅ 로그인 성공:', AppState.currentUser);
            Utils.showToast(`환영합니다, ${username}님!`, 'success');
            
            this.switchToMainView();
            
            e.target.reset();
            
        } catch (error) {
            console.error('❌ 로그인 실패:', error.message);
            Utils.showToast('로그인 실패: ' + error.message, 'error');
        } finally {
            Utils.setLoading(false);
        }
    },

    /**
     * 회원가입 폼 제출을 처리합니다.
     * @param {Event} e - 폼 제출 이벤트 객체
     */
    handleRegister: async function(e) {
        e.preventDefault();
        
        const username = e.target.elements['username']?.value?.trim();
        const password = e.target.elements['password']?.value;

        if (!username || !password) {
            Utils.showToast('사용자명과 비밀번호를 모두 입력해주세요', 'warning');
            return;
        }

        if (username.length < 2) {
            Utils.showToast('사용자명은 최소 2자 이상이어야 합니다', 'warning');
            return;
        }

        if (password.length < 4) {
            Utils.showToast('비밀번호는 최소 4자 이상이어야 합니다', 'warning');
            return;
        }

        try {
            console.log('📝 회원가입 시도:', username);
            Utils.setLoading(true, '회원가입 중...');
            
            await AppUtils.apiFetch('/register', 'POST', {
                username: username,
                password: password
            });

            Utils.showToast(`${username}님, 회원가입이 완료되었습니다!`, 'success');
            console.log('✅ 회원가입 성공:', username);
            
            e.target.reset();
            this.showLoginView();
            
            if (DOM.loginForm && DOM.loginForm.elements['username']) {
                DOM.loginForm.elements['username'].value = username;
                DOM.loginForm.elements['password']?.focus();
            }
            
        } catch (error) {
            console.error('❌ 회원가입 실패:', error.message);
            Utils.showToast('회원가입 실패: ' + error.message, 'error');
        } finally {
            Utils.setLoading(false);
        }
    },

    /**
     * 로그아웃을 처리합니다.
     */
    handleLogout: function() {
        console.log('🚪 로그아웃 실행');
        
        const username = AppState.currentUser?.username || '사용자';
        
        try {
            Utils.setLoading(true, '로그아웃 중...');

            AppState.jwtToken = null;
            AppState.currentUser = null;
            AppState.currentView = 'keys';
            
            localStorage.removeItem('jwtToken');
            
            this.switchToAuthView();
            
            AppUtils.clearError();
            
            if (typeof KeyManager !== 'undefined' && KeyManager.hideKeys) {
                KeyManager.hideKeys();
            }
            
            this.resetAllForms();
            
            Utils.showToast(`${username}님, 안전하게 로그아웃되었습니다`, 'info');
        } catch (error) {
            console.error('❌ 로그아웃 실패:', error.message);
            Utils.showToast('로그아웃 중 오류가 발생했습니다', 'error');
        } finally {
            Utils.setLoading(false);
        }
    },

    /**
     * 로그인 성공 후 메인 화면으로 전환합니다.
     */
    switchToMainView: function() {
        console.log('🎨 메인 화면으로 전환');
        
        if (DOM.authSection) DOM.authSection.classList.add('hidden');
        if (DOM.keySection) DOM.keySection.classList.remove('hidden');
        if (DOM.container) DOM.container.classList.add('container-wide');
        
        this.updateNavigation();
        
        // ViewManager의 showView 호출이 비동기적으로 로딩 상태를 관리하므로,
        // 여기서는 별도의 로딩 상태 관리를 하지 않습니다.
        if (typeof ViewManager !== 'undefined' && ViewManager.showView) {
            ViewManager.showView('keys');
        } else {
            console.warn('⚠️ ViewManager를 찾을 수 없음. 뷰 전환을 건너뜁니다.');
        }
        
        console.log('✅ 메인 화면 전환 완료');
    },

    /**
     * 로그아웃 후 인증 화면으로 전환합니다.
     */
    switchToAuthView: function() {
        console.log('🎨 인증 화면으로 전환');
        
        if (DOM.keySection) DOM.keySection.classList.add('hidden');
        if (DOM.authSection) DOM.authSection.classList.remove('hidden');
        if (DOM.container) DOM.container.classList.remove('container-wide');
        
        this.showLoginView();
        
        console.log('✅ 인증 화면 전환 완료');
    },

    /**
     * 네비게이션 버튼을 사용자 역할에 따라 업데이트합니다.
     */
    updateNavigation: function() {
        const isAdmin = AppState.currentUser?.role === 'admin';
        console.log('👤 네비게이션 업데이트 - 관리자:', isAdmin);
        
        if (DOM.navUsers) DOM.navUsers.style.display = isAdmin ? 'inline-block' : 'none';
        if (DOM.navDepartments) DOM.navDepartments.style.display = isAdmin ? 'inline-block' : 'none';
        
        console.log('✅ 네비게이션 권한 업데이트 완료');
    },

    /**
     * 로그인 폼을 표시합니다.
     */
    showLoginView: function() {
        if (DOM.registerView) DOM.registerView.classList.add('hidden');
        if (DOM.loginView) DOM.loginView.classList.remove('hidden');
        console.log('🔐 로그인 화면 표시');
    },

    /**
     * 회원가입 폼을 표시합니다.
     */
    showRegisterView: function() {
        if (DOM.loginView) DOM.loginView.classList.add('hidden');
        if (DOM.registerView) DOM.registerView.classList.remove('hidden');
        console.log('📝 회원가입 화면 표시');
    },

    /**
     * JWT 토큰의 유효성을 검사합니다.
     * @returns {Promise<boolean>} - 토큰이 유효하면 true, 아니면 false
     */
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

    /**
     * 로컬 스토리지의 토큰을 사용하여 자동 로그인을 시도합니다.
     * @returns {Promise<boolean>} - 자동 로그인 성공 여부
     */
    checkLoginStatus: async function() {
        console.log('🔍 자동 로그인 확인 중...');
        
        if (AppState.jwtToken) {
            const isValid = await this.validateToken();
            if (isValid) {
                console.log('✅ 자동 로그인 성공');
                Utils.showToast(`안녕하세요, ${AppState.currentUser?.username || '사용자'}님!`, 'success');
                this.switchToMainView();
                return true;
            }
        }
        
        console.log('📝 저장된 토큰 없음 또는 유효하지 않음');
        this.switchToAuthView();
        return false;
    },

    /**
     * 모든 폼을 초기화합니다.
     */
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

    // 헬퍼 함수들 (외부에서 사용 가능한 메서드)
    isAuthenticated: function() {
        return !!(AppState.jwtToken && AppState.currentUser);
    },

    isAdmin: function() {
        return AppState.currentUser?.role === 'admin';
    },

    getCurrentUser: function() {
        return AppState.currentUser;
    }
};