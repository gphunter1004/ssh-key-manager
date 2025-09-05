// 복사 기능 관리자 (HTTP 환경 호환 최종본)
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
            if (!targetSelector) {
                console.warn('Copy button is missing a "data-target" attribute.');
                return;
            }

            const targetElement = document.querySelector(targetSelector);
            if (targetElement) {
                this.copyToClipboard(targetElement.textContent.trim());
            } else {
                console.warn(`복사할 대상 요소를 찾을 수 없습니다: ${targetSelector}`);
                Utils.showToast('복사할 대상을 찾을 수 없습니다.', 'error');
            }
        });
    },

    /**
     * 텍스트를 클립보드에 복사합니다.
     * 보안 환경(HTTPS)에서는 Clipboard API를 사용하고,
     * 비보안 환경(HTTP)에서는 구형 execCommand 방식으로 대체 작동합니다.
     * @param {string} text - 복사할 텍스트
     * @param {string} type - 사용자에게 보여줄 데이터 타입 (예: '공개키')
     */
    copyToClipboard: async function(text, type = '텍스트') {
        if (!text) {
            return Utils.showToast('복사할 내용이 없습니다.', 'warning');
        }

        // navigator.clipboard API는 최신 브라우저 및 보안 컨텍스트(HTTPS)에서만 작동합니다.
        if (navigator.clipboard && window.isSecureContext) {
            try {
                await navigator.clipboard.writeText(text);
                Utils.showToast(`${type}가 클립보드에 복사되었습니다.`, 'success');
            } catch (err) {
                console.error('❌ Clipboard API 복사 실패:', err);
                Utils.showToast('클립보드 복사에 실패했습니다.', 'error');
            }
        } else {
            // 비보안 환경(HTTP) 또는 구형 브라우저를 위한 대체(Fallback) 방법
            console.log('비보안 환경 감지, 대체 복사 방식 사용');
            this.fallbackCopyTextToClipboard(text, type);
        }
    },

    /**
     * document.execCommand를 사용한 대체 복사 기능입니다.
     * @param {string} text - 복사할 텍스트
     * @param {string} type - 사용자에게 보여줄 데이터 타입
     */
    fallbackCopyTextToClipboard: function(text, type) {
        const textArea = document.createElement("textarea");
        textArea.value = text;

        // 화면에 보이지 않도록 스타일 설정
        textArea.style.position = "fixed";
        textArea.style.top = 0;
        textArea.style.left = 0;
        textArea.style.width = "2em";
        textArea.style.height = "2em";
        textArea.style.padding = 0;
        textArea.style.border = "none";
        textArea.style.outline = "none";
        textArea.style.boxShadow = "none";
        textArea.style.background = "transparent";

        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        try {
            const successful = document.execCommand('copy');
            if (successful) {
                Utils.showToast(`${type}가 클립보드에 복사되었습니다.`, 'success');
            } else {
                Utils.showToast('복사에 실패했습니다.', 'error');
            }
        } catch (err) {
            console.error('❌ Fallback 복사 실패:', err);
            Utils.showToast('이 브라우저에서는 복사 기능이 지원되지 않을 수 있습니다.', 'error');
        }

        document.body.removeChild(textArea);
    }
};