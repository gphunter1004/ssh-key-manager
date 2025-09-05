// 모달 관리자 (최종 수정본 - 로직 단순화)
window.ModalManager = {
    currentOpenModal: null,

    init: function() {
        console.log('📱 ModalManager 초기화 시작');
        this.setupEventListeners();
        return true;
    },

    setupEventListeners: function() {
        document.body.addEventListener('click', (e) => {
            // '닫기' 버튼 또는 모달 배경 클릭 시 닫기
            if (e.target.closest('[data-action="close-modal"]') || e.target.classList.contains('modal')) {
                this.closeActiveModal();
            }
        });

        // ESC 키로 닫기
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.currentOpenModal) {
                this.closeActiveModal();
            }
        });

        // 창 크기 변경 시 모달 크기 조정
        window.addEventListener('resize', () => this.adjustModalSize());
    },

    /**
     * 지정된 ID의 모달을 열고 스피너와 함께 로딩 메시지를 표시합니다.
     * @param {string} modalId - 열 모달의 ID
     * @param {string} title - 모달 헤더에 표시될 제목
     * @param {string} message - 스피너 아래에 표시될 메시지
     */
    openModalWithSpinner: function(modalId, title, message = '데이터를 불러오는 중...') {
        const modal = document.getElementById(modalId);
        if (!modal) {
            console.error(`Modal element not found: #${modalId}`);
            return;
        }

        const titleEl = modal.querySelector('.modal-header h3');
        const contentEl = modal.querySelector('.modal-content-area');

        if (titleEl) titleEl.textContent = title;
        if (contentEl) {
            contentEl.innerHTML = `<div class="spinner-container"><div class="spinner"></div><p>${message}</p></div>`;
        }

        this.currentOpenModal = modal;
        modal.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
        this.adjustModalSize();
    },

    /**
     * 현재 열려 있는 모달의 내용을 업데이트합니다.
     * @param {string} title - 새로 설정할 모달 헤더 제목
     * @param {string} contentHTML - 모달의 콘텐츠 영역에 삽입할 HTML 문자열
     */
    updateModalContent: function(title, contentHTML) {
        if (!this.currentOpenModal) {
            console.warn('업데이트할 모달이 열려있지 않습니다.');
            return;
        }
        const titleEl = this.currentOpenModal.querySelector('.modal-header h3');
        const contentEl = this.currentOpenModal.querySelector('.modal-content-area');

        if (titleEl) titleEl.innerHTML = title;
        if (contentEl) contentEl.innerHTML = contentHTML;
    },
    
    /**
     * 현재 활성화된 모달을 닫습니다.
     */
    closeActiveModal: function() {
        if (this.currentOpenModal) {
            const modalId = this.currentOpenModal.id;
            this.currentOpenModal.classList.add('hidden');
            
            // 닫을 때 내용 비우기
            const contentArea = this.currentOpenModal.querySelector('.modal-content-area');
            if (contentArea) {
                contentArea.innerHTML = ''; 
            }

            this.currentOpenModal = null;
            document.body.style.overflow = ''; // 배경 스크롤 복원
            console.log(`모달 닫힘: #${modalId}`);
        }
    },

    /**
     * 창 크기에 맞게 모달의 최대 높이를 조정합니다.
     */
    adjustModalSize: function() {
        if (this.currentOpenModal) {
            const modalContent = this.currentOpenModal.querySelector('.modal-content');
            if (modalContent) {
                modalContent.style.maxHeight = `${window.innerHeight * 0.9}px`;
            }
        }
    }
};