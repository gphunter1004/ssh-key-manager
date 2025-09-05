window.ModalManager = {
    currentOpenModal: null,
    init: function() {
        document.body.addEventListener('click', (e) => {
            if (e.target.closest('[data-action="close-modal"]') || e.target.classList.contains('modal')) {
                this.closeActiveModal();
            }
        });
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.currentOpenModal) this.closeActiveModal();
        });
        window.addEventListener('resize', () => this.adjustModalSize());
    },
    openModalWithSpinner: function(modalId, title, message = '데이터를 불러오는 중...') {
        const modal = document.getElementById(modalId);
        if (!modal) return;
        const titleEl = modal.querySelector('.modal-header h3');
        const contentEl = modal.querySelector('.modal-content-area');
        if (titleEl) titleEl.textContent = title;
        if (contentEl) contentEl.innerHTML = `<div class="spinner-container"><div class="spinner"></div><p>${message}</p></div>`;
        this.currentOpenModal = modal;
        modal.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
        this.adjustModalSize();
    },
    updateModalContent: function(title, contentHTML) {
        if (!this.currentOpenModal) return;
        const titleEl = this.currentOpenModal.querySelector('.modal-header h3');
        const contentEl = this.currentOpenModal.querySelector('.modal-content-area');
        if (titleEl) titleEl.innerHTML = title;
        if (contentEl) contentEl.innerHTML = contentHTML;
    },
    closeActiveModal: function() {
        if (this.currentOpenModal) {
            this.currentOpenModal.classList.add('hidden');
            const contentArea = this.currentOpenModal.querySelector('.modal-content-area');
            if (contentArea) contentArea.innerHTML = '';
            this.currentOpenModal = null;
            document.body.style.overflow = '';
        }
    },
    adjustModalSize: function() {
        if (this.currentOpenModal) {
            const modalContent = this.currentOpenModal.querySelector('.modal-content');
            if (modalContent) modalContent.style.maxHeight = `${window.innerHeight * 0.9}px`;
        }
    }
};