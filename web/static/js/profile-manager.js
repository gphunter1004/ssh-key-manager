// 프로필 관리자 - 완전 독립 버전
window.ProfileManager = {
    currentProfile: null,
    isLoading: false,

    init: function() {
        console.log('👤 ProfileManager 초기화 시작');
        this.setupEventListeners();
        console.log('✅ ProfileManager 초기화 완료');
        return true;
    },

    setupEventListeners: function() {
        console.log('🎯 ProfileManager 이벤트 리스너 설정');
        
        if (DOM.profileForm) {
            DOM.profileForm.onsubmit = (e) => this.handleProfileUpdate(e);
            console.log('✅ 프로필 폼 이벤트 설정');
        }
        
        const usernameInput = document.querySelector('#profile-form input[name="username"]');
        const passwordInput = document.querySelector('#profile-form input[name="password"]');
        
        if (usernameInput) {
            usernameInput.addEventListener('input', (e) => this.validateUsername(e));
            usernameInput.addEventListener('blur', (e) => this.checkUsernameAvailability(e));
            console.log('✅ 사용자명 입력 검증 이벤트 설정');
        }
        
        if (passwordInput) {
            passwordInput.addEventListener('input', (e) => this.validatePassword(e));
            console.log('✅ 비밀번호 입력 검증 이벤트 설정');
        }
        
        console.log('✅ ProfileManager 모든 이벤트 설정 완료');
    },

    loadCurrentUserProfile: async function() {
        if (this.isLoading) {
            console.log('ℹ️ 프로필 로딩 중, 중복 요청 무시');
            return;
        }
        
        console.log('👤 현재 사용자 프로필 로드 시작');
        this.isLoading = true;
        
        try {
            Utils.setLoading(true, '프로필 정보를 불러오는 중...');
            
            const userData = await AppUtils.apiFetch('/users/me', 'GET');
            
            this.currentProfile = userData;
            AppState.currentUser = {
                ...AppState.currentUser,
                ...userData
            };
            
            console.log('✅ 프로필 로드 성공:', userData.username);
            
            this.displayCurrentUserInfo(userData);
            this.populateForm(userData);
            
        } catch (error) {
            console.error('❌ 프로필 로드 실패:', error.message);
            this.showProfileError('프로필 정보를 불러올 수 없습니다: ' + error.message);
        } finally {
            this.isLoading = false;
            Utils.setLoading(false);
        }
    },

    displayCurrentUserInfo: function(userData) {
        console.log('📋 프로필 정보 표시');
        
        if (!DOM.currentUserInfo) {
            console.warn('⚠️ current-user-info 요소를 찾을 수 없음');
            return;
        }

        const joinedDate = userData.created_at ? Utils.formatDate(userData.created_at) : '정보 없음';
        const lastUpdate = userData.updated_at ? Utils.formatDate(userData.updated_at) : '정보 없음';

        DOM.currentUserInfo.innerHTML = `
            <div class="profile-card">
                <div class="profile-header">
                    <h3>현재 프로필 정보</h3>
                </div>
                <div class="profile-table-container">
                    <div class="profile-table-row">
                        <div class="profile-table-label">사용자명:</div>
                        <div class="profile-table-value">${Utils.escapeHtml(userData.username)}</div>
                    </div>
                    <div class="profile-table-row">
                        <div class="profile-table-label">사용자 ID:</div>
                        <div class="profile-table-value">${userData.id}</div>
                    </div>
                    <div class="profile-table-row">
                        <div class="profile-table-label">역할:</div>
                        <div class="profile-table-value">${userData.role === 'admin' ? '관리자' : '일반 사용자'}</div>
                    </div>
                    <div class="profile-table-row">
                        <div class="profile-table-label">가입일:</div>
                        <div class="profile-table-value">${joinedDate}</div>
                    </div>
                    <div class="profile-table-row">
                        <div class="profile-table-label">마지막 업데이트:</div>
                        <div class="profile-table-value">${lastUpdate}</div>
                    </div>
                    <div class="profile-table-row">
                        <div class="profile-table-label">SSH 키 상태:</div>
                        <div class="profile-table-value">
                            <span class="key-status ${userData.has_ssh_key ? 'has-key' : 'no-key'}">
                                ${userData.has_ssh_key ? '🔑 보유' : '❌ 없음'}
                            </span>
                        </div>
                    </div>
                </div>
                <div class="profile-actions-wrapper">
                    <button type="button" class="btn-secondary btn-sm" onclick="ProfileManager.refreshProfile()">
                        🔄 새로고침
                    </button>
                    <button type="button" class="btn-primary btn-sm" onclick="ProfileManager.checkKeyStatus()">
                        🔑 키 상태 확인
                    </button>
                </div>
            </div>
        `;
    },

    populateForm: function(userData) {
        console.log('📝 프로필 폼 데이터 설정');
        
        const usernameInput = document.querySelector('#profile-form input[name="username"]');
        const passwordInput = document.querySelector('#profile-form input[name="password"]');
        
        if (usernameInput) {
            usernameInput.value = userData.username;
            usernameInput.dataset.originalValue = userData.username;
            usernameInput.placeholder = `현재: ${userData.username}`;
        }
        
        if (passwordInput) {
            passwordInput.value = '';
            passwordInput.placeholder = '새 비밀번호 (변경하지 않으려면 빈 상태로 두세요)';
        }
        
        this.clearFormValidation();
        console.log('✅ 프로필 폼 데이터 설정 완료');
    },

    handleProfileUpdate: async function(e) {
        e.preventDefault();
        
        console.log('💾 프로필 업데이트 요청');
        
        const formData = new FormData(e.target);
        const newUsername = formData.get('username')?.trim();
        const newPassword = formData.get('password')?.trim();
        const originalUsername = e.target.querySelector('input[name="username"]')?.dataset.originalValue;
        
        const updateData = {};
        let hasChanges = false;
        
        if (newUsername && newUsername !== originalUsername) {
            updateData.username = newUsername;
            hasChanges = true;
        }
        
        if (newPassword && newPassword !== '') {
            updateData.new_password = newPassword;
            hasChanges = true;
        }
        
        if (!hasChanges) {
            Utils.showToast('변경할 내용이 없습니다.', 'info');
            return;
        }
        
        const confirmUpdate = this.getUpdateConfirmation(updateData);
        if (!confirm(confirmUpdate)) {
            console.log('ℹ️ 프로필 업데이트 취소됨');
            return;
        }
        
        try {
            Utils.setLoading(true, '프로필 업데이트 중...');
            
            const result = await AppUtils.apiFetch('/users/me', 'PUT', updateData);
            
            console.log('✅ 프로필 업데이트 성공:', result);
            
            Utils.showToast('프로필이 성공적으로 업데이트되었습니다!', 'success');
            
            if (updateData.username) {
                AppState.currentUser.username = updateData.username;
            }
            
            const passwordInput = e.target.querySelector('input[name="password"]');
            if (passwordInput) {
                passwordInput.value = '';
            }
            
            await this.loadCurrentUserProfile();
            
        } catch (error) {
            console.error('❌ 프로필 업데이트 실패:', error.message);
            Utils.showToast('프로필 업데이트 실패: ' + error.message, 'error');
        } finally {
            Utils.setLoading(false);
        }
    },

    getUpdateConfirmation: function(updateData) {
        let message = '다음 정보를 업데이트하시겠습니까?\n\n';
        
        if (updateData.username) {
            message += `• 사용자명: ${updateData.username}\n`;
        }
        
        if (updateData.new_password) {
            message += '• 비밀번호: 변경됨\n';
        }
        
        message += '\n⚠️ 사용자명을 변경하면 다시 로그인해야 할 수 있습니다.';
        
        return message;
    },

    validateUsername: function(e) {
        const input = e.target;
        const username = input.value.trim();
        const originalUsername = input.dataset.originalValue;
        
        let messageEl = input.parentNode.querySelector('.validation-message');
        if (!messageEl) {
            messageEl = document.createElement('div');
            messageEl.className = 'validation-message';
            input.parentNode.appendChild(messageEl);
        }
        
        if (username === '') {
            this.setValidationMessage(messageEl, '', 'none');
            return;
        }
        
        if (username === originalUsername) {
            this.setValidationMessage(messageEl, '현재 사용자명과 동일합니다', 'info');
            return;
        }
        
        if (username.length < 2) {
            this.setValidationMessage(messageEl, '사용자명은 최소 2자 이상이어야 합니다', 'error');
            return;
        }
        
        if (username.length > 30) {
            this.setValidationMessage(messageEl, '사용자명은 최대 30자까지 가능합니다', 'error');
            return;
        }
        
        if (!/^[a-zA-Z0-9_-]+$/.test(username)) {
            this.setValidationMessage(messageEl, '영문, 숫자, -, _ 만 사용 가능합니다', 'error');
            return;
        }
        
        this.setValidationMessage(messageEl, '사용 가능한 사용자명입니다', 'success');
    },

    validatePassword: function(e) {
        const input = e.target;
        const password = input.value;
        
        let messageEl = input.parentNode.querySelector('.validation-message');
        if (!messageEl) {
            messageEl = document.createElement('div');
            messageEl.className = 'validation-message';
            input.parentNode.appendChild(messageEl);
        }
        
        if (password === '') {
            this.setValidationMessage(messageEl, '비밀번호를 변경하지 않으려면 빈 상태로 두세요', 'info');
            return;
        }
        
        if (password.length < 4) {
            this.setValidationMessage(messageEl, '비밀번호는 최소 4자 이상이어야 합니다', 'error');
            return;
        }
        
        if (password.length > 100) {
            this.setValidationMessage(messageEl, '비밀번호가 너무 깁니다', 'error');
            return;
        }
        
        const strength = this.checkPasswordStrength(password);
        this.setValidationMessage(messageEl, `비밀번호 강도: ${strength.text}`, strength.type);
    },

    checkPasswordStrength: function(password) {
        let score = 0;
        const checks = {
            length: password.length >= 8,
            lowercase: /[a-z]/.test(password),
            uppercase: /[A-Z]/.test(password),
            numbers: /\d/.test(password),
            symbols: /[^A-Za-z0-9]/.test(password)
        };
        
        score = Object.values(checks).filter(Boolean).length;
        
        if (score <= 2) return { text: '약함', type: 'error' };
        if (score <= 3) return { text: '보통', type: 'warning' };
        if (score <= 4) return { text: '강함', type: 'success' };
        return { text: '매우 강함', type: 'success' };
    },

    setValidationMessage: function(element, message, type) {
        element.textContent = message;
        element.className = `validation-message ${type}`;
        element.style.display = message ? 'block' : 'none';
        
        const colors = {
            'error': '#e74c3c',
            'warning': '#f39c12',
            'success': '#27ae60',
            'info': '#3498db',
            'none': 'transparent'
        };
        element.style.color = colors[type] || '#666';
        element.style.fontSize = '12px';
        element.style.marginTop = '4px';
    },

    clearFormValidation: function() {
        if (DOM.profileForm) {
            const messages = DOM.profileForm.querySelectorAll('.validation-message');
            messages.forEach(msg => msg.remove());
        }
    },

    checkUsernameAvailability: async function(e) {
        const input = e.target;
        const username = input.value.trim();
        const originalUsername = input.dataset.originalValue;
        
        if (!username || username === originalUsername) return;
        
        if (username.length >= 2 && /^[a-zA-Z0-9_-]+$/.test(username)) {
            console.log('🔍 사용자명 가용성 확인:', username);
        }
    },

    showProfileError: function(message) {
        if (DOM.currentUserInfo) {
            DOM.currentUserInfo.innerHTML = `
                <div class="error-message">
                    ❌ ${Utils.escapeHtml(message)}
                    <button type="button" onclick="ProfileManager.loadCurrentUserProfile()" class="btn-primary" style="margin-top:10px;">
                        다시 시도
                    </button>
                </div>
            `;
        }
    },

    refreshProfile: async function() {
        console.log('🔄 프로필 새로고침');
        await this.loadCurrentUserProfile();
    },

    checkKeyStatus: function() {
        console.log('🔑 키 상태 확인 요청');
        
        if (typeof ViewManager !== 'undefined' && ViewManager.showView) {
            ViewManager.showView('keys');
            Utils.showToast('키 관리 화면으로 이동합니다', 'info');
        } else {
            Utils.showToast('키 관리 화면으로 이동할 수 없습니다', 'warning');
        }
    },

    getCurrentProfile: function() {
        return this.currentProfile;
    },

    getDebugInfo: function() {
        return {
            currentProfile: this.currentProfile,
            isLoading: this.isLoading,
            appStateUser: AppState.currentUser,
            formExists: !!DOM.profileForm,
            infoExists: !!DOM.currentUserInfo
        };
    }
};