// 복사 기능 관리자
window.CopyManager = {
    setupEventListeners: function() {
        // 키 표시 영역의 복사 버튼들
        DOM.keyDisplayArea.addEventListener('click', this.handleKeyAreaCopy);
        
        console.log('CopyManager 이벤트 리스너 설정 완료');
    },

    handleKeyAreaCopy: function(e) {
        const target = e.target;
        let textToCopy = '';
        let copyType = '';

        if (target.classList.contains('copy-key-btn')) {
            const targetId = target.dataset.targetId;
            const targetElement = document.getElementById(targetId);
            if (targetElement) {
                textToCopy = targetElement.textContent;
                copyType = CopyManager.getCopyTypeFromId(targetId);
            }
        } else if (target.classList.contains('copy-cmd-btn')) {
            const targetId = target.dataset.targetId;
            const targetElement = document.getElementById(targetId);
            if (targetElement) {
                textToCopy = targetElement.textContent;
                copyType = '명령어';
            }
        }

        if (textToCopy) {
            CopyManager.copyToClipboard(textToCopy, copyType);
        }
    },

    getCopyTypeFromId: function(targetId) {
        if (targetId.includes('public')) return '공개키';
        if (targetId.includes('pem')) return 'PEM 키';
        if (targetId.includes('ppk')) return 'PPK 키';
        return '키';
    },

    copyToClipboard: async function(text, type = '텍스트') {
        try {
            // 현대적인 Clipboard API 사용 시도
            if (navigator.clipboard && window.isSecureContext) {
                await navigator.clipboard.writeText(text);
                this.showCopySuccess(type);
                console.log(`${type} 복사 성공 (Clipboard API)`);
            } else {
                // 폴백: 구형 브라우저나 HTTP 환경에서 사용
                this.fallbackCopyTextToClipboard(text, type);
            }
        } catch (error) {
            console.error('복사 실패:', error);
            this.showCopyError(type);
        }
    },

    fallbackCopyTextToClipboard: function(text, type) {
        console.log(`${type} 복사 시도 (Fallback 방식)`);
        
        const textArea = document.createElement("textarea");
        textArea.value = text;
        
        // 화면에 보이지 않도록 스타일 설정
        textArea.style.position = "fixed";
        textArea.style.top = "0";
        textArea.style.left = "0";
        textArea.style.width = "2em";
        textArea.style.height = "2em";
        textArea.style.padding = "0";
        textArea.style.border = "none";
        textArea.style.outline = "none";
        textArea.style.boxShadow = "none";
        textArea.style.background = "transparent";
        textArea.style.zIndex = "-1";

        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        try {
            const successful = document.execCommand('copy');
            if (successful) {
                this.showCopySuccess(type);
                console.log(`${type} 복사 성공 (Fallback)`);
            } else {
                throw new Error('Copy command failed');
            }
        } catch (err) {
            console.error('Fallback 복사 실패:', err);
            this.showCopyError(type);
        } finally {
            document.body.removeChild(textArea);
        }
    },

    showCopySuccess: function(type) {
        this.showCopyMessage(`✅ ${type}가 클립보드에 복사되었습니다!`, 'success');
    },

    showCopyError: function(type) {
        this.showCopyMessage(`❌ ${type} 복사에 실패했습니다.`, 'error');
    },

    showCopyMessage: function(message, type) {
        // 기존 메시지 제거
        this.removeCopyMessage();
        
        // 메시지 요소 생성
        const messageEl = document.createElement('div');
        messageEl.className = `copy-message copy-message-${type}`;
        messageEl.textContent = message;
        messageEl.id = 'copy-notification';
        
        // 스타일 설정
        Object.assign(messageEl.style, {
            position: 'fixed',
            top: '20px',
            right: '20px',
            padding: '12px 20px',
            borderRadius: '6px',
            color: 'white',
            fontWeight: 'bold',
            zIndex: '10000',
            fontSize: '14px',
            maxWidth: '300px',
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.3)',
            transform: 'translateX(100%)',
            transition: 'transform 0.3s ease-in-out, opacity 0.3s ease-in-out',
            opacity: '0'
        });
        
        // 타입별 색상 설정
        if (type === 'success') {
            messageEl.style.backgroundColor = '#27ae60';
        } else if (type === 'error') {
            messageEl.style.backgroundColor = '#e74c3c';
        }
        
        // DOM에 추가
        document.body.appendChild(messageEl);
        
        // 애니메이션 시작
        setTimeout(() => {
            messageEl.style.transform = 'translateX(0)';
            messageEl.style.opacity = '1';
        }, 10);
        
        // 자동 제거
        setTimeout(() => {
            this.removeCopyMessage();
        }, 3000);
        
        // 클릭으로 제거
        messageEl.addEventListener('click', () => {
            this.removeCopyMessage();
        });
    },

    removeCopyMessage: function() {
        const existingMessage = document.getElementById('copy-notification');
        if (existingMessage) {
            existingMessage.style.transform = 'translateX(100%)';
            existingMessage.style.opacity = '0';
            
            setTimeout(() => {
                if (existingMessage.parentNode) {
                    existingMessage.parentNode.removeChild(existingMessage);
                }
            }, 300);
        }
    },

    // 특정 요소의 텍스트 복사
    copyElementText: function(elementId, type) {
        const element = document.getElementById(elementId);
        if (!element) {
            console.error('복사할 요소를 찾을 수 없음:', elementId);
            this.showCopyError(type || '요소');
            return;
        }
        
        const text = element.textContent || element.innerText;
        this.copyToClipboard(text, type || '텍스트');
    },

    // 텍스트 직접 복사 (외부에서 호출용)
    copyText: function(text, type) {
        this.copyToClipboard(text, type);
    },

    // 복사 지원 여부 확인
    isCopySupported: function() {
        return !!(navigator.clipboard || document.queryCommandSupported('copy'));
    },

    // 보안 컨텍스트 확인
    isSecureContext: function() {
        return window.isSecureContext || location.protocol === 'https:' || location.hostname === 'localhost';
    },

    // 복사 기능 상태 확인
    getCopyStatus: function() {
        const hasClipboardAPI = !!(navigator.clipboard);
        const hasExecCommand = !!(document.queryCommandSupported && document.queryCommandSupported('copy'));
        const isSecure = this.isSecureContext();
        
        return {
            supported: hasClipboardAPI || hasExecCommand,
            modern: hasClipboardAPI && isSecure,
            fallback: hasExecCommand,
            secure: isSecure
        };
    },

    // 디버그 정보 출력
    logCopyStatus: function() {
        const status = this.getCopyStatus();
        console.log('복사 기능 상태:', status);
        
        if (!status.supported) {
            console.warn('복사 기능이 지원되지 않습니다.');
        } else if (!status.modern) {
            console.warn('Fallback 복사 방식을 사용합니다. HTTPS 환경에서 최적의 경험을 위해 사용하세요.');
        }
    },

    // 초기화
    init: function() {
        this.logCopyStatus();
        console.log('CopyManager 초기화 완료');
    }
};