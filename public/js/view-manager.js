// 뷰 관리자
window.ViewManager = {
    currentView: 'keys',

    showView: function(viewName) {
        console.log('뷰 전환:', this.currentView, '->', viewName);
        
        // 이전 뷰 정리
        this.cleanupCurrentView();
        
        // 모든 뷰 숨기기
        DOM.keysView.classList.add('hidden');
        DOM.usersView.classList.add('hidden');
        DOM.profileView.classList.add('hidden');
        
        // 네비게이션 버튼 활성화 상태 초기화
        DOM.navKeys.classList.remove('active');
        DOM.navUsers.classList.remove('active');
        DOM.navProfile.classList.remove('active');
        
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
            default:
                console.warn('알 수 없는 뷰:', viewName);
                this.showKeysView(); // 기본값으로 키 관리 뷰 표시
                return;
        }
        
        // 현재 뷰 상태 업데이트
        this.currentView = viewName;
        AppState.currentView = viewName;
        
        // URL 해시 업데이트 (선택사항)
        if (history.replaceState) {
            history.replaceState(null, null, `#${viewName}`);
        }
    },

    showKeysView: function() {
        DOM.keysView.classList.remove('hidden');
        DOM.navKeys.classList.add('active');
        console.log('키 관리 뷰 활성화');
        
        // 키 뷰 초기화 (필요한 경우)
        // KeyManager.refreshView(); // 필요시 주석 해제
    },

    showUsersView: function() {
        DOM.usersView.classList.remove('hidden');
        DOM.navUsers.classList.add('active');
        console.log('사용자 목록 뷰 활성화');
        
        // 사용자 목록 로드
        UserManager.loadUsersList();
    },

    showProfileView: function() {
        DOM.profileView.classList.remove('hidden');
        DOM.navProfile.classList.add('active');
        console.log('프로필 뷰 활성화');
        
        // 프로필 정보 로드
        ProfileManager.loadCurrentUserProfile();
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
        }
        
        // 모달이 열려있으면 닫기
        if (DOM.userDetailModal && DOM.userDetailModal.style.display !== 'none') {
            ModalManager.closeModal();
        }
        
        // 에러 메시지 클리어
        AppUtils.clearError();
    },

    // URL 해시를 기반으로 뷰 복원
    restoreViewFromHash: function() {
        const hash = window.location.hash.substring(1); // # 제거
        if (hash && ['keys', 'users', 'profile'].includes(hash)) {
            this.showView(hash);
        } else {
            this.showView('keys'); // 기본값
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
    },

    // 뷰 접근 권한 확인
    checkViewAccess: function(viewName) {
        // 로그인 상태 확인
        if (!AppState.jwtToken) {
            console.warn('로그인이 필요한 뷰 접근 시도:', viewName);
            return false;
        }
        
        // 특정 뷰에 대한 권한 확인 (필요한 경우)
        switch(viewName) {
            case 'users':
                // 사용자 목록은 모든 로그인 사용자가 볼 수 있음
                return true;
            case 'profile':
                // 프로필은 본인만 볼 수 있음
                return true;
            case 'keys':
                // 키 관리는 본인 키만 관리 가능
                return true;
            default:
                return true;
        }
    },

    // 반응형 뷰 처리
    handleResponsiveView: function() {
        const isMobile = window.innerWidth <= 768;
        
        if (isMobile) {
            // 모바일에서는 한 번에 하나의 뷰만 표시
            console.log('모바일 뷰 모드 활성화');
        } else {
            // 데스크톱에서는 사이드바 등 추가 UI 요소 표시 가능
            console.log('데스크톱 뷰 모드 활성화');
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
        
        // 페이지 로드 시 해시에서 뷰 복원
        if (AppState.jwtToken) {
            this.restoreViewFromHash();
        }
    }
};