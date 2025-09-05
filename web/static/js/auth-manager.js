// 인증 관리자 - 개선된 버전
window.AuthManager = {
    setupEventListeners: function() {
        // 로그인/회원가입 폼 전환
        if (DOM.showRegisterLink) {
            DOM.showRegisterLink.addEventListener('click', (e) => {
                e.preventDefault();
                AppUtils.clearError();
                DOM.loginView?.classList.add('hidden');
                DOM.registerView?.classList.remove('hidden');
            });
        }

        if (DOM.showLoginLink) {
            DOM.showLoginLink.addEventListener('click', (e) => {
                e.preventDefault();
                AppUtils.clearError();
                DOM.registerView?.classList.add('hidden');
                DOM.loginView?.classList.remove('hidden');
            });
        }

        // 로그인 폼 제출
        if (DOM.loginForm) {
            DOM.loginForm.addEventListener('submit', this.handleLogin);
        }

        // 회원가입 폼 제출
        if (DOM.registerForm) {
            DOM.registerForm.addEventListener('submit', this.handleRegister);
        }

        // 로그아웃
        if (DOM.logoutBtn) {
            DOM.logoutBtn.addEventListener('click', this.handleLogout);
        }
    },

    handleLogin: async function(e) {
        e.preventDefault();
        
        const username = e.target.elements['username']?.value?.trim();
        const password = e.target.elements['password']?.value;

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

            console.log('🔍 로그인 API 전체 응답:', data);

            // 토큰 추출 (중첩된 구조와 평면 구조 모두 지원)
            AppState.jwtToken = data.token || (data.data && data.data.token);
            localStorage.setItem('jwtToken', AppState.jwtToken);
            
            // 로그인 후 사용자 정보 설정 (API 응답 구조에 맞춰 수정)
            if (data.data) {
                // 응답이 { success: true, message: "", data: { token, username, role } } 형태인 경우
                AppState.currentUser = {
                    id: data.data.user_id || data.data.id,
                    username: data.data.username,
                    role: data.data.role
                };
            } else {
                // 응답이 직접 { token, username, role } 형태인 경우  
                AppState.currentUser = {
                    id: data.user_id || data.id,
                    username: data.username,
                    role: data.role
                };
            }
            
            console.log('✅ 로그인 성공:', AppState.currentUser);
            Utils.showToast(`환영합니다, ${AppState.currentUser.username}님!`, 'success');
            
            // UI 업데이트 (먼저 수행)
            updateUI();
            
            // KeyManager가 존재하고 초기화되었을 때만 키 숨김 처리
            AuthManager.safeHideKeys();

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
            // 로딩 상태 표시
            Utils.setLoading(true, '회원가입 중...');
            
            const data = await AppUtils.apiFetch('/register', 'POST', {
                username: username,
                password: password
            });

            Utils.showToast(`${username}님, 회원가입이 완료되었습니다!`, 'success');
            console.log('회원가입 성공:', username);
            
            // 폼 초기화 및 로그인 화면으로 전환
            e.target.reset();
            DOM.registerView?.classList.add('hidden');
            DOM.loginView?.classList.remove('hidden');
            
            // 로그인 폼에 사용자명 자동 입력
            if (DOM.loginForm && DOM.loginForm.elements['username']) {
                DOM.loginForm.elements['username'].value = username;
                DOM.loginForm.elements['password']?.focus();
            }
            
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
        
        // KeyManager가 존재할 때만 키 숨김 처리
        AuthManager.safeHideKeys();
        
        // 모든 폼 초기화
        AuthManager.resetAllForms();
        
        Utils.showToast(`${username}님, 안전하게 로그아웃되었습니다`, 'info');
    },

    // 토큰 유효성 검사
    validateToken: async function() {
        if (!AppState.jwtToken) {
            return false;
        }

        try {
            // 토큰 검증 API 호출
            const data = await AppUtils.apiFetch('/validate', 'GET');
            
            if (data.valid || data.success) {
                // 응답 구조에 따라 사용자 정보 추출
                const userInfo = data.data || data;
                AppState.currentUser = {
                    id: userInfo.user_id || userInfo.id,
                    username: userInfo.username,
                    role: userInfo.role
                };
                console.log('🔍 토큰 검증 성공, 사용자 정보:', AppState.currentUser);
                return true;
            }
        } catch (error) {
            console.error('토큰 검증 실패:', error.message);
            // 토큰이 무효한 경우에만 로그아웃 처리 (404 등의 일반 에러는 제외)
            if (error.status === 401 || error.status === 403 || error.message.includes('token') || error.message.includes('unauthorized')) {
                this.handleLogout();
            }
            return false;
        }
        
        return false;
    },

    // 페이지 로드 시 자동 로그인 확인
    checkAutoLogin: async function() {
        if (AppState.jwtToken) {
            const isValid = await this.validateToken();
            if (isValid) {
                console.log('자동 로그인 성공');
                updateUI();
                return true;
            }
        }
        return false;
    },

    // ========== 안전한 헬퍼 함수들 ==========

    // 안전한 키 숨김 처리
    safeHideKeys: function() {
        try {
            if (typeof KeyManager !== 'undefined' && KeyManager.hideKeys) {
                KeyManager.hideKeys();
            } else {
                console.log('ℹ️ KeyManager를 사용할 수 없음, 키 숨김 처리 건너뜀');
            }
        } catch (error) {
            console.warn('⚠️ KeyManager.hideKeys 호출 중 오류:', error.message);
            // 치명적이지 않은 오류이므로 계속 진행
        }
    },

    // 모든 폼 초기화
    resetAllForms: function() {
        try {
            if (DOM.loginForm) {
                DOM.loginForm.reset();
            }
            if (DOM.registerForm) {
                DOM.registerForm.reset();
            }
            if (DOM.profileForm) {
                DOM.profileForm.reset();
            }
        } catch (error) {
            console.warn('⚠️ 폼 초기화 중 오류:', error.message);
        }
    },

    // 인증 상태 확인
    isAuthenticated: function() {
        return !!(AppState.jwtToken && AppState.currentUser && AppState.currentUser.username);
    },

    // 관리자 권한 확인
    isAdmin: function() {
        return AppState.currentUser?.role === 'admin';
    },

    // 현재 사용자 정보 가져오기
    getCurrentUser: function() {
        return AppState.currentUser;
    },

    // 토큰 새로고침
    refreshToken: async function() {
        if (!AppState.jwtToken) {
            return false;
        }

        try {
            const data = await AppUtils.apiFetch('/refresh', 'POST');
            
            if (data.token || (data.data && data.data.token)) {
                AppState.jwtToken = data.token || data.data.token;
                localStorage.setItem('jwtToken', AppState.jwtToken);
                console.log('🔄 토큰 새로고침 성공');
                return true;
            }
        } catch (error) {
            console.error('토큰 새로고침 실패:', error.message);
            this.handleLogout();
        }
        
        return false;
    },

    // 세션 연장 (사용자 활동 감지 시 호출)
    extendSession: function() {
        if (this.isAuthenticated()) {
            // 마지막 활동 시간 업데이트
            localStorage.setItem('lastActivity', Date.now().toString());
            
            // 토큰 만료가 임박했다면 새로고침
            this.checkTokenExpiry();
        }
    },

    // 토큰 만료 체크
    checkTokenExpiry: function() {
        if (!AppState.jwtToken) return;

        try {
            // JWT 페이로드 파싱 (간단한 방법)
            const payload = JSON.parse(atob(AppState.jwtToken.split('.')[1]));
            const now = Math.floor(Date.now() / 1000);
            const exp = payload.exp;
            
            // 만료 10분 전이면 새로고침 시도
            if (exp && (exp - now) < 600) {
                console.log('🔄 토큰 만료 임박, 새로고침 시도');
                this.refreshToken();
            }
        } catch (error) {
            console.warn('토큰 만료 체크 실패:', error.message);
        }
    },

    // 초기화 함수
    init: function() {
        console.log('AuthManager 초기화');
        
        // 이벤트 리스너 설정
        this.setupEventListeners();
        
        // 사용자 활동 감지 설정
        this.setupActivityDetection();
        
        console.log('✅ AuthManager 초기화 완료');
    },

    // 사용자 활동 감지 설정
    setupActivityDetection: function() {
        // 사용자 활동 이벤트들
        const activityEvents = ['click', 'keypress', 'scroll', 'mousemove'];
        
        let activityTimer;
        
        const handleActivity = () => {
            if (this.isAuthenticated()) {
                // 디바운싱: 1분에 한 번만 세션 연장
                clearTimeout(activityTimer);
                activityTimer = setTimeout(() => {
                    this.extendSession();
                }, 60000); // 1분
            }
        };
        
        activityEvents.forEach(event => {
            document.addEventListener(event, handleActivity, { passive: true });
        });
        
        console.log('👁️ 사용자 활동 감지 설정 완료');
    }
};