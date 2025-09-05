// 공통 유틸리티 함수들 (최종 수정본)
window.Utils = {
    showToast: function(message, type = 'success', duration = 3000) {
        const existingToast = document.getElementById('app-toast');
        if (existingToast) existingToast.remove();

        const toast = document.createElement('div');
        toast.id = 'app-toast';
        toast.className = `toast-${type}`;
        toast.textContent = message;

        document.body.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'toastOut 0.5s ease forwards';
            toast.addEventListener('animationend', () => toast.remove());
        }, duration);
    },

    setLoading: function(isLoading, message = '처리 중...') {
        this.hideLoadingIndicator();
        const elementsToDisable = document.querySelectorAll('button, input, a');

        if (isLoading) {
            document.body.style.cursor = 'wait';
            elementsToDisable.forEach(el => el.disabled = true);

            const indicator = document.createElement('div');
            indicator.id = 'loading-indicator';
            indicator.innerHTML = `<div class="spinner"></div><p>${message}</p>`;
            document.body.appendChild(indicator);
        } else {
            document.body.style.cursor = 'default';
            elementsToDisable.forEach(el => el.disabled = false);
        }
    },

    hideLoadingIndicator: function() {
        const indicator = document.getElementById('loading-indicator');
        if (indicator) indicator.remove();
    },

    formatDate: function(dateString, options = {}) {
        if (!dateString) return '정보 없음';
        const defaultOptions = { year: 'numeric', month: 'short', day: 'numeric', ...options };
        try {
            return new Date(dateString).toLocaleDateString('ko-KR', defaultOptions);
        } catch (error) { return dateString; }
    },

    escapeHtml: function(unsafe) {
        if (typeof unsafe !== 'string') return '';
        return unsafe.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
    }
};