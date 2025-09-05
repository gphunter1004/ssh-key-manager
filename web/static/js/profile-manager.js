window.ProfileManager = {
    init: function() {
        if (DOM.profileForm) DOM.profileForm.onsubmit = (e) => this.handleProfileUpdate(e);
    },
    loadCurrentUserProfile: async function() {
        Utils.setLoading(true, '프로필 정보 로딩 중...');
        try {
            const userData = await AppUtils.apiFetch('/users/me');
            this.displayCurrentUserInfo(userData);
            this.populateForm(userData);
        } catch (error) {
            if (DOM.currentUserInfo) DOM.currentUserInfo.innerHTML = '<div class="error-message">프로필 정보를 불러올 수 없습니다.</div>';
        } finally { Utils.setLoading(false); }
    },
    displayCurrentUserInfo: function(user) {
        if (!DOM.currentUserInfo) return;
        DOM.currentUserInfo.innerHTML = `
            <div class="profile-card">
                <h3>현재 프로필 정보</h3>
                <div class="profile-table-container">
                    <div class="profile-table-row"><span class="profile-table-label">사용자명</span><span class="profile-table-value">${Utils.escapeHtml(user.username)}</span></div>
                    <div class="profile-table-row"><span class="profile-table-label">역할</span><span class="profile-table-value">${user.role === 'admin' ? '관리자' : '일반 사용자'}</span></div>
                    <div class="profile-table-row"><span class="profile-table-label">가입일</span><span class="profile-table-value">${Utils.formatDate(user.created_at)}</span></div>
                </div>
            </div>`;
    },
    populateForm: function(user) {
        if (!DOM.profileForm) return;
        const usernameInput = DOM.profileForm.elements['username'];
        usernameInput.value = user.username;
        usernameInput.placeholder = `현재: ${user.username}`;
        DOM.profileForm.elements['password'].value = '';
    },
    handleProfileUpdate: async function(e) {
        e.preventDefault();
        const username = e.target.elements['username'].value.trim();
        const new_password = e.target.elements['password'].value;
        const updateData = { username };
        if (new_password) {
            if (new_password.length < 4) return Utils.showToast('새 비밀번호는 4자 이상이어야 합니다.', 'warning');
            updateData.new_password = new_password;
        }
        if (!confirm('프로필을 업데이트하시겠습니까?')) return;
        Utils.setLoading(true, '프로필 업데이트 중...');
        try {
            await AppUtils.apiFetch('/users/me', 'PUT', updateData);
            Utils.showToast('프로필이 성공적으로 업데이트되었습니다.', 'success');
            await this.loadCurrentUserProfile();
        } finally { Utils.setLoading(false); }
    }
};