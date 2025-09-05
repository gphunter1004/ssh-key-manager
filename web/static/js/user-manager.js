// 사용자 관리자 (최종 수정본 - 모달 로직 개선)
window.UserManager = {
    users: [],
    init: function() {
        console.log('👥 UserManager 초기화 시작');
        this.setupEventListeners();
        return true;
    },

    setupEventListeners: function() {
        document.addEventListener('click', (e) => {
            const button = e.target.closest('button[data-action]');
            if (!button) return;

            const action = button.dataset.action;
            const userId = button.dataset.userId;
            const username = button.dataset.username;

            if (action === 'user-detail') this.showUserDetail(parseInt(userId));
            else if (action === 'user-delete') this.confirmDeleteUser(parseInt(userId), username);
            else if (action === 'reload-users') this.loadUsersList();
        });
    },

    loadUsersList: async function() {
        if (!AuthManager.isAdmin()) return this.showAdminOnlyMessage();
        
        Utils.setLoading(true, '사용자 목록 로딩 중...');
        try {
            const usersData = await AppUtils.apiFetch('/admin/users', 'GET');
            this.users = usersData.map(u => ({...u, id: u.ID, created_at: u.CreatedAt}));
            this.displayUsersList(this.users);
            this.updateStats(this.users);
        } catch (error) {
            this.showError('사용자 목록을 불러올 수 없습니다.');
        } finally {
            Utils.setLoading(false);
        }
    },

    displayUsersList: function(users) {
        DOM.usersList.innerHTML = '';
        if (users.length === 0) {
            DOM.usersList.innerHTML = '<div class="empty-state">등록된 사용자가 없습니다.</div>';
            return;
        }
        users.forEach(user => DOM.usersList.appendChild(this.createUserCard(user)));
    },

    createUserCard: function(user) {
        const card = document.createElement('div');
        card.className = `user-card ${user.has_ssh_key ? 'has-key' : 'no-key'}`;
        const canDelete = AppState.currentUser.id !== user.id && user.role !== 'admin';

        card.innerHTML = `
            <div class="user-info">
                <span class="user-name">${Utils.escapeHtml(user.username)} ${user.role === 'admin' ? '<span class="admin-badge">Admin</span>' : ''}</span>
                <span class="user-meta">가입: ${Utils.formatDate(user.created_at, {month:'2-digit', day:'2-digit'})}</span>
            </div>
            <div class="user-actions">
                <button class="btn-secondary btn-sm" data-action="user-detail" data-user-id="${user.id}">상세</button>
                ${canDelete ? `<button class="btn-danger btn-sm" data-action="user-delete" data-user-id="${user.id}" data-username="${Utils.escapeHtml(user.username)}">삭제</button>` : ''}
            </div>
        `;
        card.addEventListener('click', (e) => {
            if (!e.target.closest('button')) this.showUserDetail(user.id);
        });
        return card;
    },

    showUserDetail: async function(userId) {
        const modalId = 'user-detail-modal';
        ModalManager.openModalWithSpinner(modalId, '사용자 정보 로딩 중...');

        try {
            // 1. 사용자 기본 정보 가져오기
            const userData = await AppUtils.apiFetch(`/admin/users/${userId}`);
            
            let keyData = null;
            // 2. 키가 있으면 키 정보도 가져오기
            if (userData.has_ssh_key) {
                try {
                     keyData = await AppUtils.apiFetch(`/admin/users/${userId}/keys`);
                } catch (keyError) {
                    console.warn(`사용자(${userId})의 키 정보를 가져오는 데 실패했습니다.`);
                }
            }

            // 3. 모든 정보로 HTML 생성
            const title = `${userData.username} 상세 정보`;
            const contentHTML = this.generateModalContent(userData, keyData);
            
            // 4. 완성된 내용으로 모달 업데이트
            ModalManager.updateModalContent(title, contentHTML);

        } catch (error) {
            ModalManager.updateModalContent('오류', `<div class="error-message">정보 로딩에 실패했습니다.</div>`);
        }
    },

    generateModalContent: function(userData, keyData) {
        const createdDate = Utils.formatDate(userData.created_at);
        const isAdmin = userData.role === 'admin';
        const isCurrentUser = this.isCurrentUser(userData.id);

        let keyInfoHTML = `
            <div class="ssh-key-section">
                <h3 class="modal-subheader">SSH 키 정보</h3>
                <div class="empty-state" style="padding: 1rem; text-align: left;">SSH 키가 생성되지 않았습니다.</div>
            </div>`;
        
        if (keyData) {
            const keyContentHTML = isCurrentUser 
                ? `<h4 class="modal-key-title">개인키 (PEM)</h4>
                   <div class="key-display-wrapper"><pre class="key-content" id="modal-pem-key">${Utils.escapeHtml(keyData.PEM)}</pre><button class="copy-btn" data-target="#modal-pem-key">복사</button></div>`
                : `<p class="security-notice warning">🔒 보안상 관리자는 다른 사용자의 개인키를 볼 수 없습니다.</p>`;

            keyInfoHTML = `
                <div class="ssh-key-section">
                    <h3 class="modal-subheader">SSH 키 정보</h3>
                    <div class="info-grid">
                        <div class="info-item"><strong>생성일:</strong> <span>${Utils.formatDate(keyData.created_at)}</span></div>
                        <div class="info-item"><strong>알고리즘:</strong> <span>${keyData.Algorithm || 'RSA'} / ${keyData.Bits || 4096} bits</span></div>
                    </div>
                    <h4 class="modal-key-title">공개키 (Public Key)</h4>
                    <div class="key-display-wrapper">
                        <pre class="key-content" id="modal-public-key">${Utils.escapeHtml(keyData.PublicKey)}</pre>
                        <button class="copy-btn" data-target="#modal-public-key">복사</button>
                    </div>
                    ${keyContentHTML}
                </div>`;
        }

        const actionButtonsHTML = isCurrentUser
            ? `<div class="modal-actions"><button class="btn-primary" onclick="ViewManager.showView('profile'); ModalManager.closeActiveModal();">프로필 편집</button></div>`
            : '';

        return `
            <div class="user-detail-container">
                <div class="user-basic-info">
                    <h3 class="modal-subheader">기본 정보</h3>
                    <div class="info-grid">
                        <div class="info-item"><strong>사용자명:</strong> <span>${Utils.escapeHtml(userData.username)} ${isAdmin ? '<span class="admin-badge">Admin</span>':''} ${isCurrentUser ? '<span style="font-size: 0.7rem; background: #eee; padding: 2px 6px; border-radius: 10px;">나</span>':''}</span></div>
                        <div class="info-item"><strong>사용자 ID:</strong> <span>${userData.id}</span></div>
                        <div class="info-item"><strong>가입일:</strong> <span>${createdDate}</span></div>
                    </div>
                </div>
                ${keyInfoHTML}
                ${actionButtonsHTML}
            </div>
        `;
    },

    confirmDeleteUser: async function(userId, username) {
        if (!confirm(`사용자 "${username}"를 정말로 삭제하시겠습니까?`)) return;
        Utils.setLoading(true, '사용자 삭제 중...');
        try {
            await AppUtils.apiFetch(`/admin/users/${userId}`, 'DELETE');
            Utils.showToast(`사용자 "${username}"가 삭제되었습니다.`, 'success');
            ModalManager.closeActiveModal();
            this.loadUsersList();
        } finally {
            Utils.setLoading(false);
        }
    },

    updateStats: function(users) {
        DOM.totalUsersSpan.textContent = users.length;
        DOM.usersWithKeysSpan.textContent = users.filter(u => u.has_ssh_key).length;
    },

    clearStats: () => {
        DOM.totalUsersSpan.textContent = '-';
        DOM.usersWithKeysSpan.textContent = '-';
    },
    
    showAdminOnlyMessage: function() {
        DOM.usersList.innerHTML = `<div class="empty-state">🔒 관리자 전용 기능입니다.</div>`;
        this.clearStats();
    },

    showError: function(message) {
        DOM.usersList.innerHTML = `<div class="error-message">${message} <button data-action="reload-users" class="btn-secondary btn-sm">다시 시도</button></div>`;
        this.clearStats();
    },

    isAdmin: () => AppState.currentUser?.role === 'admin',
    isCurrentUser: (userId) => AppState.currentUser?.id === userId
};