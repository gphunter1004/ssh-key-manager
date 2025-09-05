// 뷰 관리자 - 개선된 버전
window.ViewManager = {
    currentView: 'keys',

    setupEventListeners: function() {
        // 네비게이션 버튼 이벤트 설정 (null 체크 추가)
        if (DOM.navKeys) {
            DOM.navKeys.addEventListener('click', (e) => {
                e.preventDefault();
                this.showView('keys');
            });
        }
        
        if (DOM.navUsers) {
            DOM.navUsers.addEventListener('click', (e) => {
                e.preventDefault();
                this.showView('users');
            });
        }
        
        if (DOM.navProfile) {
            DOM.navProfile.addEventListener('click', (e) => {
                e.preventDefault();
                this.showView('profile');
            });
        }
        
        // 추가 네비게이션 버튼들 (존재하는 경우)
        const navServers = document.getElementById('nav-servers');
        const navDepartments = document.getElementById('nav-departments');
        
        if (navServers) {
            navServers.addEventListener('click', (e) => {
                e.preventDefault();
                this.showView('servers');
            });
        }
        
        if (navDepartments) {
            navDepartments.addEventListener('click', (e) => {
                e.preventDefault();
                this.showView('departments');
            });
        }
        
        console.log('ViewManager 이벤트 리스너 설정 완료');
    },

    showView: function(viewName) {
        console.log('🔄 뷰 전환 요청:', this.currentView, '->', viewName);
        
        // 권한 확인
        if (!this.checkViewAccess(viewName)) {
            Utils.showToast('해당 기능에 접근할 권한이 없습니다', 'warning');
            return;
        }
        
        // 이전 뷰 정리
        this.cleanupCurrentView();
        
        // 모든 뷰 숨기기
        this.hideAllViews();
        
        // 네비게이션 버튼 활성화 상태 초기화
        this.resetNavigation();
        
        // 선택된 뷰 표시 및 초기화
        const viewShown = this.displayView(viewName);
        
        if (viewShown) {
            // 현재 뷰 상태 업데이트
            this.currentView = viewName;
            AppState.currentView = viewName;
            
            // URL 해시 업데이트 (선택사항)
            this.updateUrlHash(viewName);
            
            console.log('✅ 뷰 전환 완료:', viewName);
        } else {
            console.warn('❌ 뷰 전환 실패, 기본 뷰로 이동');
            this.showView('keys'); // 기본값으로 키 관리 뷰 표시
        }
    },

    displayView: function(viewName) {
        switch(viewName) {
            case 'keys':
                return this.showKeysView();
            case 'users':
                return this.showUsersView();
            case 'profile':
                return this.showProfileView();
            case 'servers':
                return this.showServersView();
            case 'departments':
                return this.showDepartmentsView();
            default:
                console.warn('알 수 없는 뷰:', viewName);
                return false;
        }
    },

    hideAllViews: function() {
        const views = ['keys', 'users', 'profile', 'servers', 'departments'];
        views.forEach(view => {
            const viewElement = document.getElementById(`${view}-view`);
            if (viewElement) {
                viewElement.classList.add('hidden');
            }
        });
    },

    resetNavigation: function() {
        const navButtons = ['nav-keys', 'nav-users', 'nav-profile', 'nav-servers', 'nav-departments'];
        navButtons.forEach(navId => {
            const navElement = document.getElementById(navId);
            if (navElement) {
                navElement.classList.remove('active');
            }
        });
    },

    showKeysView: function() {
        const keysView = document.getElementById('keys-view');
        const navKeys = document.getElementById('nav-keys');
        
        if (!keysView) {
            console.error('❌ keys-view 요소를 찾을 수 없습니다');
            return false;
        }
        
        keysView.classList.remove('hidden');
        if (navKeys) {
            navKeys.classList.add('active');
        }
        
        console.log('✅ 키 관리 뷰 활성화');
        
        // 키 뷰 초기화 - 자동 로드
        if (typeof KeyManager !== 'undefined' && KeyManager.autoLoadKeys) {
            KeyManager.autoLoadKeys();
        }
        
        return true;
    },

    showUsersView: function() {
        const usersView = document.getElementById('users-view');
        const navUsers = document.getElementById('nav-users');
        
        if (!usersView) {
            console.error('❌ users-view 요소를 찾을 수 없습니다');
            return false;
        }
        
        usersView.classList.remove('hidden');
        if (navUsers) {
            navUsers.classList.add('active');
        }
        
        console.log('✅ 사용자 목록 뷰 활성화');
        
        // 사용자 목록 로드
        if (typeof UserManager !== 'undefined' && UserManager.loadUsersList) {
            UserManager.loadUsersList();
        }
        
        return true;
    },

    showProfileView: function() {
        const profileView = document.getElementById('profile-view');
        const navProfile = document.getElementById('nav-profile');
        
        if (!profileView) {
            console.error('❌ profile-view 요소를 찾을 수 없습니다');
            return false;
        }
        
        profileView.classList.remove('hidden');
        if (navProfile) {
            navProfile.classList.add('active');
        }
        
        console.log('✅ 프로필 뷰 활성화');
        
        // 프로필 정보 로드
        if (typeof ProfileManager !== 'undefined' && ProfileManager.loadCurrentUserProfile) {
            ProfileManager.loadCurrentUserProfile();
        }
        
        return true;
    },

    showServersView: function() {
        const serversView = document.getElementById('servers-view');
        const navServers = document.getElementById('nav-servers');
        
        if (!serversView) {
            console.error('❌ servers-view 요소를 찾을 수 없습니다');
            return false;
        }
        
        serversView.classList.remove('hidden');
        if (navServers) {
            navServers.classList.add('active');
        }
        
        console.log('✅ 서버 관리 뷰 활성화');
        
        // 서버 목록 로드 (ServerManager가 있는 경우)
        if (typeof ServerManager !== 'undefined' && ServerManager.loadServersList) {
            ServerManager.loadServersList();
        }
        
        return true;
    },

    showDepartmentsView: function() {
        const departmentsView = document.getElementById('departments-view');
        const navDepartments = document.getElementById('nav-departments');
        
        if (!departmentsView) {
            console.error('❌ departments-view 요소를 찾을 수 없습니다');
            return false;
        }
        
        departmentsView.classList.remove('hidden');
        if (navDepartments) {
            navDepartments.classList.add('active');
        }
        
        console.log('✅ 부서 관리 뷰 활성화');
        
        // 부서 목록 로드 (DepartmentManager가 있는 경우)
        if (typeof DepartmentManager !== 'undefined' && DepartmentManager.loadDepartmentsList) {
            DepartmentManager.loadDepartmentsList();
        }
        
        return true;
    },

    cleanupCurrentView: function() {
        // 현재 뷰에서 정리가 필요한 작업 수행
        switch(this.currentView) {
            case 'keys':
                // 키 뷰 정리 (예: 타이머 정리, 이벤트 리스너 제거 등)
                break;
            case 'users':
                // 사용자 뷰 정리
                break;
            case 'profile':
                // 프로필 뷰 정리
                break;
            case 'servers':
                // 서버 뷰 정리
                break;
            case 'departments':
                // 부서 뷰 정리
                break;
        }
        
        // 모달이 열려있으면 닫기
        if (typeof ModalManager !== 'undefined' && ModalManager.isOpen) {
            ModalManager.closeModal();
        }
        
        // 에러 메시지 클리어
        if (typeof AppUtils !== 'undefined' && AppUtils.clearError) {
            AppUtils.clearError();
        }
    },

    // 뷰 접근 권한 확인
    checkViewAccess: function(viewName) {
        // 로그인 상태 확인
        if (!AppState.jwtToken || !AppState.currentUser) {
            console.warn('로그인이 필요한 뷰 접근 시도:', viewName);
            return false;
        }
        
        // 특정 뷰에 대한 권한 확인
        switch(viewName) {
            case 'users':
                // 사용자 목록은 관리자만 접근 가능
                return this.isAdmin();
            case 'departments':
                // 부서 관리는 관리자만 접근 가능 (있는 경우)
                return this.isAdmin();
            case 'profile':
                // 프로필은 본인만 볼 수 있음
                return true;
            case 'keys':
                // 키 관리는 본인 키만 관리 가능
                return true;
            case 'servers':
                // 서버 관리는 모든 로그인 사용자 (구현에 따라 다름)
                return true;
            default:
                return true;
        }
    },

    // 관리자 권한 확인
    isAdmin: function() {
        return AppState.currentUser?.role === 'admin';
    },

    // 네비게이션 업데이트 (로그인 상태에 따라)
    updateNavigation: function() {
        const isAdmin = this.isAdmin();
        const usersNav = document.getElementById('nav-users');
        const departmentsNav = document.getElementById('nav-departments');
        
        console.log(`👤 사용자 권한: ${AppState.currentUser?.role || 'unknown'}, 관리자: ${isAdmin}`);
        
        // 사용자 목록 네비게이션
        if (usersNav) {
            if (isAdmin) {
                usersNav.style.display = 'inline-block';
                usersNav.title = '사용자 목록 관리';
                console.log('✅ 사용자 관리 버튼 표시');
            } else {
                usersNav.style.display = 'none';
                console.log('❌ 사용자 관리 버튼 숨김');
                
                // 현재 사용자 뷰에 있으면 키 관리로 이동
                if (this.currentView === 'users') {
                    console.log('🔄 사용자 뷰에서 키 뷰로 이동');
                    this.showView('keys');
                }
            }
        }
        
        // 부서 관리 네비게이션 (있는 경우)
        if (departmentsNav) {
            if (isAdmin) {
                departmentsNav.style.display = 'inline-block';
                departmentsNav.title = '부서 관리';
            } else {
                departmentsNav.style.display = 'none';
                
                // 현재 부서 뷰에 있으면 키 관리로 이동
                if (this.currentView === 'departments') {
                    this.showView('keys');
                }
            }
        }
    },

    // URL 해시를 기반으로 뷰 복원
    restoreViewFromHash: function() {
        const hash = window.location.hash.substring(1); // # 제거
        const validViews = ['keys', 'users', 'profile', 'servers', 'departments'];
        
        if (hash && validViews.includes(hash)) {
            this.showView(hash);
        } else {
            this.showView('keys'); // 기본값
        }
    },

    // URL 해시 업데이트
    updateUrlHash: function(viewName) {
        if (history.replaceState) {
            history.replaceState(null, null, `#${viewName}`);
        }
    },

    // 브라우저 뒤로/앞으로 가기 처리
    setupHistoryHandling: function() {
        window.addEventListener('popstate', () => {
            this.restoreViewFromHash();
        });
        
        window.addEventListener('hashchange', () => {
            this.restoreViewFromHash();
        });
        
        console.log('히스토리 처리 설정 완료');
    },

    // 반응형 뷰 처리
    handleResponsiveView: function() {
        const isMobile = window.innerWidth <= 768;
        
        if (isMobile) {
            // 모바일에서는 한 번에 하나의 뷰만 표시
            console.log('모바일 뷰 모드 활성화');
            document.body.classList.add('mobile-view');
        } else {
            // 데스크톱에서는 사이드바 등 추가 UI 요소 표시 가능
            console.log('데스크톱 뷰 모드 활성화');
            document.body.classList.remove('mobile-view');
        }
    },

    // 뷰 상태 저장/복원
    saveViewState: function() {
        const viewState = {
            currentView: this.currentView,
            timestamp: Date.now()
        };
        localStorage.setItem('viewState', JSON.stringify(viewState));
    },

    restoreViewState: function() {
        try {
            const savedState = localStorage.getItem('viewState');
            if (savedState) {
                const viewState = JSON.parse(savedState);
                // 1시간 이내의 상태만 복원
                if (Date.now() - viewState.timestamp < 3600000) {
                    return viewState.currentView;
                }
            }
        } catch (error) {
            console.warn('뷰 상태 복원 실패:', error);
        }
        return null;
    },

    // 뷰 히스토리 관리
    getViewHistory: function() {
        return JSON.parse(localStorage.getItem('viewHistory') || '[]');
    },

    addToViewHistory: function(viewName) {
        const history = this.getViewHistory();
        history.push({ view: viewName, timestamp: Date.now() });
        
        // 최근 10개만 유지
        if (history.length > 10) {
            history.shift();
        }
        
        localStorage.setItem('viewHistory', JSON.stringify(history));
    },

    // 이전 뷰로 돌아가기
    goToPreviousView: function() {
        const history = this.getViewHistory();
        if (history.length >= 2) {
            const previousView = history[history.length - 2];
            this.showView(previousView.view);
        }
    },

    // 초기화
    init: function() {
        console.log('ViewManager 초기화');
        
        // 히스토리 처리 설정
        this.setupHistoryHandling();
        
        // 반응형 처리
        window.addEventListener('resize', () => {
            this.handleResponsiveView();
        });
        
        // 초기 반응형 설정
        this.handleResponsiveView();
        
        // 페이지 언로드 시 뷰 상태 저장
        window.addEventListener('beforeunload', () => {
            this.saveViewState();
        });
        
        console.log('✅ ViewManager 초기화 완료');
    },

    // 디버깅용 메서드
    getCurrentViewInfo: function() {
        return {
            currentView: this.currentView,
            isAuthenticated: !!(AppState.jwtToken && AppState.currentUser),
            isAdmin: this.isAdmin(),
            availableViews: ['keys', 'users', 'profile', 'servers', 'departments'],
            userRole: AppState.currentUser?.role || 'none'
        };
    },

    // 강제 뷰 새로고침
    forceRefreshCurrentView: function() {
        const current = this.currentView;
        this.showView(current);
    }
};