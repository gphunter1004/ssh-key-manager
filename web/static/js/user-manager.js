window.UserManager = {
    users: [],
    init: function() {
        document.addEventListener('click', (e) => {
            const target = e.target;
            const card = target.closest('.user-card');
            const button = target.closest('button[data-action]');

            if (!card && !button) return;
            
            let action, userId, username;

            if (button) { // 버튼 클릭
                action = button.dataset.action;
                userId = button.closest('.user-card')?.dataset.userId || button.dataset.userId;
                username = button.dataset.username;
            } else if (card) { // 카드 자체 클릭
                action = 'user-detail';
                userId = card.dataset.userId;
            }
            
            if (!action || !userId) return;

            if (action === 'user-detail') this.showUserDetail(parseInt(userId));
            else if (action === 'user-delete') this.confirmDeleteUser(parseInt(userId), username);
            else if (action === 'reload-users') this.loadUsersList();
        });
    },
    loadUsersList: async function() {
        if (!AuthManager.isAdmin()) return this.showAdminOnlyMessage();
        Utils.setLoading(true, '사용자 목록 로딩 중...');
        try {
            const usersData = await AppUtils.apiFetch('/admin/users');
            this.users = usersData.map(u => ({...u, id: u.ID, created_at: u.CreatedAt}));
            this.displayUsersList(this.users);
            this.updateStats(this.users);
        } catch (error) {
            this.showError('사용자 목록을 불러올 수 없습니다.');
        } finally { Utils.setLoading(false); }
    },
    displayUsersList: function(users) {
        const listEl = DOM.usersList;
        if (!listEl) return;
        listEl.innerHTML = '';
        if (users.length === 0) {
            listEl.innerHTML = '<div class="empty-state">등록된 사용자가 없습니다.</div>';
            return;
        }
        users.forEach(user => listEl.appendChild(this.createUserCard(user)));
    },
    createUserCard: function(user) {
        const card = document.createElement('div');
        card.className = 'user-card';
        card.dataset.userId = user.id;
        const canDelete = AppState.currentUser.id !== user.id && user.role !== 'admin';

        card.innerHTML = `
            <div class="user-info">
                <span class="user-name">${Utils.escapeHtml(user.username)} ${user.role === 'admin' ? '<span class="admin-badge">Admin</span>' : ''}</span>
                <span class="user-meta">가입: ${Utils.formatDate(user.created_at, {month:'2-digit', day:'2-digit'})}</span>
            </div>
            <div class="user-actions">
                <button class="btn-secondary btn-sm" data-action="user-detail">상세</button>
                ${canDelete ? `<button class="btn-danger btn-sm" data-action="user-delete" data-username="${Utils.escapeHtml(user.username)}">삭제</button>` : ''}
            </div>`;
        return card;
    },
    showUserDetail: async function(userId) {
        ModalManager.openModalWithSpinner('user-detail-modal', '사용자 정보 로딩 중...');
        try {
            const userData = await AppUtils.apiFetch(`/admin/users/${userId}`);
            let keyData = null;
            if (userData.has_ssh_key) {
                try { keyData = this.isCurrentUser(userId) ? await AppUtils.apiFetch('/keys') : await AppUtils.apiFetch(`/admin/users/${userId}/keys`); } 
                catch (e) { console.warn("Could not fetch user's key"); }
            }
            const title = `${userData.username} 상세 정보`;
            const contentHTML = this.generateModalContent(userData, keyData);
            ModalManager.updateModalContent(title, contentHTML);
        } catch (error) {
            ModalManager.updateModalContent('오류', `<div class="error-message">정보 로딩 실패</div>`);
        }
    },
    generateModalContent: function(user, key) {
        const isSelf = this.isCurrentUser(user.id);
        let keyHTML = `<div class="empty-state" style="padding:1rem 0;">SSH 키 없음</div>`;
        if (key) {
            const keyContent = isSelf 
                ? `<h4 class="modal-key-title">개인키 (PEM)</h4><div class="key-display-wrapper"><pre id="m-pem">${Utils.escapeHtml(key.PEM)}</pre><button class="copy-btn" data-target="#m-pem">복사</button></div>`
                : `<p class="security-notice warning">🔒 타인의 개인키는 볼 수 없습니다.</p>`;
            keyHTML = `<h4 class="modal-key-title">공개키</h4><div class="key-display-wrapper"><pre id="m-pub">${Utils.escapeHtml(key.PublicKey)}</pre><button class="copy-btn" data-target="#m-pub">복사</button></div>${keyContent}`;
        }
        const actions = isSelf ? `<div class="modal-actions"><button class="btn-primary" onclick="ViewManager.showView('profile'); ModalManager.closeActiveModal();">프로필 편집</button></div>` : '';
        return `<div class="user-detail-container">
                    <div><h3 class="modal-subheader">기본 정보</h3><div class="info-grid">
                        <div class="info-item"><strong>사용자명:</strong><span>${Utils.escapeHtml(user.username)} ${user.role === 'admin' ? '<span class="admin-badge">Admin</span>':''}</span></div>
                        <div class="info-item"><strong>가입일:</strong><span>${Utils.formatDate(user.created_at)}</span></div>
                    </div></div>
                    <div><h3 class="modal-subheader">SSH 키 정보</h3>${keyHTML}</div>
                    ${actions}
                </div>`;
    },
    confirmDeleteUser: async function(userId, username) {
        if (!confirm(`사용자 "${username}"를 정말로 삭제하시겠습니까?`)) return;
        Utils.setLoading(true, '사용자 삭제 중...');
        try {
            await AppUtils.apiFetch(`/admin/users/${userId}`, 'DELETE');
            Utils.showToast(`사용자 "${username}"가 삭제되었습니다.`, 'success');
            ModalManager.closeActiveModal();
            this.loadUsersList();
        } finally { Utils.setLoading(false); }
    },
    updateStats: function(users) {
        if(DOM.totalUsersSpan) DOM.totalUsersSpan.textContent = users.length;
        if(DOM.usersWithKeysSpan) DOM.usersWithKeysSpan.textContent = users.filter(u => u.has_ssh_key).length;
    },
    clearStats: () => {
        if(DOM.totalUsersSpan) DOM.totalUsersSpan.textContent = '-';
        if(DOM.usersWithKeysSpan) DOM.usersWithKeysSpan.textContent = '-';
    },
    showAdminOnlyMessage: function() {
        if(DOM.usersList) DOM.usersList.innerHTML = `<div class="empty-state">🔒 관리자 전용 기능입니다.</div>`;
        this.clearStats();
    },
    showError: function(message) {
        if(DOM.usersList) DOM.usersList.innerHTML = `<div class="error-message">${message} <button data-action="reload-users" class="btn-secondary btn-sm">다시 시도</button></div>`;
        this.clearStats();
    },
    isAdmin: () => AppState.currentUser?.role === 'admin',
    isCurrentUser: (userId) => AppState.currentUser?.id === userId
};