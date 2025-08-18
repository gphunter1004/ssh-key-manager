// 모달 관리자
window.ModalManager = {
    isOpen: false,
    currentModalData: null,

    setupEventListeners: function() {
        // 모달 닫기 버튼
        DOM.closeModalBtn.addEventListener('click', this.closeModal);
        
        // 모달 배경 클릭으로 닫기
        DOM.userDetailModal.addEventListener('click', (e) => {
            if (e.target === DOM.userDetailModal) {
                this.closeModal();
            }
        });
        
        // ESC 키로 모달 닫기
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen) {
                this.closeModal();
            }
        });
        
        console.log('ModalManager 이벤트 리스너 설정 완료');
    },

    showModal: function() {
        DOM.userDetailModal.style.display = 'block';
        this.isOpen = true;
        
        // 스크롤 방지
        document.body.style.overflow = 'hidden';
        
        // 모달 포커스
        DOM.userDetailModal.focus();
        
        console.log('모달 열림');
    },

    closeModal: function() {
        DOM.userDetailModal.style.display = 'none';
        this.isOpen = false;
        this.currentModalData = null;
        
        // 스크롤 복원
        document.body.style.overflow = '';
        
        // 모달 내용 초기화
        DOM.userDetailContent.innerHTML = '';
        
        console.log('모달 닫힘');
    },

    showLoadingModal: function(message = '로딩 중...') {
        DOM.userDetailContent.innerHTML = `
            <div class="modal-loading">
                <div class="loading-spinner"></div>
                <p>${message}</p>
            </div>
        `;
        this.showModal();
    },

    showErrorModal: function(message) {
        DOM.userDetailContent.innerHTML = `
            <div class="modal-error">
                <div class="error-icon">❌</div>
                <h3>오류 발생</h3>
                <p>${this.escapeHtml(message)}</p>
                <button type="button" class="retry-btn" onclick="ModalManager.closeModal()">
                    확인
                </button>
            </div>
        `;
        this.showModal();
    },

    displayUserDetail: function(userData) {
        this.currentModalData = userData;
        
        console.log('사용자 상세 정보 모달 표시:', userData.username);
        
        // 기본 사용자 정보
        const userInfo = this.generateUserInfo(userData);
        
        // SSH 키 정보
        const sshKeyInfo = this.generateSSHKeyInfo(userData);
        
        // 활동 정보
        const activityInfo = this.generateActivityInfo(userData);
        
        // 모달 제목 업데이트
        const modalTitle = document.querySelector('.modal-title');
        modalTitle.textContent = `${userData.username} 상세 정보`;
        
        // 모달 내용 구성
        DOM.userDetailContent.innerHTML = `
            <div class="user-detail-container">
                ${userInfo}
                ${sshKeyInfo}
                ${activityInfo}
                ${this.generateActionButtons(userData)}
            </div>
        `;
        
        this.showModal();
        
        // SSH 키 복사 버튼 이벤트 설정
        this.setupKeyActions(userData);
    },

    generateUserInfo: function(userData) {
        const createdDate = new Date(userData.created_at).toLocaleString('ko-KR');
        const updatedDate = new Date(userData.updated_at).toLocaleString('ko-KR');
        const isCurrentUser = UserManager.isCurrentUser(userData.id);
        
        return `
            <div class="user-basic-info">
                <h3>기본 정보 ${isCurrentUser ? '<span class="current-user-badge">현재 사용자</span>' : ''}</h3>
                <div class="info-grid">
                    <div class="info-item">
                        <strong>사용자명:</strong> 
                        <span class="username">${this.escapeHtml(userData.username)}</span>
                    </div>
                    <div class="info-item">
                        <strong>사용자 ID:</strong> ${userData.id}
                    </div>
                    <div class="info-item">
                        <strong>가입일:</strong> ${createdDate}
                    </div>
                    <div class="info-item">
                        <strong>마지막 업데이트:</strong> ${updatedDate}
                    </div>
                    <div class="info-item">
                        <strong>SSH 키 상태:</strong> 
                        <span class="key-status ${userData.has_ssh_key ? 'has-key' : 'no-key'}">
                            ${userData.has_ssh_key ? '🔑 보유' : '❌ 없음'}
                        </span>
                    </div>
                </div>
            </div>
        `;
    },

    generateSSHKeyInfo: function(userData) {
        if (!userData.has_ssh_key || !userData.ssh_key) {
            return `
                <div class="ssh-key-section">
                    <h3>SSH 키 정보</h3>
                    <div class="no-key-message">
                        <p>SSH 키가 생성되지 않았습니다.</p>
                        ${UserManager.isCurrentUser(userData.id) ? 
                            '<p><small>키 관리 탭에서 SSH 키를 생성할 수 있습니다.</small></p>' : 
                            ''}
                    </div>
                </div>
            `;
        }

        const keyCreated = new Date(userData.ssh_key.created_at).toLocaleString('ko-KR');
        const keyUpdated = new Date(userData.ssh_key.updated_at).toLocaleString('ko-KR');
        const isCurrentUser = UserManager.isCurrentUser(userData.id);
        
        // 다른 사용자의 키는 보안상 일부 정보만 표시
        if (!isCurrentUser) {
            return `
                <div class="ssh-key-section">
                    <h3>SSH 키 정보</h3>
                    <div class="ssh-key-info">
                        <div class="key-field">
                            <strong>알고리즘:</strong> ${userData.ssh_key.algorithm}
                        </div>
                        <div class="key-field">
                            <strong>키 크기:</strong> ${userData.ssh_key.bits} bits
                        </div>
                        <div class="key-field">
                            <strong>생성일:</strong> ${keyCreated}
                        </div>
                        <div class="key-field">
                            <strong>수정일:</strong> ${keyUpdated}
                        </div>
                        ${userData.ssh_key.fingerprint ? `
                        <div class="key-field">
                            <strong>핑거프린트:</strong> 
                            <code class="fingerprint">${userData.ssh_key.fingerprint}</code>
                        </div>
                        ` : ''}
                    </div>
                    <p class="security-notice">🔒 보안상 다른 사용자의 키 내용은 표시되지 않습니다.</p>
                </div>
            `;
        }

        // 현재 사용자의 키는 전체 정보 표시
        return `
            <div class="ssh-key-section">
                <h3>SSH 키 정보</h3>
                <div class="ssh-key-info">
                    <div class="key-field">
                        <strong>알고리즘:</strong> ${userData.ssh_key.algorithm}
                    </div>
                    <div class="key-field">
                        <strong>키 크기:</strong> ${userData.ssh_key.bits} bits
                    </div>
                    <div class="key-field">
                        <strong>생성일:</strong> ${keyCreated}
                    </div>
                    <div class="key-field">
                        <strong>수정일:</strong> ${keyUpdated}
                    </div>
                    ${userData.ssh_key.fingerprint ? `
                    <div class="key-field">
                        <strong>핑거프린트:</strong> 
                        <code class="fingerprint">${userData.ssh_key.fingerprint}</code>
                    </div>
                    ` : ''}
                </div>
                
                <div class="key-content-section">
                    <h4>공개키 (Public Key)</h4>
                    <div class="key-display-wrapper">
                        <pre class="key-content" id="modal-public-key">${userData.ssh_key.public_key}</pre>
                        <button type="button" class="copy-key-btn" data-key-type="public">
                            📋 복사
                        </button>
                    </div>
                    
                    <h4>개인키 (PEM Format)</h4>
                    <div class="key-display-wrapper">
                        <pre class="key-content" id="modal-pem-key">${userData.ssh_key.pem}</pre>
                        <button type="button" class="copy-key-btn" data-key-type="pem">
                            📋 복사
                        </button>
                    </div>
                    
                    <h4>개인키 (PPK Format)</h4>
                    <div class="key-display-wrapper">
                        <pre class="key-content" id="modal-ppk-key">${userData.ssh_key.ppk}</pre>
                        <button type="button" class="copy-key-btn" data-key-type="ppk">
                            📋 복사
                        </button>
                    </div>
                </div>
            </div>
        `;
    },

    generateActivityInfo: function(userData) {
        const createdDate = new Date(userData.created_at);
        const updatedDate = new Date(userData.updated_at);
        const now = new Date();
        
        const daysSinceCreated = Math.floor((now - createdDate) / (1000 * 60 * 60 * 24));
        const daysSinceUpdated = Math.floor((now - updatedDate) / (1000 * 60 * 60 * 24));
        
        let activityStatus = '';
        if (daysSinceUpdated === 0) {
            activityStatus = '🟢 오늘 활동';
        } else if (daysSinceUpdated <= 7) {
            activityStatus = '🟡 최근 활동';
        } else if (daysSinceUpdated <= 30) {
            activityStatus = '🟠 한 달 내 활동';
        } else {
            activityStatus = '🔴 비활성';
        }

        return `
            <div class="activity-info">
                <h3>활동 정보</h3>
                <div class="activity-grid">
                    <div class="activity-item">
                        <strong>가입 기간:</strong> ${daysSinceCreated}일
                    </div>
                    <div class="activity-item">
                        <strong>마지막 활동:</strong> ${daysSinceUpdated === 0 ? '오늘' : `${daysSinceUpdated}일 전`}
                    </div>
                    <div class="activity-item">
                        <strong>활동 상태:</strong> ${activityStatus}
                    </div>
                </div>
            </div>
        `;
    },

    generateActionButtons: function(userData) {
        const isCurrentUser = UserManager.isCurrentUser(userData.id);
        
        if (isCurrentUser) {
            return `
                <div class="modal-actions">
                    <button type="button" class="action-btn primary" onclick="ViewManager.showView('profile')">
                        ✏️ 프로필 편집
                    </button>
                    <button type="button" class="action-btn secondary" onclick="ViewManager.showView('keys')">
                        🔑 키 관리
                    </button>
                    <button type="button" class="action-btn secondary" onclick="ModalManager.closeModal()">
                        닫기
                    </button>
                </div>
            `;
        } else {
            return `
                <div class="modal-actions">
                    <button type="button" class="action-btn secondary" onclick="ModalManager.closeModal()">
                        닫기
                    </button>
                </div>
            `;
        }
    },

    setupKeyActions: function(userData) {
        if (!userData.has_ssh_key || !userData.ssh_key || !UserManager.isCurrentUser(userData.id)) {
            return;
        }

        // 키 복사 버튼 이벤트 설정
        const copyButtons = DOM.userDetailContent.querySelectorAll('.copy-key-btn');
        
        copyButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                const keyType = e.target.dataset.keyType;
                let keyContent = '';
                
                switch (keyType) {
                    case 'public':
                        keyContent = userData.ssh_key.public_key;
                        break;
                    case 'pem':
                        keyContent = userData.ssh_key.pem;
                        break;
                    case 'ppk':
                        keyContent = userData.ssh_key.ppk;
                        break;
                }
                
                if (keyContent) {
                    CopyManager.copyToClipboard(keyContent, `${keyType.toUpperCase()} 키`);
                }
            });
        });
    },

    // 모달 크기 조정
    adjustModalSize: function() {
        const modalContent = document.querySelector('.modal-content');
        const windowHeight = window.innerHeight;
        const maxHeight = windowHeight * 0.9;
        
        if (modalContent) {
            modalContent.style.maxHeight = `${maxHeight}px`;
        }
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

    // 모달 초기화
    init: function() {
        // 윈도우 리사이즈 시 모달 크기 조정
        window.addEventListener('resize', this.adjustModalSize);
        
        // 초기 모달 크기 설정
        this.adjustModalSize();
        
        console.log('ModalManager 초기화 완료');
    }
};