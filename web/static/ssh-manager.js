// ===============================
// 통합된 SSH Key Manager
// ===============================

// 전역 상태 관리
const AppState = {
    jwtToken: localStorage.getItem('jwtToken') || null,
    currentUser: null,
    currentView: 'keys'
};

const API_BASE_URL = '/api';

// ===============================
// 공통 유틸리티
// ===============================
const Utils = {
    // 토스트 알림
    showToast(message, type = 'success', duration = 3000) {
        const existingToast = document.getElementById('app-toast');
        if (existingToast) existingToast.remove();
        
        const toast = document.createElement('div');
        toast.id = 'app-toast';
        toast.innerHTML = `
            <span>${type === 'success' ? '✅' : type === 'error' ? '❌' : '⚠️'} ${message}</span>
            <button onclick="this.parentElement.remove()">×</button>
        `;
        
        Object.assign(toast.style, {
            position: 'fixed', top: '20px', right: '20px', padding: '12px 16px',
            borderRadius: '8px', color: 'white', fontSize: '14px', zIndex: '10000',
            backgroundColor: type === 'success' ? '#27ae60' : type === 'error' ? '#e74c3c' : '#f39c12',
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)', display: 'flex', alignItems: 'center', gap: '8px'
        });
        
        document.body.appendChild(toast);
        setTimeout(() => toast.remove(), duration);
    },

    // 로딩 상태
    setLoading(isLoading, message = '처리 중...') {
        const buttons = document.querySelectorAll('button:not(.toast button)');
        if (isLoading) {
            document.body.style.cursor = 'wait';
            buttons.forEach(btn => { btn.disabled = true; btn.style.opacity = '0.6'; });
            this.showLoadingIndicator(message);
        } else {
            document.body.style.cursor = '';
            buttons.forEach(btn => { btn.disabled = false; btn.style.opacity = '1'; });
            this.hideLoadingIndicator();
        }
    },

    showLoadingIndicator(message) {
        this.hideLoadingIndicator();
        const indicator = document.createElement('div');
        indicator.id = 'loading-indicator';
        indicator.innerHTML = `
            <div style="position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.3);
                        display:flex;align-items:center;justify-content:center;z-index:9999;">
                <div style="background:white;padding:20px;border-radius:8px;text-align:center;">
                    <div style="width:30px;height:30px;border:3px solid #f3f3f3;border-top:3px solid #3498db;
                               border-radius:50%;animation:spin 1s linear infinite;margin:0 auto 10px;"></div>
                    <div>${message}</div>
                </div>
            </div>
        `;
        document.body.appendChild(indicator);
    },

    hideLoadingIndicator() {
        const indicator = document.getElementById('loading-indicator');
        if (indicator) indicator.remove();
    },

    // HTML 이스케이프
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    },

    // 날짜 포맷팅
    formatDate(dateString) {
        try {
            return new Date(dateString).toLocaleString('ko-KR');
        } catch {
            return dateString;
        }
    }
};

// ===============================
// API 통신
// ===============================
const API = {
    async request(endpoint, method = 'GET', body = null) {
        const headers = { 'Content-Type': 'application/json' };
        if (AppState.jwtToken) {
            headers['Authorization'] = `Bearer ${AppState.jwtToken}`;
        }

        const options = { method, headers };
        if (body) options.body = JSON.stringify(body);

        try {
            const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
            const data = await response.json();

            if (!response.ok) {
                const error = new Error(data.error || `HTTP ${response.status}`);
                error.status = response.status;
                throw error;
            }

            // 키 관련 엔드포인트는 래핑 없이 직접 데이터를 반환
            if (endpoint.includes('/keys')) {
                return data;
            }
            
            // 다른 엔드포인트는 data 래핑 체크
            return data.data !== undefined ? data.data : data;
        } catch (error) {
            console.error(`API Error: ${method} ${endpoint}`, error);
            
            // 인증 오류 시 자동 로그아웃
            if (error.status === 401 || error.status === 403) {
                Auth.logout();
                Utils.showToast('세션이 만료되어 로그아웃됩니다', 'warning');
            } else {
                Utils.showToast(error.message, 'error');
            }
            throw error;
        }
    }
};

