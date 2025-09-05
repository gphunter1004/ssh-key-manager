// 사용자 관리자 - 완전 독립 버전
window.UserManager = {
    users: [],
    stats: null,

    init: function() {
        console.log('👥 UserManager 초기화 시작');
        this.setupEventListeners();
        console.log('✅ UserManager 초기화 완료');
        return true;
    },

    setupEventListeners: function() {
        console.log('🎯 UserManager 이벤트 리스너 설정');
        
        document.addEventListener('click', (e) => {
            const action = e.target.dataset.action;
            if (!action) return;

            const [actionType, actionData] = action.split(':');
            
            switch (actionType) {
                case 'user-detail':
                    e.preventDefault();
                    this.showUserDetail(parseInt(actionData));
                    break;
                case 'user-delete':
                    e.preventDefault();
                    this.confirmDeleteUser(parseInt(actionData), e.target.dataset.username);
                    break;
                case 'reload-users':
                    e.preventDefault();
                    this.loadUsersList();
                    break;
            }
        });
        
        console.log('✅ UserManager 이벤트 리스너 설정 완료');
    },

    loadUsersList: async function() {
        console.log('📋 사용자 목록 로드 시작');
        
        if (!this.isAdmin()) {
            this.showAdminOnlyMessage();
            return;
        }
        
        try {
            Utils.setLoading(true, '사용자 목록을 불러오는 중...');
            
            const users = await AppUtils.apiFetch('/admin/users', 'GET');
            
            this.users = Array.isArray(users) ? users.map(user => ({
                id: user.ID,
                username: user.username,
                role: user.role,
                is_active: user.is_active,
                is_locked: user.is_locked,
                created_at: user.CreatedAt,
                updated_at: user.UpdatedAt,
                has_ssh_key: user.has_ssh_key // 이 필드가 존재한다면
            })) : [];
            
            console.log(`✅ 사용자 목록 로드 성공: ${this.users.length}명`);
            
            this.displayUsersList(this.users);
            this.updateStats(this.users);
            
        } catch (error) {
            console.error('❌ 사용자 목록 로드 실패:', error.message);
            this.showError('사용자 목록을 불러올 수 없습니다: ' + error.message);
            this.clearStats();
        } finally {
            Utils.setLoading(false);
        }
    },

    displayUsersList: function(users) {
        console.log(`👥 사용자 목록 표시: ${users.length}명`);
        
        if (!DOM.usersList) {
            console.error('❌ users-list 요소를 찾을 수 없음');
            return;
        }

        DOM.usersList.innerHTML = '';
        
        if (users.length === 0) {
            DOM.usersList.innerHTML = '<div class="empty-state">등록된 사용자가 없습니다.</div>';
            return;
        }

        users.forEach(user => {
            const userCard = this.createUserCard(user);
            DOM.usersList.appendChild(userCard);
        });
        
        console.log(`✅ 사용자 카드 ${users.length}개 생성 완료`);
    },

    createUserCard: function(user) {
        const userCard = document.createElement('div');
        userCard.className = `user-card ${user.has_ssh_key ? 'has-key' : 'no-key'}`;
        userCard.dataset.userId = user.id;
        
        if (user.role === 'admin') {
            userCard.classList.add('admin-user');
        }
        
        const createdDate = user.created_at ? Utils.formatDate(user.created_at, {year: 'numeric', month: 'short', day: 'numeric'}) : '정보 없음';
        
        userCard.innerHTML = `
            <div class="user-info-header">
                <div class="user-name-group">
                    <span class="user-name">${Utils.escapeHtml(user.username)}</span>
                    ${user.role === 'admin' ? '<span class="admin-badge">관리자</span>' : ''}
                </div>
                <div class="user-actions">
                    <button class="btn-secondary btn-sm" data-action="user-detail:${user.id}">
                        상세 보기
                    </button>
                    ${this.isAdmin() && user.role !== 'admin' ? `
                        <button class="btn-danger btn-sm" data-action="user-delete:${user.id}" 
                                data-username="${Utils.escapeHtml(user.username)}">
                            삭제
                        </button>
                    ` : ''}
                </div>
            </div>
            <div class="user-info-body">
                <div class="user-meta-item">
                    <strong>ID:</strong> 
                    <span>${user.id || '정보 없음'}</span>
                </div>
                <div class="user-meta-item">
                    <strong>가입일:</strong> 
                    <span>${createdDate}</span>
                </div>
                <div class="user-meta-item">
                    <strong>SSH 키 상태:</strong>
                    <span class="user-status ${user.has_ssh_key ? 'has-key' : 'no-key'}">
                        ${user.has_ssh_key ? '🔑 보유' : '❌ 없음'}
                    </span>
                </div>
            </div>
        `;
        
        userCard.addEventListener('click', (e) => {
            if (!e.target.matches('button')) {
                this.showUserDetail(user.id);
            }
        });
        
        return userCard;
    },

    showUserDetail: async function(userId) {
        console.log(`👤 사용자 상세 정보 요청: ${userId}`);
        
        try {
            if (typeof ModalManager !== 'undefined' && ModalManager.showLoadingModal) {
                ModalManager.showLoadingModal('사용자 정보를 불러오는 중...');
            }
            
            const userData = await AppUtils.apiFetch(`/admin/users/${userId}`, 'GET');
            
            console.log('✅ 사용자 상세 정보 로드 성공:', userData.username);
            
            if (typeof ModalManager !== 'undefined' && ModalManager.displayUserDetail) {
                ModalManager.displayUserDetail(userData);
            } else {
                console.warn('⚠️ ModalManager를 찾을 수 없음');
                Utils.showToast(`사용자 정보: ${userData.username} (${userData.role})`, 'info');
            }
            
        } catch (error) {
            console.error('❌ 사용자 상세 정보 로드 실패:', error.message);
            
            if (typeof ModalManager !== 'undefined' && ModalManager.showErrorModal) {
                ModalManager.showErrorModal('사용자 정보를 불러올 수 없습니다.');
            } else {
                Utils.showToast('사용자 정보를 불러올 수 없습니다: ' + error.message, 'error');
            }
        }
    },

    confirmDeleteUser: async function(userId, username) {
        console.log(`🗑️ 사용자 삭제 확인: ${userId} (${username})`);
        
        if (!confirm(`정말로 사용자 "${username}"를 삭제하시겠습니까?\n\n이 작업은 되돌릴 수 없습니다.`)) {
            return;
        }

        try {
            Utils.setLoading(true, '사용자 삭제 중...');
            
            await AppUtils.apiFetch(`/admin/users/${userId}`, 'DELETE');
            
            Utils.showToast(`사용자 "${username}"가 삭제되었습니다`, 'success');
            console.log('✅ 사용자 삭제 성공:', username);
            
            if (typeof ModalManager !== 'undefined' && ModalManager.isOpen) {
                ModalManager.closeModal();
            }
            
            this.loadUsersList();
            
        } catch (error) {
            console.error('❌ 사용자 삭제 실패:', error.message);
            Utils.showToast('사용자 삭제에 실패했습니다: ' + error.message, 'error');
        } finally {
            Utils.setLoading(false);
        }
    },

    updateStats: function(users) {
        const totalUsers = users.length;
        const usersWithKeys = users.filter(user => user.has_ssh_key).length;
        const usersWithoutKeys = totalUsers - usersWithKeys;
        const coveragePercent = totalUsers > 0 ? Math.round((usersWithKeys / totalUsers) * 100) : 0;
        
        if (DOM.totalUsersSpan) {
            DOM.totalUsersSpan.textContent = totalUsers;
        }
        if (DOM.usersWithKeysSpan) {
            DOM.usersWithKeysSpan.textContent = usersWithKeys;
        }
        
        this.stats = {
            total: totalUsers,
            withKeys: usersWithKeys,
            withoutKeys: usersWithoutKeys,
            coverage: coveragePercent
        };
        
        console.log('📊 사용자 통계 업데이트:', this.stats);
        
        const statsSection = document.querySelector('.stats-section');
        if (statsSection) {
            statsSection.className = 'stats-section';
            if (coveragePercent >= 80) {
                statsSection.classList.add('high-coverage');
            } else if (coveragePercent >= 50) {
                statsSection.classList.add('medium-coverage');
            } else {
                statsSection.classList.add('low-coverage');
            }
        }
    },

    clearStats: function() {
        if (DOM.totalUsersSpan) {
            DOM.totalUsersSpan.textContent = '-';
        }
        if (DOM.usersWithKeysSpan) {
            DOM.usersWithKeysSpan.textContent = '-';
        }
        this.stats = null;
    },

    showAdminOnlyMessage: function() {
        if (DOM.usersList) {
            DOM.usersList.innerHTML = `
                <div class="admin-only-message">
                    <div class="icon">🔒</div>
                    <h3>관리자 전용 기능</h3>
                    <p>사용자 목록은 관리자만 조회할 수 있습니다.</p>
                </div>
            `;
        }
        this.clearStats();
    },

    showError: function(message) {
        if (DOM.usersList) {
            DOM.usersList.innerHTML = `
                <div class="error-message">
                    <div class="icon">❌</div>
                    <h3>오류 발생</h3>
                    <p>${Utils.escapeHtml(message)}</p>
                    <button data-action="reload-users" class="btn-primary">다시 시도</button>
                </div>
            `;
        }
        this.clearStats();
    },

    filterUsers: function(searchTerm, keyFilter = 'all') {
        if (!this.users || this.users.length === 0) {
            console.log('ℹ️ 필터링할 사용자가 없음');
            return;
        }
        
        let filteredUsers = this.users;
        
        if (searchTerm && searchTerm.trim() !== '') {
            const term = searchTerm.toLowerCase().trim();
            filteredUsers = filteredUsers.filter(user => 
                user.username.toLowerCase().includes(term) ||
                user.id.toString().includes(term)
            );
        }
        
        if (keyFilter === 'with-keys') {
            filteredUsers = filteredUsers.filter(user => user.has_ssh_key);
        } else if (keyFilter === 'without-keys') {
            filteredUsers = filteredUsers.filter(user => !user.has_ssh_key);
        }
        
        console.log(`🔍 사용자 필터링 결과: ${filteredUsers.length}/${this.users.length}`);
        
        this.displayUsersList(filteredUsers);
        this.updateFilteredStats(filteredUsers);
    },

    updateFilteredStats: function(filteredUsers) {
        const totalFiltered = filteredUsers.length;
        const withKeysFiltered = filteredUsers.filter(user => user.has_ssh_key).length;
        
        if (DOM.totalUsersSpan) {
            DOM.totalUsersSpan.textContent = `${totalFiltered} (전체: ${this.stats?.total || '-'})`;
        }
        if (DOM.usersWithKeysSpan) {
            DOM.usersWithKeysSpan.textContent = `${withKeysFiltered} (전체: ${this.stats?.withKeys || '-'})`;
        }
    },

    refresh: async function() {
        console.log('🔄 사용자 목록 새로고침');
        await this.loadUsersList();
    },

    isAdmin: function() {
        return AppState.currentUser?.role === 'admin';
    },

    isCurrentUser: function(userId) {
        return AppState.currentUser && AppState.currentUser.id === userId;
    },

    getStats: function() {
        return {
            totalUsers: this.users.length,
            currentStats: this.stats,
            isAdmin: this.isAdmin(),
            currentUser: AppState.currentUser
        };
    }
};