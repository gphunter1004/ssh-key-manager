window.AuthManager = {
    init: function() {
        if (DOM.loginForm) DOM.loginForm.onsubmit = (e) => this.handleLogin(e);
        if (DOM.registerForm) DOM.registerForm.onsubmit = (e) => this.handleRegister(e);
        if (DOM.showRegisterLink) DOM.showRegisterLink.onclick = (e) => { e.preventDefault(); this.showRegisterView(); };
        if (DOM.showLoginLink) DOM.showLoginLink.onclick = (e) => { e.preventDefault(); this.showLoginView(); };
        if (DOM.logoutBtn) DOM.logoutBtn.onclick = (e) => { e.preventDefault(); this.handleLogout(); };
    },
    handleLogin: async function(e) {
        e.preventDefault();
        const username = e.target.elements['username'].value.trim();
        const password = e.target.elements['password'].value;
        if (!username || !password) return Utils.showToast('사용자명과 비밀번호를 입력해주세요.', 'warning');
        Utils.setLoading(true, '로그인 중...');
        try {
            const data = await AppUtils.apiFetch('/login', 'POST', { username, password });
            await this.processLoginSuccess(data);
            e.target.reset();
        } finally { Utils.setLoading(false); }
    },
    handleRegister: async function(e) {
        e.preventDefault();
        const username = e.target.elements['username'].value.trim();
        const password = e.target.elements['password'].value;
        if (!username || !password) return Utils.showToast('사용자명과 비밀번호를 입력해주세요.', 'warning');
        if (password.length < 4) return Utils.showToast('비밀번호는 4자 이상이어야 합니다.', 'warning');
        Utils.setLoading(true, '회원가입 중...');
        try {
            await AppUtils.apiFetch('/register', 'POST', { username, password });
            Utils.showToast(`${username}님, 회원가입이 완료되었습니다!`, 'success');
            this.showLoginView();
            DOM.loginForm.elements['username'].value = username;
            DOM.loginForm.elements['password'].focus();
            e.target.reset();
        } finally { Utils.setLoading(false); }
    },
    handleLogout: function() {
        Utils.showToast('안전하게 로그아웃되었습니다.', 'info');
        AppState.jwtToken = null; AppState.currentUser = null;
        localStorage.removeItem('jwtToken');
        this.switchToAuthView();
    },
    processLoginSuccess: async function(data) {
        AppState.jwtToken = data.token;
        localStorage.setItem('jwtToken', data.token);
        const isValid = await this.validateToken();
        if (isValid) {
            Utils.showToast(`환영합니다, ${AppState.currentUser.username}님!`, 'success');
            this.switchToMainView();
        } else { this.handleLogout(); }
    },
    checkLoginStatus: async function() {
        if (!AppState.jwtToken) return this.switchToAuthView();
        Utils.setLoading(true, '인증 확인 중...');
        const isValid = await this.validateToken();
        Utils.setLoading(false);
        if (isValid) { this.switchToMainView(); } 
        else { this.switchToAuthView(); }
    },
    validateToken: async function() {
        try {
            const data = await AppUtils.apiFetch('/validate');
            if (data && data.valid) {
                AppState.currentUser = { id: data.user_id, username: data.username, role: data.role };
                return true;
            }
        } catch (error) { /* silent fail */ }
        return false;
    },
    switchToMainView: function() {
        DOM.authSection.classList.add('hidden');
        DOM.keySection.classList.remove('hidden');
        DOM.container.classList.add('container-wide');
        ViewManager.updateNavigation();
        ViewManager.showView('keys');
    },
    switchToAuthView: function() {
        DOM.keySection.classList.add('hidden');
        DOM.authSection.classList.remove('hidden');
        DOM.container.classList.remove('container-wide');
        this.showLoginView();
    },
    showLoginView: () => { DOM.registerView.classList.add('hidden'); DOM.loginView.classList.remove('hidden'); },
    showRegisterView: () => { DOM.loginView.classList.add('hidden'); DOM.registerView.classList.remove('hidden'); },
    isAdmin: () => AppState.currentUser?.role === 'admin'
};