// ===============================
// 인증 관리
// ===============================
const Auth = {
    async login(username, password) {
        try {
            Utils.setLoading(true, '로그인 중...');
            const data = await API.request('/login', 'POST', { username, password });
            
            AppState.jwtToken = data.token;
            localStorage.setItem('jwtToken', AppState.jwtToken);
            
            // 로그인 후 즉시 토큰 검증으로 사용자 정보 가져오기
            const isValid = await this.validateToken();
            
            if (isValid) {
                console.log('✅ 로그인 성공, 사용자 정보:', AppState.currentUser);
                Utils.showToast(`환영합니다, ${AppState.currentUser.username}님!`, 'success');
                UI.updateAuthState();
                return true;
            } else {
                throw new Error('사용자 정보를 가져올 수 없습니다');
            }
        } catch (error) {
            console.error('로그인 실패:', error);
            this.clearTokens();
            return false;
        } finally {
            Utils.setLoading(false);
        }
    },

    async register(username, password) {
        try {
            Utils.setLoading(true, '회원가입 중...');
            await API.request('/register', 'POST', { username, password });
            Utils.showToast('회원가입이 완료되었습니다!', 'success');
            UI.showLogin();
            return true;
        } catch (error) {
            return false;
        } finally {
            Utils.setLoading(false);
        }
    },

    async validateToken() {
        if (!AppState.jwtToken) return false;
        
        try {
            const data = await API.request('/validate');
            if (data.valid) {
                // 기존 사용자 정보 업데이트 (역할 정보 포함)
                AppState.currentUser = {
                    id: data.user_id,
                    username: data.username,
                    role: data.role
                };
                
                console.log('🔍 토큰 검증 성공, 사용자 정보 업데이트:', AppState.currentUser);
                
                // 네비게이션 업데이트 (중요!)
                if (typeof UI !== 'undefined') {
                    UI.updateNavigation();
                }
                
                return true;
            }
        } catch (error) {
            console.error('토큰 검증 실패:', error);
            this.clearTokens();
        }
        return false;
    },

    logout() {
        const username = AppState.currentUser?.username || '사용자';
        console.log('🚪 로그아웃:', username);
        
        this.clearTokens();
        UI.updateAuthState();
        
        // 네비게이션 상태도 리셋
        const usersNav = document.getElementById('nav-users');
        if (usersNav) {
            usersNav.style.display = 'none';
        }
        
        Utils.showToast(`${username}님, 안전하게 로그아웃되었습니다`, 'info');
    },

    clearTokens() {
        AppState.jwtToken = null;
        AppState.currentUser = null;
        localStorage.removeItem('jwtToken');
    }
};

