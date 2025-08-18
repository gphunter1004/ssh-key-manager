// 사용자 관리자
window.UserManager = {
    users: [],
    stats: null,

    setupEventListeners: function() {
        // 사용자 카드 클릭 이벤트는 동적으로 생성되므로 여기서는 설정하지 않음
        console.log('UserManager 이벤트 리스너 설정 완료');
    },

    loadUsersList: async function() {
        console.log('사용자 목록 로드 시작');
        
        try {
            // 로딩 상태 표시
            this.setLoadingState('사용자 목록을 불러오는 중...');
            
            const data = await AppUtils.apiFetch('/users', 'GET');
            
            this.users = data.users;
            console.log(`사용자 목록 로드 성공: ${data.count}명`);
            
            // 사용자 목록 표시
            this.displayUsersList(data.users);
            
            // 통계 정보 업데이트
            this.updateStats(data);
            
        } catch (error) {
            console.error('사용자 목록 로드 실패:', error.message);
            DOM.usersList.innerHTML = '<div class="error-message">사용자 목록을 불러올 수 없습니다.</div>';
            this.clearStats();
        } finally {
            this.clearLoadingState();
        }
    },

    displayUsersList: function(users) {
        console.log('사용자 목록 표시 중...');
        
        // 목록 초기화
        DOM.usersList.innerHTML = '';
        
        if (users.length === 0) {
            DOM.usersList.innerHTML = '<div class="empty-message">등록된 사용자가 없습니다.</div>';
            return;
        }

        // 사용자 카드 생성
        users.forEach(user => {
            const userCard = this.createUserCard(user);
            DOM.usersList.appendChild(userCard);
        });
        
        console.log(`사용자 카드 ${users.length}개 생성 완료`);
    },

    createUserCard: function(user) {
        const userCard = document.createElement('div');
        userCard.className = `user-card ${user.has_ssh_key ? 'has-key' : 'no-key'}`;
        userCard.dataset.userId = user.id;
        
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
            <div class="user-name">${this.escapeHtml(user.username)}</div>
            <div class="user-meta">
                ID: ${user.id} | 가입: ${createdDate} | 마지막 활동: ${lastActivity}
            </div>
            <div class="user-status ${user.has_ssh_key ? 'has-key' : 'no-key'}">
                ${user.has_ssh_key ? '🔑 SSH 키 보유' : '❌ SSH 키 없음'}
            </div>
            <div class="user-actions">
                <button class="view-detail-btn" data-user-id="${user.id}">상세 보기</button>
            </div>
        `;
        
        // 카드 클릭 이벤트
        userCard.addEventListener('click', (e) => {
            // 버튼 클릭이 아닌 경우에만 상세 정보 표시
            if (!e.target.classList.contains('view-detail-btn')) {
                this.showUserDetail(user.id);
            }
        });
        
        // 상세 보기 버튼 이벤트 (이벤트 전파 방지)
        const detailBtn = userCard.querySelector('.view-detail-btn');
        detailBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.showUserDetail(user.id);
        });
        
        return userCard;
    },

    showUserDetail: async function(userId) {
        console.log('사용자 상세 정보 요청:', userId);
        
        try {
            // 모달 로딩 상태
            ModalManager.showLoadingModal('사용자 정보를 불러오는 중...');
            
            const userData = await AppUtils.apiFetch(`/users/${userId}`, 'GET');
            
            console.log('사용자 상세 정보 로드 성공:', userData.username);
            
            // 상세 정보 표시
            ModalManager.displayUserDetail(userData);
            
        } catch (error) {
            console.error('사용자 상세 정보 로드 실패:', error.message);
            ModalManager.showErrorModal('사용자 정보를 불러올 수 없습니다.');
        }
    },

    updateStats: function(data) {
        const totalUsers = data.count;
        const usersWithKeys = data.users.filter(user => user.has_ssh_key).length;
        const usersWithoutKeys = totalUsers - usersWithKeys;
        const coveragePercent = totalUsers > 0 ? Math.round((usersWithKeys / totalUsers) * 100) : 0;
        
        // 통계 업데이트
        DOM.totalUsersSpan.textContent = totalUsers;
        DOM.usersWithKeysSpan.textContent = usersWithKeys;
        
        // 추가 통계 정보 저장
        this.stats = {
            total: totalUsers,
            withKeys: usersWithKeys,
            withoutKeys: usersWithoutKeys,
            coverage: coveragePercent
        };
        
        console.log('사용자 통계 업데이트:', this.stats);
        
        // 통계 색상 업데이트
        this.updateStatsColors(coveragePercent);
    },

    updateStatsColors: function(coveragePercent) {
        // 커버리지에 따른 색상 변경
        const statsSection = document.getElementById('user-stats');
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
        DOM.totalUsersSpan.textContent = '-';
        DOM.usersWithKeysSpan.textContent = '-';
        this.stats = null;
    },

    setLoadingState: function(message) {
        DOM.usersList.innerHTML = `<div class="loading-message">${message}</div>`;
        this.clearStats();
    },

    clearLoadingState: function() {
        // 로딩 상태는 displayUsersList에서 자동으로 클리어됨
    },

    // 사용자 검색 및 필터링
    filterUsers: function(searchTerm, keyFilter = 'all') {
        if (!this.users || this.users.length === 0) {
            console.log('필터링할 사용자가 없음');
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
        
        console.log(`사용자 필터링 결과: ${filteredUsers.length}/${this.users.length}`);
        
        // 필터링된 목록 표시
        this.displayUsersList(filteredUsers);
        
        // 필터링된 통계 업데이트
        this.updateFilteredStats(filteredUsers);
    },

    updateFilteredStats: function(filteredUsers) {
        const totalFiltered = filteredUsers.length;
        const withKeysFiltered = filteredUsers.filter(user => user.has_ssh_key).length;
        
        // 임시 통계 표시 (원본 통계는 유지)
        DOM.totalUsersSpan.textContent = `${totalFiltered} (전체: ${this.stats?.total || '-'})`;
        DOM.usersWithKeysSpan.textContent = `${withKeysFiltered} (전체: ${this.stats?.withKeys || '-'})`;
    },

    // 사용자 정렬
    sortUsers: function(sortBy = 'username', sortOrder = 'asc') {
        if (!this.users || this.users.length === 0) {
            console.log('정렬할 사용자가 없음');
            return;
        }
        
        const sortedUsers = [...this.users].sort((a, b) => {
            let aValue, bValue;
            
            switch (sortBy) {
                case 'username':
                    aValue = a.username.toLowerCase();
                    bValue = b.username.toLowerCase();
                    break;
                case 'id':
                    aValue = a.id;
                    bValue = b.id;
                    break;
                case 'created_at':
                    aValue = new Date(a.created_at);
                    bValue = new Date(b.created_at);
                    break;
                case 'has_ssh_key':
                    aValue = a.has_ssh_key ? 1 : 0;
                    bValue = b.has_ssh_key ? 1 : 0;
                    break;
                default:
                    return 0;
            }
            
            let comparison = 0;
            if (aValue < bValue) comparison = -1;
            if (aValue > bValue) comparison = 1;
            
            return sortOrder === 'desc' ? -comparison : comparison;
        });
        
        console.log(`사용자 정렬: ${sortBy} ${sortOrder}`);
        this.displayUsersList(sortedUsers);
    },

    // 새로고침
    refresh: async function() {
        console.log('사용자 목록 새로고침');
        await this.loadUsersList();
    },

    // HTML 이스케이프 유틸리티
    escapeHtml: function(unsafe) {
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    },

    // 현재 사용자인지 확인
    isCurrentUser: function(userId) {
        return AppState.currentUser && AppState.currentUser.id === userId;
    }
};