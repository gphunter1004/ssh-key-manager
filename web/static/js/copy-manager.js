// 복사 기능 관리자 (최종 수정본)
window.CopyManager = {
    init: function() {
        console.log('📋 CopyManager 초기화 시작');
        this.setupEventListeners();
        return true;
    },

    setupEventListeners: function() {
        document.body.addEventListener('click', (e) => {
            const copyBtn = e.target.closest('.copy-btn, [data-action="copy"]');
            if (!copyBtn) return;

            e.preventDefault();
            const targetSelector = copyBtn.dataset.target;
            const targetElement = document.querySelector(targetSelector);
            
            if (targetElement) {
                this.copyToClipboard(targetElement.textContent.trim());
            } else {
                console.warn(`복사 대상 없음: ${targetSelector}`);
            }
        });
    },

    copyToClipboard: async function(text, type = '텍스트') {
        if (!text) {
            return Utils.showToast('복사할 내용이 없습니다.', 'warning');
        }

        try {
            await navigator.clipboard.writeText(text);
            Utils.showToast(`${type}가 클립보드에 복사되었습니다.`, 'success');
        } catch (err) {
            console.error('❌ 클립보드 복사 실패:', err);
            Utils.showToast('클립보드 복사에 실패했습니다.', 'error');
        }
    }
};