// ===============================
// UI 관리
// ===============================
const UI = {
    elements: {},

    init() {
        // DOM 요소 캐싱
        this.elements = {
            authSection: document.getElementById('auth-section'),
            keySection: document.getElementById('key-section'),
            loginView: document.getElementById('login-view'),
            registerView: document.getElementById('register-view'),
            keysView: document.getElementById('keys-view'),
            usersView: document.getElementById('users-view'),
            profileView: document.getElementById('profile-view'),
            usersList: document.getElementById('users-list'),
            keyDisplayArea: document.getElementById('key-display-area'),
            keyInfo: document.getElementById('key-info'),
            container: document.querySelector('.container'),
            modal: document.getElementById('user-detail-modal'),
            modalContent: document.getElementById('user-detail-content')
        };

        this.setupEventListeners();
    },

    setupEventListeners() {
        // 통합 이벤트 위임
        document.addEventListener('click', this.handleClick.bind(this));
        document.addEventListener('submit', this.handleSubmit.bind(this));
        document.addEventListener('keydown', this.handleKeydown.bind(this));
    },

    handleClick(e) {
        const target = e.target;
        const action = target.dataset.action;
        
        // data-action 기반 통합 처리
        if (action) {
            e.preventDefault();
            this.executeAction(action, target);
            return;
        }

        // 기존 ID/클래스 기반 처리
        switch (target.id) {
            case 'show-register':
                e.preventDefault();
                this.showRegister();
                break;
            case 'show-login':
                e.preventDefault();
                this.showLogin();
                break;
            case 'logout-btn':
                Auth.logout();
                break;
            case 'nav-keys':
                e.preventDefault();
                this.showView('keys');
                break;
            case 'nav-users':
                e.preventDefault();
                this.showView('users');
                break;
            case 'nav-profile':
                e.preventDefault();
                this.showView('profile');
                break;
        }

        // 복사 버튼 처리
        if (target.classList.contains('copy-btn')) {
            e.preventDefault();
            this.copyToClipboard(target);
        }

        // 모달 닫기
        if (target.classList.contains('close') || target === this.elements.modal) {
            this.closeModal();
        }
    },

    async handleSubmit(e) {
        e.preventDefault();
        const form = e.target;
        const formData = new FormData(form);

        switch (form.id) {
            case 'login-form':
                const loginSuccess = await Auth.login(
                    formData.get('username'),
                    formData.get('password')
                );
                if (loginSuccess) form.reset();
                break;

            case 'register-form':
                const registerSuccess = await Auth.register(
                    formData.get('username'),
                    formData.get('password')
                );
                if (registerSuccess) {
                    form.reset();
                    // 자동으로 사용자명 입력
                    this.elements.loginView.querySelector('[name="username"]').value = formData.get('username');
                }
                break;

            case 'profile-form':
                await this.updateProfile(formData);
                break;
        }
    },

    handleKeydown(e) {
        // ESC로 모달 닫기
        if (e.key === 'Escape' && this.isModalOpen()) {
            this.closeModal();
        }
    },

    // 통합 액션 실행기
    async executeAction(action, element) {
        const [actionType, actionData] = action.split(':');
        
        try {
            switch (actionType) {
                case 'key-create':
                    await KeyManager.create();
                    break;
                case 'key-view':
                    await KeyManager.view();
                    break;
                case 'key-delete':
                    await KeyManager.delete();
                    break;
                case 'user-detail':
                    await this.showUserDetail(parseInt(actionData));
                    break;
                case 'user-delete':
                    await this.deleteUser(parseInt(actionData), element.dataset.username);
                    break;
                case 'copy':
                    this.copyToClipboard(element);
                    break;
                case 'toggle-section':
                    this.toggleSection(actionData);
                    break;
                default:
                    console.warn('알 수 없는 액션:', action);
            }
        } catch (error) {
            console.error('액션 실행 오류:', error);
        }
    },

    showView(viewName) {
        // 권한 확인
        if (viewName === 'users' && !this.isAdmin()) {
            Utils.showToast('관리자만 접근 가능합니다', 'warning');
            return;
        }

        // 모든 뷰 숨기기
        ['keys', 'users', 'profile'].forEach(view => {
            this.elements[`${view}View`]?.classList.add('hidden');
            document.getElementById(`nav-${view}`)?.classList.remove('active');
        });

        // 선택된 뷰 표시
        this.elements[`${viewName}View`]?.classList.remove('hidden');
        document.getElementById(`nav-${viewName}`)?.classList.add('active');

        AppState.currentView = viewName;

        // 뷰별 초기화
        switch (viewName) {
            case 'keys':
                KeyManager.autoLoadKeys(); // 키 자동 로드 추가
                break;
            case 'users':
                UserManager.loadList();
                break;
            case 'profile':
                ProfileManager.loadProfile();
                break;
        }
    },

    updateAuthState() {
        const isLoggedIn = !!AppState.jwtToken && !!AppState.currentUser;
        
        this.elements.authSection.classList.toggle('hidden', isLoggedIn);
        this.elements.keySection.classList.toggle('hidden', !isLoggedIn);
        this.elements.container.classList.toggle('container-wide', isLoggedIn);

        if (isLoggedIn) {
            console.log('🔄 네비게이션 업데이트:', AppState.currentUser);
            this.updateNavigation(); // 로그인 시 네비게이션 업데이트
            this.showView('keys'); // 키 뷰로 가면서 자동 로드됨
        } else {
            this.showLogin();
        }
    },

    updateNavigation() {
        const isAdmin = this.isAdmin();
        const usersNav = document.getElementById('nav-users');
        
        console.log(`👤 사용자 권한: ${AppState.currentUser?.role || 'unknown'}, 관리자: ${isAdmin}`);
        
        if (usersNav) {
            if (isAdmin) {
                usersNav.style.display = 'inline-block';
                usersNav.title = '사용자 목록 관리';
                console.log('✅ 사용자 관리 버튼 표시');
            } else {
                usersNav.style.display = 'none';
                console.log('❌ 사용자 관리 버튼 숨김');
                
                // 현재 사용자 뷰에 있으면 키 관리로 이동
                if (AppState.currentView === 'users') {
                    console.log('🔄 사용자 뷰에서 키 뷰로 이동');
                    this.showView('keys');
                }
            }
        }
    },

    isAdmin() {
        return AppState.currentUser?.role === 'admin';
    },

    showLogin() {
        this.elements.loginView.classList.remove('hidden');
        this.elements.registerView.classList.add('hidden');
    },

    showRegister() {
        this.elements.loginView.classList.add('hidden');
        this.elements.registerView.classList.remove('hidden');
    },

    // 모달 관리 (간소화)
    showModal(content) {
        this.elements.modalContent.innerHTML = content;
        this.elements.modal.style.display = 'block';
        document.body.style.overflow = 'hidden';
        
        // 첫 번째 포커스 가능 요소에 포커스
        setTimeout(() => {
            const firstFocusable = this.elements.modal.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
            if (firstFocusable) firstFocusable.focus();
        }, 100);
    },

    closeModal() {
        this.elements.modal.style.display = 'none';
        document.body.style.overflow = '';
        this.elements.modalContent.innerHTML = '';
        
        // 중요: 포커스를 body로 이동 후 즉시 해제하여 키보드 이벤트 정상화
        setTimeout(() => {
            document.body.focus();
            document.body.blur();
        }, 50);
    },

    isModalOpen() {
        return this.elements.modal.style.display === 'block';
    },

    async showUserDetail(userId) {
        try {
            Utils.setLoading(true, '사용자 정보 로딩 중...');
            const user = await API.request(`/admin/users/${userId}`);
            
            const content = this.createUserDetailContent(user);
            this.showModal(content);
        } catch (error) {
            Utils.showToast('사용자 정보를 불러올 수 없습니다', 'error');
        } finally {
            Utils.setLoading(false);
        }
    },

    createUserDetailContent(user) {
        const isAdmin = user.role === 'admin';
        const hasKey = user.has_ssh_key;
        
        return `
            <div class="user-detail">
                <div class="user-header">
                    <h3>${Utils.escapeHtml(user.username)}</h3>
                    <span class="badge ${isAdmin ? 'admin' : 'user'}">
                        ${isAdmin ? '👑 관리자' : '👤 일반 사용자'}
                    </span>
                </div>
                
                <div class="user-info">
                    <div class="info-row">
                        <span>ID:</span> <span>${user.id}</span>
                    </div>
                    <div class="info-row">
                        <span>가입일:</span> <span>${Utils.formatDate(user.created_at)}</span>
                    </div>
                    <div class="info-row">
                        <span>SSH 키:</span> 
                        <span class="${hasKey ? 'has-key' : 'no-key'}">
                            ${hasKey ? '🔑 보유' : '🔓 없음'}
                        </span>
                    </div>
                </div>

                ${this.isAdmin() && !isAdmin ? `
                    <div class="admin-actions">
                        <h4>관리자 액션</h4>
                        <div class="action-buttons">
                            <button class="btn-warning" data-action="user-role-change:${user.id}" 
                                    data-username="${Utils.escapeHtml(user.username)}" 
                                    data-current-role="${user.role}">
                                권한 변경
                            </button>
                            <button class="btn-danger" data-action="user-delete:${user.id}" 
                                    data-username="${Utils.escapeHtml(user.username)}">
                                사용자 삭제
                            </button>
                        </div>
                    </div>
                ` : ''}
            </div>
        `;
    },

    async deleteUser(userId, username) {
        if (!confirm(`정말로 사용자 "${username}"를 삭제하시겠습니까?`)) {
            return;
        }

        try {
            Utils.setLoading(true, '사용자 삭제 중...');
            await API.request(`/admin/users/${userId}`, 'DELETE');
            Utils.showToast(`사용자 "${username}"가 삭제되었습니다`, 'success');
            this.closeModal();
            UserManager.loadList();
        } catch (error) {
            Utils.showToast('사용자 삭제에 실패했습니다', 'error');
        } finally {
            Utils.setLoading(false);
        }
    },

    async copyToClipboard(element) {
        const targetId = element.dataset.target;
        const text = element.dataset.text || document.getElementById(targetId)?.textContent;
        const type = element.dataset.type || '텍스트';

        if (!text) {
            Utils.showToast('복사할 내용이 없습니다', 'warning');
            return;
        }

        try {
            if (navigator.clipboard) {
                await navigator.clipboard.writeText(text);
            } else {
                // 폴백 방식
                const textarea = document.createElement('textarea');
                textarea.value = text;
                textarea.style.position = 'fixed';
                textarea.style.opacity = '0';
                document.body.appendChild(textarea);
                textarea.select();
                document.execCommand('copy');
                document.body.removeChild(textarea);
            }
            Utils.showToast(`${type}가 클립보드에 복사되었습니다`, 'success');
        } catch (error) {
            Utils.showToast('복사에 실패했습니다', 'error');
        }
    },

    toggleSection(targetSelector) {
        const element = document.querySelector(targetSelector);
        if (element) {
            element.classList.toggle('hidden');
        }
    }
};

