// 뷰 관리자 - 완전 독립 버전 (네비게이션 완전 담당)
window.ViewManager = {
    currentView: 'keys',

    init: function() {
        console.log('📱 ViewManager 초기화 시작');
        
        // 네비게이션 이벤트 설정
        this.setupNavigationEvents();
        
        console.log('✅ ViewManager 초기화 완료');
        return true;
    },

    setupNavigationEvents: function() {
        console.log('🎯 ViewManager 네비게이션 이벤트 설정');
        
        // 네비게이션 버튼들
        const navButtons = [
            { element: DOM.navKeys, view: 'keys', name: '키 관리' },
            { element: DOM.navUsers, view: 'users', name: '사용자 목록' },
            { element: DOM.navProfile, view: 'profile', name: '프로필' },
            { element: DOM.navServers, view: 'servers', name: '서버 관리' },
            { element: DOM.navDepartments, view: 'departments', name: '부서 관리' }
        ];
        
        navButtons.forEach(({ element, view, name }) => {
            if (element) {
                element.onclick = (e) => {
                    e.preventDefault();
                    console.log(`🔘 ${name} 버튼 클릭`);
                    this.showView(view);
                };
                
                // 버튼 활성화 보장
                element.style.pointerEvents = 'auto';
                element.style.cursor = 'pointer';
                element.disabled = false;
                
                console.log(`✅ ${name} 이벤트 설정 완료`);
            } else {
                console.warn(`⚠️ ${name} 버튼 요소를 찾을 수 없음`);
            }
        });
        
        console.log('✅ ViewManager 모든 네비게이션 이벤트 설정 완료');
    },

    showView: function(viewName) {
        console.log(`🔄 뷰 전환 요청: ${this.currentView} → ${viewName}`);
        
        // 권한 확인
        if (!this.checkViewAccess(viewName)) {
            Utils.showToast('해당 기능에 접근할 권한이 없습니다', 'warning');
            return false;
        }
        
        // 이전 뷰 정리
        this.cleanupCurrentView();
        
        // 모든 뷰 숨기기
        this.hideAllViews();
        
        // 네비게이션 초기화
        this.resetNavigation();
        
        // 선택된 뷰 표시
        const success = this.displayView(viewName);
        
        if (success) {
            this.currentView = viewName;
            AppState.currentView = viewName;
            console.log(`✅ 뷰 전환 완료: ${viewName}`);
            return true;
        } else {
            console.warn(`❌ 뷰 전환 실패: ${viewName}`);
            // 실패 시 키 뷰로 fallback
            this.showView('keys');
            return false;
        }
    },

    displayView: function(viewName) {
        const viewElement = DOM[viewName + 'View'];
        const navElement = DOM['nav' + viewName.charAt(0).toUpperCase() + viewName.slice(1)];
        
        if (!viewElement) {
            console.error(`❌ ${viewName}-view 요소를 찾을 수 없음`);
            return false;
        }
        
        // 뷰 표시
        viewElement.classList.remove('hidden');
        console.log(`📱 ${viewName}-view 활성화`);
        
        // 네비게이션 활성화
        if (navElement) {
            navElement.classList.add('active');
            console.log(`🎯 nav-${viewName} 활성화`);
        }
        
        // 뷰별 초기화
        this.initializeView(viewName);
        
        return true;
    },

    initializeView: function(viewName) {
        console.log(`🚀 ${viewName} 뷰 초기화`);
        
        switch(viewName) {
            case 'keys':
                this.initializeKeysView();
                break;
            case 'users':
                this.initializeUsersView();
                break;
            case 'profile':
                this.initializeProfileView();
                break;
            case 'servers':
                this.initializeServersView();
                break;
            case 'departments':
                this.initializeDepartmentsView();
                break;
            default:
                console.warn(`알 수 없는 뷰: ${viewName}`);
        }
    },

    initializeKeysView: function() {
        // KeyManager에게 키 로드 요청
        if (typeof KeyManager !== 'undefined' && KeyManager.autoLoadKeys) {
            KeyManager.autoLoadKeys();
        } else {
            console.warn('⚠️ KeyManager를 찾을 수 없음');
        }
    },

    initializeUsersView: function() {
        // UserManager에게 사용자 목록 로드 요청
        if (typeof UserManager !== 'undefined' && UserManager.loadUsersList) {
            UserManager.loadUsersList();
        } else {
            console.warn('⚠️ UserManager를 찾을 수 없음');
        }
    },

    initializeProfileView: function() {
        // ProfileManager에게 프로필 로드 요청
        if (typeof ProfileManager !== 'undefined' && ProfileManager.loadCurrentUserProfile) {
            ProfileManager.loadCurrentUserProfile();
        } else {
            console.warn('⚠️ ProfileManager를 찾을 수 없음');
        }
    },

    initializeServersView: function() {
        // 서버 관리 기능 (준비 중)
        if (typeof ServerManager !== 'undefined' && ServerManager.loadServersList) {
            ServerManager.loadServersList();
        } else {
            Utils.showToast('서버 관리 기능은 준비 중입니다', 'info');
        }
    },

    initializeDepartmentsView: function() {
        // 부서 관리 기능 (준비 중)
        if (typeof DepartmentManager !== 'undefined' && DepartmentManager.loadDepartmentsList) {
            DepartmentManager.loadDepartmentsList();
        } else {
            Utils.showToast('부서 관리 기능은 준비 중입니다', 'info');
        }
    },

    hideAllViews: function() {
        const views = ['keys', 'users', 'profile', 'servers', 'departments'];
        views.forEach(view => {
            const viewElement = DOM[view + 'View'];
            if (viewElement) {
                viewElement.classList.add('hidden');
            }
        });
    },

    resetNavigation: function() {
        const navs = ['navKeys', 'navUsers', 'navProfile', 'navServers', 'navDepartments'];
        navs.forEach(nav => {
            const navElement = DOM[nav];
            if (navElement) {
                navElement.classList.remove('active');
            }
        });
    },

    cleanupCurrentView: function() {
        // 현재 뷰에서 정리가 필요한 작업 수행
        switch(this.currentView) {
            case 'keys':
                // 키 뷰 정리 작업
                break;
            case 'users':
                // 사용자 뷰 정리 작업
                break;
            case 'profile':
                // 프로필 뷰 정리 작업
                break;
        }
        
        // 모달이 열려있으면 닫기
        if (typeof ModalManager !== 'undefined' && ModalManager.isOpen) {
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
            console.warn(`❌ 미인증 상태에서 ${viewName} 뷰 접근 시도`);
            return false;
        }
        
        // 특정 뷰에 대한 권한 확인
        switch(viewName) {
            case 'users':
            case 'departments':
                // 관리자만 접근 가능
                if (AppState.currentUser.role !== 'admin') {
                    console.warn(`❌ 비관리자가 ${viewName} 뷰 접근 시도`);
                    return false;
                }
                break;
            case 'profile':
            case 'keys':
            case 'servers':
                // 모든 로그인 사용자 접근 가능
                break;
            default:
                console.warn(`❌ 알 수 없는 뷰: ${viewName}`);
                return false;
        }
        
        return true;
    },

    // 네비게이션 업데이트 (권한에 따른 버튼 표시/숨김)
    updateNavigation: function() {
        const isAdmin = AppState.currentUser?.role === 'admin';
        console.log(`👤 네비게이션 권한 업데이트 - 관리자: ${isAdmin}`);
        
        // 관리자 전용 버튼들
        const adminButtons = [
            { element: DOM.navUsers, name: '사용자 목록' },
            { element: DOM.navDepartments, name: '부서 관리' }
        ];
        
        adminButtons.forEach(({ element, name }) => {
            if (element) {
                if (isAdmin) {
                    element.style.display = 'inline-block';
                    console.log(`✅ ${name} 버튼 표시`);
                } else {
                    element.style.display = 'none';
                    console.log(`❌ ${name} 버튼 숨김`);
                }
            }
        });
        
        // 현재 뷰가 접근 불가능하면 키 뷰로 이동
        if (!this.checkViewAccess(this.currentView)) {
            console.log(`🔄 현재 뷰(${this.currentView}) 접근 불가, 키 뷰로 이동`);
            this.showView('keys');
        }
        
        console.log('✅ 네비게이션 권한 업데이트 완료');
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
        
        console.log('🔄 브라우저 히스토리 처리 설정 완료');
    },

    // 이전 뷰로 돌아가기
    goToPreviousView: function() {
        const viewHistory = this.getViewHistory();
        if (viewHistory.length >= 2) {
            const previousView = viewHistory[viewHistory.length - 2];
            this.showView(previousView.view);
        }
    },

    // 뷰 히스토리 관리
    getViewHistory: function() {
        try {
            return JSON.parse(localStorage.getItem('viewHistory') || '[]');
        } catch (error) {
            return [];
        }
    },

    addToViewHistory: function(viewName) {
        const history = this.getViewHistory();
        history.push({ view: viewName, timestamp: Date.now() });
        
        // 최근 10개만 유지
        if (history.length > 10) {
            history.shift();
        }
        
        try {
            localStorage.setItem('viewHistory', JSON.stringify(history));
        } catch (error) {
            console.warn('뷰 히스토리 저장 실패:', error);
        }
    },

    // 강제 뷰 새로고침
    refreshCurrentView: function() {
        const current = this.currentView;
        console.log(`🔄 현재 뷰(${current}) 새로고침`);
        this.showView(current);
    },

    // 현재 뷰 정보 가져오기 (디버깅용)
    getCurrentViewInfo: function() {
        return {
            currentView: this.currentView,
            isAuthenticated: !!(AppState.jwtToken && AppState.currentUser),
            userRole: AppState.currentUser?.role || 'none',
            canAccessUsers: this.checkViewAccess('users'),
            canAccessDepartments: this.checkViewAccess('departments')
        };
    }
};