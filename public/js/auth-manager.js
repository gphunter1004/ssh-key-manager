// 인증 관리자
window.AuthManager = {
    setupEventListeners: function() {
        // 로그인/회원가입 폼 전환
        DOM.showRegisterLink.addEventListener('click', (e) => {
            e.preventDefault();
            AppUtils.clearError();
            DOM.loginView.classList.add('hidden');
            DOM.registerView.classList.remove('hidden');
        });

        DOM.showLoginLink.addEventListener('click', (e) => {
            e.preventDefault();
            AppUtils.clearError();
            DOM.registerView.classList.add('hidden');
            DOM.loginView.classList.remove('hidden');
        });

        // 로그인 폼 제출
        DOM.loginForm.addEventListener('submit', this.handleLogin);

        // 회원가입 폼 제출
        DOM.registerForm.addEventListener('submit', this.handleRegister);

        // 로그아웃
        DOM.logoutBtn.addEventListener('click', this.handleLogout);
    },

    handleLogin: async function(e) {
        e.preventDefault();
        
        const username = e.target.elements['login-username'].value.trim();
        const password = e.target.elements['login-password'].value;

        if (!username || !password) {
            Utils.showToast('사용자명과 비밀번호를 모두 입력해주세요', 'warning');
            return;
        }

        try {
            // 로딩 상태 표시
            Utils.setLoading(true, '로그인 중...');
            
            const data = await AppUtils.apiFetch('/login', 'POST', {
                username: username,
                password: password
            });

            AppState.jwtToken = data.token;
            localStorage.setItem('jwtToken', AppState.jwtToken);
            
            console.log('로그인 성공:', username);
            Utils.showToast(`환영합니다, ${username}님!`, 'success');
            
            updateUI();
            KeyManager.hideKeys();

            // 로그인 폼 초기화
            e.target.reset();
            
        } catch (error) {
            console.error('로그인 실패:', error.message);
            // 에러는 이미 Utils.handleError에서 처리됨
        } finally {
            Utils.setLoading(false);
        }
    },

    handleRegister: async function(e) {
        e.preventDefault();
        
        const username = e.target.elements['register-username'].value.trim();
        const password = e.target.elements['register-password'].value;

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
            // 로딩 상태 표시
            Utils.setLoading(true, '회원가입 중...');
            
            await AppUtils.apiFetch('/register', 'POST', {
                username: username,
                password: password
            });

            Utils.showToast(`${username}님, 회원가입이 완료되었습니다!`, 'success');
            console.log('회원가입 성공:', username);
            
            // 폼 초기화 및 로그인 화면으로 전환
            e.target.reset();
            DOM.registerView.classList.add('hidden');
            DOM.loginView.classList.remove('hidden');
            
            // 로그인 폼에 사용자명 자동 입력
            DOM.loginForm.elements['login-username'].value = username;
            DOM.loginForm.elements['login-password'].focus();
            
        } catch (error) {
            console.error('회원가입 실패:', error.message);
            // 에러는 이미 Utils.handleError에서 처리됨
        } finally {
            Utils.setLoading(false);
        }
    },

    handleLogout: function() {
        console.log('로그아웃 실행');
        
        const username = AppState.currentUser?.username || '사용자';
        
        // 상태 초기화
        AppState.jwtToken = null;
        AppState.currentUser = null;
        AppState.currentView = 'keys';
        
        // 로컬 스토리지에서 토큰 제거
        localStorage.removeItem('jwtToken');
        
        // UI 업데이트
        updateUI();
        AppUtils.clearError();
        KeyManager.hideKeys();
        
        // 모든 폼 초기화
        DOM.loginForm.reset();
        DOM.registerForm.reset();
        if (DOM.profileForm) {
            DOM.profileForm.reset();
        }
        
        Utils.showToast(`${username}님, 안전하게 로그아웃되었습니다`, 'info');
    },

    // 토큰 유효성 검사
    validateToken: async function() {
        if (!AppState.jwtToken) {
            return false;
        }

        try {
            // 현재 사용자 정보를 가져와서 토큰 유효성 확인
            const userData = await AppUtils.apiFetch('/users/me', 'GET');
            AppState.currentUser = userData;
            console.log('토큰 유효, 현재 사용자:', userData.username);
            return true;
        } catch (error) {
            console.error('토큰 유효성 검사 실패:', error.message);
            // 토큰이 무효한 경우 로그아웃 처리
            this.handleLogout();
            return false;
        }
    },

    // 페이지 로드 시 자동 로그인 확인
    checkAutoLogin: async function() {
        if (AppState.jwtToken) {
            const isValid = await this.validateToken();
            if (isValid) {
                console.log('자동 로그인 성공');
                updateUI();
            }
        }
    }
};