// ===============================
// 키 관리
// ===============================
const KeyManager = {
    // 로그인 후 키 자동 로드
    async autoLoadKeys() {
        console.log('🔍 키 자동 로드 시도');
        try {
            await this.view();
        } catch (error) {
            // 키가 없는 경우는 정상적인 상황이므로 에러 로그만 출력
            console.log('ℹ️ 키가 없거나 로드 실패:', error.message);
            UI.elements.keyInfo.textContent = '생성된 SSH 키가 없습니다. 키를 생성해주세요.';
        }
    },

    async create() {
        if (!confirm('SSH 키를 생성하시겠습니까?\n기존 키가 있다면 새 키로 교체됩니다.')) {
            return;
        }

        try {
            Utils.setLoading(true, '키 생성 중...');
            const keyData = await API.request('/keys', 'POST');
            this.displayKeys(keyData);
            Utils.showToast('SSH 키가 성공적으로 생성되었습니다!', 'success');
        } catch (error) {
            this.hideKeys();
        } finally {
            Utils.setLoading(false);
        }
    },

    async view() {
        try {
            Utils.setLoading(true, '키 조회 중...');
            const keyData = await API.request('/keys');
            this.displayKeys(keyData);
            console.log('✅ 키 로드 성공');
        } catch (error) {
            this.hideKeys();
            if (error.message.includes('키를 찾을 수 없습니다') || error.status === 404) {
                UI.elements.keyInfo.textContent = '생성된 SSH 키가 없습니다. 먼저 키를 생성해주세요.';
                console.log('ℹ️ 키가 없음 - 정상 상황');
            } else {
                console.error('❌ 키 로드 실패:', error.message);
                UI.elements.keyInfo.textContent = '키를 불러오는 중 오류가 발생했습니다.';
            }
            throw error; // autoLoadKeys에서 catch할 수 있도록
        } finally {
            Utils.setLoading(false);
        }
    },

    async delete() {
        if (!confirm('정말로 SSH 키를 삭제하시겠습니까?\n이 작업은 되돌릴 수 없습니다.')) {
            return;
        }

        try {
            Utils.setLoading(true, '키 삭제 중...');
            await API.request('/keys', 'DELETE');
            this.hideKeys();
            Utils.showToast('SSH 키가 성공적으로 삭제되었습니다', 'success');
        } catch (error) {
            // 에러는 API에서 이미 표시됨
        } finally {
            Utils.setLoading(false);
        }
    },

    displayKeys(keyData) {
        console.log('📋 키 데이터 수신:', keyData); // 디버깅용
        
        // API 응답 구조 정규화
        const normalizedData = this.normalizeKeyData(keyData);
        
        UI.elements.keyInfo.textContent = `Algorithm: ${normalizedData.algorithm} / Bits: ${normalizedData.bits}`;
        
        // 키 데이터 설정
        const keyTypes = {
            'public': normalizedData.publicKey,
            'pem': normalizedData.privateKeyPem,
            'ppk': normalizedData.privateKeyPpk
        };
        
        Object.entries(keyTypes).forEach(([type, keyContent]) => {
            const keyElement = document.getElementById(`key-${type}`);
            const cmdElement = document.getElementById(`cmd-${type}`);
            
            if (keyElement && keyContent) {
                keyElement.textContent = keyContent;
            }
            
            if (cmdElement && keyContent) {
                cmdElement.textContent = this.generateCommand(type, keyContent);
            }
        });

        // authorized_keys 명령어 별도 처리
        const cmdAuthKeys = document.getElementById('cmd-authorized-keys');
        if (cmdAuthKeys && normalizedData.publicKey) {
            cmdAuthKeys.textContent = `echo '${normalizedData.publicKey}' >> ~/.ssh/authorized_keys`;
        }

        UI.elements.keyDisplayArea.classList.remove('hidden');
    },

    // API 응답 데이터 정규화
    normalizeKeyData(keyData) {
        console.log('🔧 정규화 전 데이터:', keyData);
        
        // 다양한 API 응답 형태를 표준화
        const normalized = {
            algorithm: keyData.Algorithm || keyData.algorithm || 'RSA',
            bits: keyData.Bits || keyData.bits || keyData.key_size || 2048,
            publicKey: keyData.PublicKey || keyData.public_key || keyData.publicKey || keyData.pub || '',
            privateKeyPem: keyData.PEM || keyData.private_key_pem || keyData.privateKeyPem || keyData.pem || '',
            privateKeyPpk: keyData.PPK || keyData.private_key_ppk || keyData.privateKeyPpk || keyData.ppk || ''
        };
        
        console.log('🔧 정규화 후 데이터:', normalized);
        
        // 빈 값 체크
        if (!normalized.publicKey) {
            console.warn('⚠️ 공개키가 비어있습니다');
        }
        if (!normalized.privateKeyPem) {
            console.warn('⚠️ PEM 개인키가 비어있습니다');
        }
        if (!normalized.privateKeyPpk) {
            console.warn('⚠️ PPK 개인키가 비어있습니다');
        }
        
        return normalized;
    },

    generateCommand(type, keyContent) {
        if (!keyContent) return '';
        
        // 안전한 문자열 처리를 위한 이스케이프
        const escapeShell = (str) => {
            return str.replace(/'/g, "'\"'\"'");
        };

        const commands = {
            public: `echo '${escapeShell(keyContent)}' > id_rsa.pub`,
            pem: `echo '${escapeShell(keyContent)}' > id_rsa`,
            ppk: `echo '${escapeShell(keyContent)}' > id_rsa.ppk`
        };
        return commands[type] || '';
    },

    hideKeys() {
        UI.elements.keyDisplayArea.classList.add('hidden');
        UI.elements.keyInfo.textContent = '';
    }
};

