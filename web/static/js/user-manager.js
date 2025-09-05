// 사용자 관리자 - 완전 독립 버전
window.UserManager = {
    users: [],
    stats: null,

    init: function() {
        console.log('👥 UserManager 초기화 시작');
        
        // 이벤트 리스너 설정
        this.setupEventListeners();
        
        console.log('✅ UserManager 초기화 완료');
        return true;
    },

    setupEventListeners: function() {
        console.log('🎯 UserManager 이벤트 리스너 설정');
        
        // 액션 기반 이벤트 위임
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
        
        // 관리자 권한 확인
        if (!this.isAdmin()) {
            this.showAdminOnlyMessage();
            return;
        }
        
        try {
            // 로딩 상태 표시
            this.setLoadingState('사용자 목록을 불러오는 중...');
            
            const users = await AppUtils.apiFetch('/admin/users', 'GET');
            
            this.users = Array.isArray(users) ? users : (users.items || users.data || []);
            console.log(`✅ 사용자 목록 로드 성공: ${this.users.length}명`);
            
            // 사용자 목록 표시
            this.displayUsersList(this.users);
            
            // 통계 정보 업데이트
            this.updateStats(this.users);
            
        } catch (error) {
            console.error('❌ 사용자 목록 로드 실패:', error.message);
            this.showError('사용자 목록을 불러올 수 없습니다: ' + error.message);
            this.clearStats();
        } finally {
            this.clearLoadingState();
        }
    },

    displayUsersList: function(users) {
        console.log(`👥 사용자 목록 표시: ${users.length}명`);
        
        if (!DOM.usersList) {
            console.error('❌ users-list 요소를 찾을 수 없음');
            return;
        }

        // 목록 초기화
        DOM.usersList.innerHTML = '';
        
        if (users.length === 0) {
            DOM.usersList.innerHTML = '<div class="empty-state">등록된 사용자가 없습니다.</div>';
            return;
        }

        // 사용자 카드 생성
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
        
        // 관리자 사용자 특별 스타일
        if (user.role === 'admin') {
            userCard.classList.add('admin-user');
        }
        
        // 날짜 포맷팅
        const createdDate = new Date(user.created_at).toLocaleDateString('ko-KR', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
        
        // 마지막 활동 계산
        const updatedDate = new Date(user.updated_at);
        const now = new Date();
        const daysDiff = Math.floor((now - updatedDate) / (1000 * 60 * 60 * 24));
        const lastActivity = daysDiff === 0 ? '오늘' : `${daysDiff}일 전`;
        
        userCard.innerHTML = `
            <div class="user-info">
                <div class="user-name">
                    ${this.escapeHtml(user.username)}
                    ${user.role === 'admin' ? '<span class="admin-badge">관리자</span>' : ''}
                </div>
                <div class="user-meta">
                    ID: ${user.id} • 가입: ${createdDate} • 마지막 활동: ${lastActivity}
                </div>
                <div class="user-status ${user.has_ssh_key ? 'has-key' : 'no-key'}">
                    ${user.has_ssh_key ? '🔑 SSH 키 보유' : '❌ SSH 키 없음'}
                </div>
            </div>
            <div class="user-actions">
                <button class="btn-secondary btn-sm" data-action="user-detail:${user.id}">
                    상세 보기
                </button>
                ${this.isAdmin() && user.role !== 'admin' ? `
                    <button class="btn-danger btn-sm" data-action="user-delete:${user.id}" 
                            data-username="${this.escapeHtml(user.username)}">
                        삭제
                    </button>
                ` : ''}
            </div>
        `;
        
        // 카드 클릭 이벤트 (상세 보기)
        userCard.addEventListener('click', (e) => {
            // 버튼 클릭이 아닌 경우에만 상세 정보 표시
            if (!e.target.matches('button')) {
                this.showUserDetail(user.id);
            }
        });
        
        return userCard;
    },

    showUserDetail: async function(userId) {
        console.log(`👤 사용자 상세 정보 요청: ${userId}`);
        
        try {
            // ModalManager에게 로딩 모달 요청
            if (typeof ModalManager !== 'undefined' && ModalManager.showLoadingModal) {
                ModalManager.showLoadingModal('사용자 정보를 불러오는 중...');
            }
            
            const userData = await AppUtils.apiFetch(`/admin/users/${userId}`, 'GET');
            
            console.log('✅ 사용자 상세 정보 로드 성공:', userData.username);
            
            // ModalManager에게 상세 정보 표시 요청
            if (typeof ModalManager !== 'undefined' && ModalManager.displayUserDetail) {
                ModalManager.displayUserDetail(userData);
            } else {
                console.warn('⚠️ ModalManager를 찾을 수 없음');
                this.showMessage(`사용자 정보: ${userData.username} (${userData.role})`, 'info');
            }
            
        } catch (error) {
            console.error('❌ 사용자 상세 정보 로드 실패:', error.message);
            
            if (typeof ModalManager !== 'undefined' && ModalManager.showErrorModal) {
                ModalManager.showErrorModal('사용자 정보를 불러올 수 없습니다.');
            } else {
                this.showMessage('사용자 정보를 불러올 수 없습니다: ' + error.message, 'error');
            }
        }
    },

    confirmDeleteUser: async function(userId, username) {
        console.log(`🗑️ 사용자 삭제 확인: ${userId} (${username})`);
        
        if (!confirm(`정말로 사용자 "${username}"를 삭제하시겠습니까?\n\n이 작업은 되돌릴 수 없습니다.`)) {
            return;
        }

        try {
            this.setLoading(true, '사용자 삭제 중...');
            
            await AppUtils.apiFetch(`/admin/users/${userId}`, 'DELETE');
            
            this.showMessage(`사용자 "${username}"가 삭제되었습니다`, 'success');
            console.log('✅ 사용자 삭제 성공:', username);
            
            // 모달이 열려있으면 닫기
            if (typeof ModalManager !== 'undefined' && ModalManager.isOpen) {
                ModalManager.closeModal();
            }
            
            // 목록 새로고침
            this.loadUsersList();
            
        } catch (error) {
            console.error('❌ 사용자 삭제 실패:', error.message);
            this.showMessage('사용자 삭제에 실패했습니다: ' + error.message, 'error');
        } finally {
            this.setLoading(false);
        }
    },

    updateStats: function(users) {
        const totalUsers = users.length;
        const usersWithKeys = users.filter(user => user.has_ssh_key).length;
        const usersWithoutKeys = totalUsers - usersWithKeys;
        const coveragePercent = totalUsers > 0 ? Math.round((usersWithKeys / totalUsers) * 100) : 0;
        
        // 통계 업데이트
        if (DOM.totalUsersSpan) {
            DOM.totalUsersSpan.textContent = totalUsers;
        }
        if (DOM.usersWithKeysSpan) {
            DOM.usersWithKeysSpan.textContent = usersWithKeys;
        }
        
        // 추가 통계 정보 저장
        this.stats = {
            total: totalUsers,
            withKeys: usersWithKeys,
            withoutKeys: usersWithoutKeys,
            coverage: coveragePercent
        };
        
        console.log('📊 사용자 통계 업데이트:', this.stats);
        
        // 통계 색상 업데이트
        this.updateStatsColors(coveragePercent);
    },

    updateStatsColors: function(coveragePercent) {
        // 커버리지에 따른 색상 변경
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

    setLoadingState: function(message) {
        if (DOM.usersList) {
            DOM.usersList.innerHTML = `<div class="loading-state">${message}</div>`;
        }
        this.clearStats();
    },

    clearLoadingState: function() {
        // 로딩 상태는 displayUsersList에서 자동으로 클리어됨
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
                    <p>${this.escapeHtml(message)}</p>
                    <button data-action="reload-users" class="btn-primary">다시 시도</button>
                </div>
            `;
        }
        this.clearStats();
    },

    // 사용자 검색 및 필터링
    filterUsers: function(searchTerm, keyFilter = 'all') {
        if (!this.users || this.users.length === 0) {
            console.log('ℹ️ 필터링할 사용자가 없음');
            return;
        }
        
        let filteredUsers = this.users;
        
        // 검색어 필터링
        if (searchTerm && searchTerm.trim() !== '') {
            const term = searchTerm.toLowerCase().trim();
            filteredUsers = filteredUsers.filter(user => 
                user.username.toLowerCase().includes(term) ||
                user.id.toString().includes(term)
            );
        }
        
        // 키 상태 필터링
        if (keyFilter === 'with-keys') {
            filteredUsers = filteredUsers.filter(user => user.has_ssh_key);
        } else if (keyFilter === 'without-keys') {
            filteredUsers = filteredUsers.filter(user => !user.has_ssh_key);
        }
        
        console.log(`🔍 사용자 필터링 결과: ${filteredUsers.length}/${this.users.length}`);
        
        // 필터링된 목록 표시
        this.displayUsersList(filteredUsers);
        
        // 필터링된 통계 업데이트
        this.updateFilteredStats(filteredUsers);
    },

    updateFilteredStats: function(filteredUsers) {
        const totalFiltered = filteredUsers.length;
        const withKeysFiltered = filteredUsers.filter(user => user.has_ssh_key).length;
        
        // 임시 통계 표시 (원본 통계는 유지)
        if (DOM.totalUsersSpan) {
            DOM.totalUsersSpan.textContent = `${totalFiltered} (전체: ${this.stats?.total || '-'})`;
        }
        if (DOM.usersWithKeysSpan) {
            DOM.usersWithKeysSpan.textContent = `${withKeysFiltered} (전체: ${this.stats?.withKeys || '-'})`;
        }
    },

    // 새로고침
    refresh: async function() {
        console.log('🔄 사용자 목록 새로고침');
        await this.loadUsersList();
    },

    // 권한 확인
    isAdmin: function() {
        return AppState.currentUser?.role === 'admin';
    },

    // 현재 사용자인지 확인
    isCurrentUser: function(userId) {
        return AppState.currentUser && AppState.currentUser.id === userId;
    },

    // 로딩 상태 관리
    setLoading: function(isLoading, message = '처리 중...') {
        if (typeof Utils !== 'undefined' && Utils.setLoading) {
            Utils.setLoading(isLoading, message);
        } else {
            console.log(`로딩 상태: ${isLoading} - ${message}`);
        }
    },

    // 메시지 표시
    showMessage: function(message, type = 'info') {
        if (typeof Utils !== 'undefined' && Utils.showToast) {
            Utils.showToast(message, type);
        } else {
            console.log(`[${type.toUpperCase()}] ${message}`);
            
            // 중요한 메시지는 alert으로도 표시
            if (type === 'error' || type === 'warning') {
                alert(message);
            }
        }
    },

    // HTML 이스케이프 유틸리티
    escapeHtml: function(unsafe) {
        if (!unsafe) return '';
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    },

    // 디버깅 정보
    getStats: function() {
        return {
            totalUsers: this.users.length,
            currentStats: this.stats,
            isAdmin: this.isAdmin(),
            currentUser: AppState.currentUser
        };
    }
};