window.ViewManager = {
    currentView: 'keys',
    init: function() {
        const navContainer = document.querySelector('.nav-buttons');
        if (navContainer) {
            navContainer.addEventListener('click', (e) => {
                const button = e.target.closest('.nav-btn');
                if (!button || !button.id) return;
                e.preventDefault();
                const viewName = button.id.replace('nav-', '');
                this.showView(viewName);
            });
        }
    },
    showView: function(viewName) {
        if (!this.checkViewAccess(viewName)) {
            return Utils.showToast('해당 기능에 접근할 권한이 없습니다.', 'warning');
        }
        document.querySelectorAll('.view-section').forEach(view => view.classList.add('hidden'));
        document.querySelectorAll('.nav-btn').forEach(btn => btn.classList.remove('active'));
        const viewElement = document.getElementById(`${viewName}-view`);
        const navElement = document.getElementById(`nav-${viewName}`);
        if (viewElement) viewElement.classList.remove('hidden');
        if (navElement) navElement.classList.add('active');
        this.currentView = viewName;
        this.initializeView(viewName);
    },
    initializeView: function(viewName) {
        switch(viewName) {
            case 'keys': KeyManager.autoLoadKeys(); break;
            case 'users': UserManager.loadUsersList(); break;
            case 'profile': ProfileManager.loadCurrentUserProfile(); break;
            case 'servers': ServerManager.loadServers(); break;
        }
    },
    checkViewAccess: function(viewName) {
        return !((viewName === 'users' || viewName === 'departments') && !AuthManager.isAdmin());
    },
    updateNavigation: function() {
        if (DOM.navUsers) DOM.navUsers.style.display = AuthManager.isAdmin() ? 'inline-flex' : 'none';
        if (DOM.navDepartments) DOM.navDepartments.style.display = AuthManager.isAdmin() ? 'inline-flex' : 'none';
        if (!this.checkViewAccess(this.currentView)) this.showView('keys');
    }
};