// ===============================
// 사용자 관리
// ===============================
const UserManager = {
    async loadList() {
        if (!UI.isAdmin()) {
            this.showAdminOnlyMessage();
            return;
        }

        try {
            this.setLoadingState(true);
            const data = await API.request('/admin/users-list');
            const users = data.items || data;
            this.displayUsers(users);
            this.updateStats(users);
        } catch (error) {
            this.showError('사용자 목록을 불러올 수 없습니다');
        } finally {
            this.setLoadingState(false);
        }
    },

    displayUsers(users) {
        if (!users.length) {
            UI.elements.usersList.innerHTML = '<div class="empty-state">등록된 사용자가 없습니다</div>';
            return;
        }

        const html = users.map(user => this.createUserCard(user)).join('');
        UI.elements.usersList.innerHTML = html;
    },

    createUserCard(user) {
        const isAdmin = user.role === 'admin';
        const hasKey = user.has_ssh_key;
        
        return `
            <div class="user-card ${isAdmin ? 'admin-user' : ''}">
                <div class="user-info">
                    <div class="user-name">
                        ${Utils.escapeHtml(user.username)}
                        ${isAdmin ? '<span class="admin-badge">관리자</span>' : ''}
                    </div>
                    <div class="user-meta">
                        ID: ${user.id} • 가입: ${Utils.formatDate(user.created_at)}
                    </div>
                    <div class="user-status ${hasKey ? 'has-key' : 'no-key'}">
                        ${hasKey ? '🔑 키 보유' : '🔓 키 없음'}
                    </div>
                </div>
                <div class="user-actions">
                    <button class="btn-secondary btn-sm" data-action="user-detail:${user.id}">
                        상세 보기
                    </button>
                    ${UI.isAdmin() && !isAdmin ? `
                        <button class="btn-danger btn-sm" data-action="user-delete:${user.id}" 
                                data-username="${Utils.escapeHtml(user.username)}">
                            삭제
                        </button>
                    ` : ''}
                </div>
            </div>
        `;
    },

    updateStats(users) {
        const total = users.length;
        const withKeys = users.filter(u => u.has_ssh_key).length;
        
        const totalElement = document.getElementById('total-users');
        const withKeysElement = document.getElementById('users-with-keys');
        
        if (totalElement) totalElement.textContent = total;
        if (withKeysElement) withKeysElement.textContent = withKeys;
    },

    showAdminOnlyMessage() {
        UI.elements.usersList.innerHTML = `
            <div class="admin-only-message">
                <div class="icon">🔒</div>
                <h3>관리자 전용 기능</h3>
                <p>사용자 목록은 관리자만 조회할 수 있습니다.</p>
            </div>
        `;
    },

    showError(message) {
        UI.elements.usersList.innerHTML = `
            <div class="error-message">
                <div class="icon">❌</div>
                <h3>오류 발생</h3>
                <p>${message}</p>
                <button data-action="reload-users">다시 시도</button>
            </div>
        `;
    },

    setLoadingState(isLoading) {
        if (isLoading) {
            UI.elements.usersList.innerHTML = `
                <div class="loading-state">
                    <div>사용자 목록을 불러오는 중...</div>
                </div>
            `;
        }
    }
};

