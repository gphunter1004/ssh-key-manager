// 모달 관리자
window.ModalManager = {
    isOpen: false,
    currentModalData: null,

    init: function() {
        console.log('📱 ModalManager 초기화 시작');
        
        // 이벤트 리스너 설정
        this.setupEventListeners();
        
        // 초기 모달 크기 설정
        this.adjustModalSize();
        
        console.log('✅ ModalManager 초기화 완료');
        return true;
    },

    setupEventListeners: function() {
        // 모달 닫기 버튼
        if (DOM.closeModalBtn) {
            DOM.closeModalBtn.addEventListener('click', () => this.closeModal());
        }
        
        // 모달 배경 클릭으로 닫기
        if (DOM.userDetailModal) {
            DOM.userDetailModal.addEventListener('click', (e) => {
                if (e.target === DOM.userDetailModal) {
                    this.closeModal();
                }
            });
        }
        
        // ESC 키로 모달 닫기
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen) {
                this.closeModal();
            }
        });
        
        // 윈도우 리사이즈 시 모달 크기 조정
        window.addEventListener('resize', () => this.adjustModalSize());
        
        console.log('✅ ModalManager 이벤트 리스너 설정 완료');
    },

    showModal: function() {
        if (DOM.userDetailModal) {
            DOM.userDetailModal.style.display = 'block';
            this.isOpen = true;
            
            // 스크롤 방지
            document.body.style.overflow = 'hidden';
            
            // 모달 포커스
            DOM.userDetailModal.focus();
            
            console.log('모달 열림');
        }
    },

    closeModal: function() {
        if (DOM.userDetailModal) {
            DOM.userDetailModal.style.display = 'none';
        }
        
        this.isOpen = false;
        this.currentModalData = null;
        
        // 스크롤 복원
        document.body.style.overflow = '';
        
        // 모달 내용 초기화
        if (DOM.userDetailContent) {
            DOM.userDetailContent.innerHTML = '';
        }
        
        console.log('모달 닫힘');
    },

    showLoadingModal: function(message = '로딩 중...') {
        if (DOM.userDetailContent) {
            DOM.userDetailContent.innerHTML = `
                <div class="modal-loading">
                    <div class="loading-spinner" style="width:40px;height:40px;border:4px solid #f3f3f3;border-top:4px solid #3498db;border-radius:50%;animation:spin 1s linear infinite;margin:0 auto 15px;"></div>
                    <p style="text-align:center;color:#666;">${message}</p>
                </div>
            `;
        }
        this.showModal();
    },

    showErrorModal: function(message) {
        if (DOM.userDetailContent) {
            DOM.userDetailContent.innerHTML = `
                <div class="modal-error">
                    <div class="error-icon">❌</div>
                    <h3>오류 발생</h3>
                    <p>${Utils.escapeHtml(message)}</p>
                    <button type="button" class="btn-primary" onclick="ModalManager.closeModal()">
                        확인
                    </button>
                </div>
            `;
        }
        this.showModal();
    },

    displayUserDetail: async function(userData) {
        this.currentModalData = userData;
        
        console.log('사용자 상세 정보 모달 표시:', userData.username);
        
        const userInfo = this.generateUserInfo(userData);
        const sshKeyInfo = await this.generateSSHKeyInfo(userData);
        const actionButtons = this.generateActionButtons(userData);
        
        const modalTitle = document.querySelector('.modal-header h3');
        if (modalTitle) {
            modalTitle.textContent = `${userData.username} 상세 정보`;
        }
        
        if (DOM.userDetailContent) {
            DOM.userDetailContent.innerHTML = `
                <div class="user-detail-container">
                    ${userInfo}
                    ${sshKeyInfo}
                    ${actionButtons}
                </div>
            `;
        }
        
        this.showModal();
    },

    generateUserInfo: function(userData) {
        const createdDate = userData.created_at ? Utils.formatDate(userData.created_at, {year: 'numeric', month: 'short', day: 'numeric'}) : '정보 없음';
        const updatedDate = userData.updated_at ? Utils.formatDate(userData.updated_at) : '정보 없음';
        const isCurrentUser = UserManager.isCurrentUser(userData.id);
        const isAdmin = userData.role === 'admin';
        
        return `
            <div class="user-basic-info">
                <h3 class="modal-subheader">
                    기본 정보 
                    ${isCurrentUser ? '<span class="current-user-badge">현재 사용자</span>' : ''}
                    ${isAdmin ? '<span class="admin-badge">관리자</span>' : ''}
                </h3>
                <div class="info-grid">
                    <div class="info-item">
                        <strong>사용자명:</strong> 
                        <span class="username">${Utils.escapeHtml(userData.username)}</span>
                    </div>
                    <div class="info-item">
                        <strong>사용자 ID:</strong> ${userData.id}
                    </div>
                    <div class="info-item">
                        <strong>역할:</strong> ${isAdmin ? '관리자' : '일반 사용자'}
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

    generateSSHKeyInfo: async function(userData) {
        if (!userData.has_ssh_key) {
            return `
                <div class="ssh-key-section">
                    <h3 class="modal-subheader">SSH 키 정보</h3>
                    <div class="no-key-message">
                        <p>SSH 키가 생성되지 않았습니다.</p>
                        ${UserManager.isCurrentUser(userData.id) ? 
                            '<p><small>키 관리 탭에서 SSH 키를 생성할 수 있습니다.</small></p>' : 
                            ''}
                    </div>
                </div>
            `;
        }

        const isCurrentUser = UserManager.isCurrentUser(userData.id);
        
        let keyData = {};
        if (isCurrentUser) {
            try {
                keyData = await AppUtils.apiFetch('/keys', 'GET');
            } catch (error) {
                console.error('❌ 현재 사용자 키 정보 로드 실패:', error.message);
                return `
                    <div class="ssh-key-section">
                        <h3 class="modal-subheader">SSH 키 정보</h3>
                        <p class="security-notice error">
                            ⚠️ 키 정보 로드 실패: ${Utils.escapeHtml(error.message)}
                        </p>
                    </div>
                `;
            }
        } else if (AuthManager.isAdmin()) {
            try {
                keyData = await AppUtils.apiFetch(`/admin/users/${userData.id}/keys`, 'GET');
            } catch (error) {
                console.error('❌ 관리자 키 정보 로드 실패:', error.message);
                return `
                    <div class="ssh-key-section">
                        <h3 class="modal-subheader">SSH 키 정보</h3>
                        <p class="security-notice warning">
                            🔒 키 정보 로드 실패 또는 접근 권한 없음: ${Utils.escapeHtml(error.message)}
                        </p>
                    </div>
                `;
            }
        }
        
        const keyCreated = Utils.formatDate(keyData.created_at || new Date().toISOString());
        const keyUpdated = Utils.formatDate(keyData.updated_at || new Date().toISOString());
        
        return `
            <div class="ssh-key-section">
                <h3 class="modal-subheader">SSH 키 정보</h3>
                <div class="ssh-key-info">
                    <div class="key-field">
                        <strong>알고리즘:</strong> ${Utils.escapeHtml(keyData.Algorithm || keyData.algorithm || 'RSA')}
                    </div>
                    <div class="key-field">
                        <strong>키 크기:</strong> ${keyData.Bits || keyData.bits || '4096'} bits
                    </div>
                    <div class="key-field">
                        <strong>생성일:</strong> ${keyCreated}
                    </div>
                    <div class="key-field">
                        <strong>수정일:</strong> ${keyUpdated}
                    </div>
                </div>
                
                ${isCurrentUser ? this.generateKeyContents(keyData) : this.generateAdminKeyContents(keyData)}
            </div>
        `;
    },

    generateKeyContents: function(keyData) {
        return `
            <div class="key-content-section">
                <h4 class="modal-key-title">공개키 (Public Key)</h4>
                <div class="key-display-wrapper">
                    <pre class="key-content" id="modal-public-key">${Utils.escapeHtml(keyData.PublicKey || keyData.public_key || '')}</pre>
                    <button type="button" class="copy-key-btn" data-key-type="public">
                        📋 복사
                    </button>
                </div>
                
                <h4 class="modal-key-title">개인키 (PEM Format)</h4>
                <div class="key-display-wrapper">
                    <pre class="key-content" id="modal-pem-key">${Utils.escapeHtml(keyData.PEM || keyData.private_key || '')}</pre>
                    <button type="button" class="copy-key-btn" data-key-type="pem">
                        📋 복사
                    </button>
                </div>
                
                <h4 class="modal-key-title">개인키 (PPK Format)</h4>
                <div class="key-display-wrapper">
                    <pre class="key-content" id="modal-ppk-key">${Utils.escapeHtml(keyData.PPK || keyData.ppk || '')}</pre>
                    <button type="button" class="copy-key-btn" data-key-type="ppk">
                        📋 복사
                    </button>
                </div>
            </div>
        `;
    },
    
    generateAdminKeyContents: function(keyData) {
        return `
            <div class="key-content-section">
                <h4 class="modal-key-title">공개키 (Public Key)</h4>
                <div class="key-display-wrapper">
                    <pre class="key-content" id="modal-public-key">${Utils.escapeHtml(keyData.PublicKey || keyData.public_key || '')}</pre>
                    <button type="button" class="copy-key-btn" data-key-type="public">
                        📋 복사
                    </button>
                </div>
                
                <p class="security-notice warning">
                    🔒 보안상 관리자도 다른 사용자의 개인키는 볼 수 없습니다.
                </p>
            </div>
        `;
    },

    generateActionButtons: function(userData) {
        const isCurrentUser = UserManager.isCurrentUser(userData.id);
        
        if (isCurrentUser) {
            return `
                <div class="modal-actions">
                    <button type="button" class="btn-primary" onclick="ViewManager.showView('profile'); ModalManager.closeModal();">
                        ✏️ 프로필 편집
                    </button>
                    <button type="button" class="btn-secondary" onclick="ViewManager.showView('keys'); ModalManager.closeModal();">
                        🔑 키 관리
                    </button>
                    <button type="button" class="btn-secondary" onclick="ModalManager.closeModal()">
                        닫기
                    </button>
                </div>
            `;
        } else {
            return `
                <div class="modal-actions">
                    <button type="button" class="btn-secondary" onclick="ModalManager.closeModal()">
                        닫기
                    </button>
                </div>
            `;
        }
    },
    
    adjustModalSize: function() {
        const modalContent = document.querySelector('.modal-content');
        const windowHeight = window.innerHeight;
        const maxHeight = windowHeight * 0.9;
        
        if (modalContent) {
            modalContent.style.maxHeight = `${maxHeight}px`;
        }
    }
};