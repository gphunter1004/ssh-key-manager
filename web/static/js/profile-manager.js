// 프로필 관리자
window.ProfileManager = {
    currentProfile: null,
    isLoading: false,

    setupEventListeners: function() {
        // 프로필 업데이트 폼 제출
        if (DOM.profileForm) {
            DOM.profileForm.addEventListener('submit', this.handleProfileUpdate.bind(this));
        }
        
        // 실시간 입력 검증
        const usernameInput = document.getElementById('profile-username');
        const passwordInput = document.getElementById('profile-password');
        
        if (usernameInput) {
            usernameInput.addEventListener('input', this.validateUsername.bind(this));
            usernameInput.addEventListener('blur', this.checkUsernameAvailability.bind(this));
        }
        
        if (passwordInput) {
            passwordInput.addEventListener('input', this.validatePassword.bind(this));
        }
        
        console.log('ProfileManager 이벤트 리스너 설정 완료');
    },

    loadCurrentUserProfile: async function() {
        if (this.isLoading) {
            console.log('프로필 로딩 중, 중복 요청 무시');
            return;
        }
        
        console.log('현재 사용자 프로필 로드 시작');
        this.isLoading = true;
        
        try {
            // 로딩 상태 표시
            this.setLoadingState('프로필 정보를 불러오는 중...');
            
            const userData = await AppUtils.apiFetch('/users/me', 'GET');
            
            this.currentProfile = userData;
            AppState.currentUser = userData;
            
            console.log('프로필 로드 성공:', userData.username);
            
            // 프로필 정보 표시
            this.displayCurrentUserInfo(userData);
            
            // 폼에 현재 값 설정
            this.populateForm(userData);
            
        } catch (error) {
            console.error('프로필 로드 실패:', error.message);
            this.showProfileError('프로필 정보를 불러올 수 없습니다.');
        } finally {
            this.isLoading = false;
            this.clearLoadingState();
        }
    },

    displayCurrentUserInfo: function(userData) {
        let sshKeyInfo = '';
        
        if (userData.has_ssh_key) {
            // SSH 키 정보가 있는 경우
            const keyCreated = userData.ssh_key ? 
                new Date(userData.ssh_key.created_at).toLocaleString('ko-KR') : 
                '정보 없음';
            const keyUpdated = userData.ssh_key ? 
                new Date(userData.ssh_key.updated_at).toLocaleString('ko-KR') : 
                '정보 없음';
            
            sshKeyInfo = `
                <div class="user-detail">
                    <strong>SSH 키:</strong> 
                    <span class="key-status has-key">보유</span>
                    (RSA 4096bits)
                </div>
                <div class="user-detail">
                    <strong>키 생성일:</strong> ${keyCreated}
                </div>
                <div class="user-detail">
                    <strong>키 수정일:</strong> ${keyUpdated}
                </div>
            `;
        } else {
            sshKeyInfo = `
                <div class="user-detail">
                    <strong>SSH 키:</strong> 
                    <span class="key-status no-key">없음</span>
                    <small> - 키 관리 탭에서 생성하세요</small>
                </div>
            `;
        }

        const joinedDate = new Date(userData.created_at).toLocaleString('ko-KR');
        const lastUpdate = new Date(userData.updated_at).toLocaleString('ko-KR');

        if (DOM.currentUserInfo) {
            DOM.currentUserInfo.innerHTML = `
                <h3>현재 프로필 정보</h3>
                <div class="profile-summary">
                    <div class="user-detail">
                        <strong>사용자명:</strong> 
                        <span class="username">${this.escapeHtml(userData.username)}</span>
                    </div>
                    <div class="user-detail">
                        <strong>사용자 ID:</strong> ${userData.id}
                    </div>
                    <div class="user-detail">
                        <strong>역할:</strong> ${userData.role === 'admin' ? '관리자' : '일반 사용자'}
                    </div>
                    <div class="user-detail">
                        <strong>가입일:</strong> ${joinedDate}
                    </div>
                    <div class="user-detail">
                        <strong>마지막 업데이트:</strong> ${lastUpdate}
                    </div>
                    ${sshKeyInfo}
                </div>
                <div class="profile-actions">
                    <button type="button" class="btn-secondary" onclick="ProfileManager.refreshProfile()">
                        🔄 새로고침
                    </button>
                    <button type="button" class="btn-secondary" onclick="KeyManager.refresh()">
                        🔑 키 상태 확인
                    </button>
                </div>
            `;
        }
    },

    populateForm: function(userData) {
        // 폼 필드에 현재 값 설정
        const usernameInput = document.querySelector('#profile-form input[name="username"]');
        const passwordInput = document.querySelector('#profile-form input[name="password"]');
        
        if (usernameInput) {
            usernameInput.value = userData.username;
            usernameInput.dataset.originalValue = userData.username;
            usernameInput.placeholder = `현재: ${userData.username}`;
        }
        
        if (passwordInput) {
            passwordInput.value = ''; // 비밀번호는 항상 빈 값으로 시작
            passwordInput.placeholder = '새 비밀번호 (변경하지 않으려면 빈 상태로 두세요)';
        }
        
        // 폼 검증 상태 초기화
        this.clearFormValidation();
    },

    handleProfileUpdate: async function(e) {
        e.preventDefault();
        
        console.log('프로필 업데이트 요청');
        
        const formData = new FormData(e.target);
        const newUsername = formData.get('username')?.trim();
        const newPassword = formData.get('password')?.trim();
        const originalUsername = e.target.querySelector('input[name="username"]').dataset.originalValue;
        
        // 변경사항 확인
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
        
        // 사용자 확인
        const confirmUpdate = this.getUpdateConfirmation(updateData);
        if (!confirm(confirmUpdate)) {
            console.log('프로필 업데이트 취소됨');
            return;
        }
        
        try {
            // 로딩 상태 표시
            this.setFormLoadingState(true);
            
            const result = await AppUtils.apiFetch('/users/me', 'PUT', updateData);
            
            console.log('프로필 업데이트 성공:', result);
            
            // 성공 메시지
            Utils.showToast('프로필이 성공적으로 업데이트되었습니다!', 'success');
            
            // 사용자명이 변경된 경우 AppState 업데이트
            if (updateData.username) {
                AppState.currentUser.username = updateData.username;
            }
            
            // 비밀번호 필드 클리어
            const passwordInput = e.target.querySelector('input[name="password"]');
            if (passwordInput) {
                passwordInput.value = '';
            }
            
            // 프로필 정보 다시 로드
            await this.loadCurrentUserProfile();
            
        } catch (error) {
            console.error('프로필 업데이트 실패:', error.message);
            // 에러는 이미 AppUtils.apiFetch에서 표시됨
        } finally {
            this.setFormLoadingState(false);
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
        
        // 검증 메시지 요소 찾기 또는 생성
        let messageEl = input.parentNode.querySelector('.validation-message');
        if (!messageEl) {
            messageEl = document.createElement('div');
            messageEl.className = 'validation-message';
            input.parentNode.appendChild(messageEl);
        }
        
        // 검증 로직
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
        
        // 검증 메시지 요소 찾기 또는 생성
        let messageEl = input.parentNode.querySelector('.validation-message');
        if (!messageEl) {
            messageEl = document.createElement('div');
            messageEl.className = 'validation-message';
            input.parentNode.appendChild(messageEl);
        }
        
        // 검증 로직
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
        
        // 비밀번호 강도 검사
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
        
        // 타입별 색상 설정
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
        
        // 기본 검증 통과한 경우에만 중복 확인
        if (username.length >= 2 && /^[a-zA-Z0-9_-]+$/.test(username)) {
            // 실제로는 서버에서 중복 확인 API가 있어야 하지만,
            // 현재는 업데이트 시점에 확인하므로 여기서는 생략
            console.log('사용자명 가용성 확인:', username);
        }
    },

    setLoadingState: function(message) {
        if (DOM.currentUserInfo) {
            DOM.currentUserInfo.innerHTML = `<div class="loading-message">${message}</div>`;
        }
    },

    clearLoadingState: function() {
        // loadCurrentUserProfile에서 자동으로 클리어됨
    },

    setFormLoadingState: function(isLoading) {
        if (!DOM.profileForm) return;
        
        const submitBtn = DOM.profileForm.querySelector('button[type="submit"]');
        const inputs = DOM.profileForm.querySelectorAll('input');
        
        if (isLoading) {
            if (submitBtn) {
                submitBtn.disabled = true;
                submitBtn.textContent = '업데이트 중...';
            }
            inputs.forEach(input => input.disabled = true);
        } else {
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.textContent = '프로필 업데이트';
            }
            inputs.forEach(input => input.disabled = false);
        }
    },

    showProfileError: function(message) {
        if (DOM.currentUserInfo) {
            DOM.currentUserInfo.innerHTML = `
                <div class="error-message">
                    ❌ ${message}
                    <button type="button" onclick="ProfileManager.loadCurrentUserProfile()" class="btn-primary" style="margin-top:10px;">
                        다시 시도
                    </button>
                </div>
            `;
        }
    },

    refreshProfile: async function() {
        console.log('프로필 새로고침');
        await this.loadCurrentUserProfile();
    },

    // HTML 이스케이프 유틸리티
    escapeHtml: function(unsafe) {
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }
};