// ===============================
// 프로필 관리
// ===============================
const ProfileManager = {
    async loadProfile() {
        try {
            const user = await API.request('/users/me');
            this.displayProfile(user);
        } catch (error) {
            Utils.showToast('프로필 정보를 불러올 수 없습니다', 'error');
        }
    },

    displayProfile(user) {
        const profileInfo = document.getElementById('current-user-info');
        if (!profileInfo) return;

        profileInfo.innerHTML = `
            <div class="profile-card">
                <h3>${Utils.escapeHtml(user.username)}</h3>
                <div class="profile-details">
                    <div><strong>ID:</strong> ${user.id}</div>
                    <div><strong>역할:</strong> ${user.role === 'admin' ? '관리자' : '일반 사용자'}</div>
                    <div><strong>가입일:</strong> ${Utils.formatDate(user.created_at)}</div>
                    <div><strong>SSH 키:</strong> ${user.has_ssh_key ? '보유' : '없음'}</div>
                </div>
            </div>
        `;
    },

    async updateProfile(formData) {
        const updateData = {};
        const username = formData.get('username')?.trim();
        const password = formData.get('password')?.trim();

        if (username) updateData.username = username;
        if (password) updateData.new_password = password;

        if (Object.keys(updateData).length === 0) {
            Utils.showToast('변경할 내용이 없습니다', 'info');
            return;
        }

        try {
            Utils.setLoading(true, '프로필 업데이트 중...');
            await API.request('/users/me', 'PUT', updateData);
            
            if (updateData.username) {
                AppState.currentUser.username = updateData.username;
            }
            
            Utils.showToast('프로필이 성공적으로 업데이트되었습니다', 'success');
            this.loadProfile();
        } catch (error) {
            Utils.showToast('프로필 업데이트에 실패했습니다', 'error');
        } finally {
            Utils.setLoading(false);
        }
    }
};

// ===============================
// 애플리케이션 초기화
// ===============================
const App = {
    async init() {
        console.log('🚀 SSH Key Manager 시작');
        
        // UI 초기화
        UI.init();
        
        // 자동 로그인 확인
        if (AppState.jwtToken) {
            console.log('🔐 저장된 토큰으로 자동 로그인 시도');
            const isValid = await Auth.validateToken();
            if (isValid) {
                console.log('✅ 자동 로그인 성공');
                Utils.showToast(`안녕하세요, ${AppState.currentUser.username}님!`, 'success');
            } else {
                console.log('❌ 자동 로그인 실패');
            }
        }
        
        // UI 상태 업데이트 (네비게이션 포함)
        UI.updateAuthState();
        
        console.log('✅ 애플리케이션 초기화 완료');
    }
};

// DOM이 로드되면 초기화
document.addEventListener('DOMContentLoaded', () => {
    // 앱 초기화
    App.init();
});

// 전역 객체로 노출 (디버깅용)
window.AppState = AppState;
window.Auth = Auth;
window.API = API;
window.Utils = Utils;
window.UI = UI;
window.KeyManager = KeyManager;
window.UserManager = UserManager;
window.ProfileManager = ProfileManager;