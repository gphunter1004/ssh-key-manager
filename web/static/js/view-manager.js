// 뷰 관리자
window.ViewManager = {
    currentView: 'keys',

    setupEventListeners: function() {
        // 네비게이션 버튼 이벤트 설정
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
        console.log('뷰 전환:', this.currentView, '->', viewName);
        
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
        switch(viewName) {
            case 'keys':
                this.showKeysView();
                break;
            case 'users':
                this.showUsersView();
                break;
            case 'profile':
                this.showProfileView();
                break;
            case 'servers':
                this.showServersView();
                break;
            case 'departments':
                this.showDepartmentsView();
                break;
            default:
                console.warn('알 수 없는 뷰:', viewName);
                this.showKeysView(); // 기본값으로 키 관리 뷰 표시
                return;
        }
        
        // 현재 뷰 상태 업데이트
        this.currentView = viewName;
        AppState.currentView = viewName;
        
        // URL 해시 업데이트 (선택사항)
        this.updateUrlHash(viewName);
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
        if (DOM.keysView) {
            DOM.keysView.classList.remove('hidden');
        }
        if (DOM.navKeys) {
            DOM.navKeys.classList.add('active');
        }
        console.log('키 관리 뷰 활성화');
        
        // 키 뷰 초기화 - 자동 로드
        if (KeyManager && KeyManager.autoLoadKeys) {
            KeyManager.autoLoadKeys();
        }
    },

    showUsersView: function() {
        if (DOM.usersView) {
            DOM.usersView.classList.remove('hidden');
        }
        if (DOM.navUsers) {
            DOM.navUsers.classList.add('active');
        }
        console.log('사용자 목록 뷰 활성화');
        
        // 사용자 목록 로드
        if (UserManager && UserManager.loadUsersList) {
            UserManager.loadUsersList();
        }
    },

    showProfileView: function() {
        if (DOM.profileView) {
            DOM.profileView.classList.remove('hidden');
        }
        if (DOM.navProfile) {
            DOM.navProfile.classList.add('active');
        }
        console.log('프로필 뷰 활성화');
        
        // 프로필 정보 로드
        if (ProfileManager && ProfileManager.loadCurrentUserProfile) {
            ProfileManager.loadCurrentUserProfile();
        }
    },

    showServersView: function() {
        const serversView = document.getElementById('servers-view');
        const navServers = document.getElementById('nav-servers');
        
        if (serversView) {
            serversView.classList.remove('hidden');
        }
        if (navServers) {
            navServers.classList.add('active');
        }
        console.log('서버 관리 뷰 활성화');
        
        // 서버 목록 로드 (ServerManager가 있는 경우)
        if (window.ServerManager && ServerManager.loadServersList) {
            ServerManager.loadServersList();
        }
    },

    showDepartmentsView: function() {
        const departmentsView = document.getElementById('departments-view');
        const navDepartments = document.getElementById('nav-departments');
        
        if (departmentsView) {
            departmentsView.classList.remove('hidden');
        }
        if (navDepartments) {
            navDepartments.classList.add('active');
        }
        console.log('부서 관리 뷰 활성화');
        
        // 부서 목록 로드 (DepartmentManager가 있는 경우)
        if (window.DepartmentManager && DepartmentManager.loadDepartmentsList) {
            DepartmentManager.loadDepartmentsList();
        }
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
        if (ModalManager && ModalManager.isOpen) {
            ModalManager.closeModal();
        }
        
        // 에러 메시지 클리어
        if (AppUtils && AppUtils.clearError) {
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
        const usersNav = DOM.navUsers;
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

    // 뷰 전환 애니메이션 (선택사항)
    animateViewTransition: function(fromView, toView) {
        // 부드러운 전환 효과를 원하는 경우 구현
        console.log(`뷰 전환 애니메이션: ${fromView} → ${toView}`);
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
        
        console.log('ViewManager 초기화 완료');
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
    }
};