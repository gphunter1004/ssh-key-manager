// 모달 관리자
window.ModalManager = {
    isOpen: false,
    currentModalData: null,

    setupEventListeners: function() {
        // 모달 닫기 버튼
        if (DOM.closeModalBtn) {
            DOM.closeModalBtn.addEventListener('click', this.closeModal);
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
        
        console.log('ModalManager 이벤트 리스너 설정 완료');
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
        
        ModalManager.isOpen = false;
        ModalManager.currentModalData = null;
        
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
                <div class="modal-error" style="text-align:center;padding:40px;">
                    <div class="error-icon" style="font-size:3rem;margin-bottom:20px;">❌</div>
                    <h3 style="color:#e74c3c;margin-bottom:15px;">오류 발생</h3>
                    <p style="color:#666;margin-bottom:20px;">${this.escapeHtml(message)}</p>
                    <button type="button" class="btn-primary" onclick="ModalManager.closeModal()">
                        확인
                    </button>
                </div>
            `;
        }
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
        const modalTitle = document.querySelector('.modal-header h3');
        if (modalTitle) {
            modalTitle.textContent = `${userData.username} 상세 정보`;
        }
        
        // 모달 내용 구성
        if (DOM.userDetailContent) {
            DOM.userDetailContent.innerHTML = `
                <div class="user-detail-container">
                    ${userInfo}
                    ${sshKeyInfo}
                    ${activityInfo}
                    ${this.generateActionButtons(userData)}
                </div>
            `;
        }
        
        this.showModal();
        
        // SSH 키 복사 버튼 이벤트 설정
        this.setupKeyActions(userData);
    },

    generateUserInfo: function(userData) {
        const createdDate = new Date(userData.created_at).toLocaleString('ko-KR');
        const updatedDate = new Date(userData.updated_at).toLocaleString('ko-KR');
        const isCurrentUser = UserManager.isCurrentUser(userData.id);
        const isAdmin = userData.role === 'admin';
        
        return `
            <div class="user-basic-info" style="margin-bottom:25px;">
                <h3 style="margin:0 0 20px 0;color:#212529;text-align:left;">
                    기본 정보 
                    ${isCurrentUser ? '<span class="current-user-badge" style="background:#007bff;color:white;padding:4px 8px;border-radius:12px;font-size:0.75rem;margin-left:10px;">현재 사용자</span>' : ''}
                    ${isAdmin ? '<span class="admin-badge" style="background:#ffc107;color:#212529;padding:4px 8px;border-radius:12px;font-size:0.75rem;margin-left:10px;">관리자</span>' : ''}
                </h3>
                <div class="info-grid" style="display:grid;gap:12px;">
                    <div class="info-item" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>사용자명:</strong> 
                        <span class="username">${this.escapeHtml(userData.username)}</span>
                    </div>
                    <div class="info-item" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>사용자 ID:</strong> ${userData.id}
                    </div>
                    <div class="info-item" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>역할:</strong> ${isAdmin ? '관리자' : '일반 사용자'}
                    </div>
                    <div class="info-item" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>가입일:</strong> ${createdDate}
                    </div>
                    <div class="info-item" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>마지막 업데이트:</strong> ${updatedDate}
                    </div>
                    <div class="info-item" style="display:flex;justify-content:space-between;padding:8px 0;">
                        <strong>SSH 키 상태:</strong> 
                        <span class="key-status ${userData.has_ssh_key ? 'has-key' : 'no-key'}" style="color:${userData.has_ssh_key ? '#28a745' : '#dc3545'};font-weight:600;">
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
                <div class="ssh-key-section" style="margin-bottom:25px;">
                    <h3 style="margin:0 0 15px 0;color:#212529;text-align:left;">SSH 키 정보</h3>
                    <div class="no-key-message" style="text-align:center;padding:20px;background:#f8f9fa;border-radius:8px;color:#666;">
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
                <div class="ssh-key-section" style="margin-bottom:25px;">
                    <h3 style="margin:0 0 15px 0;color:#212529;text-align:left;">SSH 키 정보</h3>
                    <div class="ssh-key-info">
                        <div class="key-field" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                            <strong>알고리즘:</strong> RSA
                        </div>
                        <div class="key-field" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                            <strong>키 크기:</strong> 4096 bits
                        </div>
                        <div class="key-field" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                            <strong>생성일:</strong> ${keyCreated}
                        </div>
                        <div class="key-field" style="display:flex;justify-content:space-between;padding:8px 0;">
                            <strong>수정일:</strong> ${keyUpdated}
                        </div>
                    </div>
                    <p class="security-notice" style="margin-top:15px;padding:10px;background:#fff9c4;border-left:4px solid #ffc107;border-radius:4px;color:#856404;font-size:0.9rem;">
                        🔒 보안상 다른 사용자의 키 내용은 표시되지 않습니다.
                    </p>
                </div>
            `;
        }

        // 현재 사용자의 키는 전체 정보 표시
        const keyData = userData.ssh_key;
        return `
            <div class="ssh-key-section" style="margin-bottom:25px;">
                <h3 style="margin:0 0 15px 0;color:#212529;text-align:left;">SSH 키 정보</h3>
                <div class="ssh-key-info" style="margin-bottom:20px;">
                    <div class="key-field" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>알고리즘:</strong> RSA
                    </div>
                    <div class="key-field" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>키 크기:</strong> 4096 bits
                    </div>
                    <div class="key-field" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>생성일:</strong> ${keyCreated}
                    </div>
                    <div class="key-field" style="display:flex;justify-content:space-between;padding:8px 0;">
                        <strong>수정일:</strong> ${keyUpdated}
                    </div>
                </div>
                
                <div class="key-content-section">
                    <h4 style="margin:20px 0 10px 0;color:#495057;">공개키 (Public Key)</h4>
                    <div class="key-display-wrapper" style="position:relative;margin-bottom:20px;">
                        <pre class="key-content" id="modal-public-key" style="background:#f8f9fa;border:1px solid #dee2e6;border-radius:4px;padding:12px;font-family:monospace;font-size:12px;white-space:pre-wrap;word-wrap:break-word;margin:0;padding-right:80px;">${keyData.public_key || keyData.PublicKey || ''}</pre>
                        <button type="button" class="copy-key-btn" data-key-type="public" style="position:absolute;top:8px;right:8px;background:#6c757d;color:white;border:none;padding:4px 8px;border-radius:3px;font-size:11px;cursor:pointer;">
                            📋 복사
                        </button>
                    </div>
                    
                    <h4 style="margin:20px 0 10px 0;color:#495057;">개인키 (PEM Format)</h4>
                    <div class="key-display-wrapper" style="position:relative;margin-bottom:20px;">
                        <pre class="key-content" id="modal-pem-key" style="background:#f8f9fa;border:1px solid #dee2e6;border-radius:4px;padding:12px;font-family:monospace;font-size:12px;white-space:pre-wrap;word-wrap:break-word;margin:0;padding-right:80px;">${keyData.pem || keyData.PEM || ''}</pre>
                        <button type="button" class="copy-key-btn" data-key-type="pem" style="position:absolute;top:8px;right:8px;background:#6c757d;color:white;border:none;padding:4px 8px;border-radius:3px;font-size:11px;cursor:pointer;">
                            📋 복사
                        </button>
                    </div>
                    
                    <h4 style="margin:20px 0 10px 0;color:#495057;">개인키 (PPK Format)</h4>
                    <div class="key-display-wrapper" style="position:relative;margin-bottom:20px;">
                        <pre class="key-content" id="modal-ppk-key" style="background:#f8f9fa;border:1px solid #dee2e6;border-radius:4px;padding:12px;font-family:monospace;font-size:12px;white-space:pre-wrap;word-wrap:break-word;margin:0;padding-right:80px;">${keyData.ppk || keyData.PPK || ''}</pre>
                        <button type="button" class="copy-key-btn" data-key-type="ppk" style="position:absolute;top:8px;right:8px;background:#6c757d;color:white;border:none;padding:4px 8px;border-radius:3px;font-size:11px;cursor:pointer;">
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
            <div class="activity-info" style="margin-bottom:25px;">
                <h3 style="margin:0 0 15px 0;color:#212529;text-align:left;">활동 정보</h3>
                <div class="activity-grid" style="display:grid;gap:12px;">
                    <div class="activity-item" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>가입 기간:</strong> ${daysSinceCreated}일
                    </div>
                    <div class="activity-item" style="display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #f8f9fa;">
                        <strong>마지막 활동:</strong> ${daysSinceUpdated === 0 ? '오늘' : `${daysSinceUpdated}일 전`}
                    </div>
                    <div class="activity-item" style="display:flex;justify-content:space-between;padding:8px 0;">
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
                <div class="modal-actions" style="display:flex;gap:10px;justify-content:center;padding-top:20px;border-top:1px solid #e9ecef;">
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
                <div class="modal-actions" style="display:flex;gap:10px;justify-content:center;padding-top:20px;border-top:1px solid #e9ecef;">
                    <button type="button" class="btn-secondary" onclick="ModalManager.closeModal()">
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
                
                const keyData = userData.ssh_key;
                switch (keyType) {
                    case 'public':
                        keyContent = keyData.public_key || keyData.PublicKey || '';
                        break;
                    case 'pem':
                        keyContent = keyData.pem || keyData.PEM || '';
                        break;
                    case 'ppk':
                        keyContent = keyData.ppk || keyData.PPK || '';
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