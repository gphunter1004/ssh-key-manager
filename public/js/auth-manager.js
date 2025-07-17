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
            AppUtils.showError('사용자명과 비밀번호를 모두 입력해주세요.');
            return;
        }

        try {
            const data = await AppUtils.apiFetch('/login', 'POST', {
                username: username,
                password: password
            });

            AppState.jwtToken = data.token;
            localStorage.setItem('jwtToken', AppState.jwtToken);
            
            console.log('로그인 성공:', username);
            updateUI();
            KeyManager.hideKeys();

            // 로그인 폼 초기화
            e.target.reset();
            
        } catch (error) {
            console.error('로그인 실패:', error.message);
            // 에러는 이미 AppUtils.apiFetch에서 표시됨
        }
    },

    handleRegister: async function(e) {
        e.preventDefault();
        
        const username = e.target.elements['register-username'].value.trim();
        const password = e.target.elements['register-password'].value;

        if (!username || !password) {
            AppUtils.showError('사용자명과 비밀번호를 모두 입력해주세요.');
            return;
        }

        if (username.length < 2) {
            AppUtils.showError('사용자명은 최소 2자 이상이어야 합니다.');
            return;
        }

        if (password.length < 4) {
            AppUtils.showError('비밀번호는 최소 4자 이상이어야 합니다.');
            return;
        }

        try {
            await AppUtils.apiFetch('/register', 'POST', {
                username: username,
                password: password
            });

            alert('회원가입이 완료되었습니다! 로그인해주세요.');
            console.log('회원가입 성공:', username);
            
            // 폼 초기화 및 로그인 화면으로 전환
            e.target.reset();
            DOM.registerView.classList.add('hidden');
            DOM.loginView.classList.remove('hidden');
            
        } catch (error) {
            console.error('회원가입 실패:', error.message);
            // 에러는 이미 AppUtils.apiFetch에서 표시됨
        }
    },

    handleLogout: function() {
        const confirmLogout = confirm('정말 로그아웃하시겠습니까?');
        if (!confirmLogout) return;

        console.log('로그아웃 실행');
        
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
            AuthManager.handleLogout();
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