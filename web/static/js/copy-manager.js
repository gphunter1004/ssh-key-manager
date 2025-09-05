// 복사 기능 관리자
window.CopyManager = {
    setupEventListeners: function() {
        // 전역 클릭 이벤트 위임
        document.addEventListener('click', (e) => {
            // 복사 버튼 처리
            if (e.target.classList.contains('copy-btn') || e.target.classList.contains('copy-key-btn')) {
                e.preventDefault();
                this.handleCopyClick(e.target);
            }
            
            // data-action 기반 복사 처리
            if (e.target.dataset.action === 'copy') {
                e.preventDefault();
                this.handleActionCopy(e.target);
            }
        });
        
        console.log('CopyManager 이벤트 리스너 설정 완료');
    },

    handleCopyClick: function(button) {
        let textToCopy = '';
        let copyType = '';

        // 모달 내 키 복사 버튼
        if (button.classList.contains('copy-key-btn')) {
            const keyType = button.dataset.keyType;
            textToCopy = this.getKeyContent(keyType);
            copyType = this.getCopyTypeFromKeyType(keyType);
        }
        // 일반 복사 버튼
        else if (button.classList.contains('copy-btn')) {
            const targetId = button.dataset.target;
            const targetElement = document.getElementById(targetId);
            if (targetElement) {
                textToCopy = targetElement.textContent;
                copyType = this.getCopyTypeFromId(targetId);
            }
        }

        if (textToCopy) {
            this.copyToClipboard(textToCopy, copyType);
        }
    },

    handleActionCopy: function(element) {
        const targetId = element.dataset.target;
        const copyType = element.dataset.type || '텍스트';
        
        let textToCopy = '';
        
        if (element.dataset.text) {
            // 직접 텍스트가 지정된 경우
            textToCopy = element.dataset.text;
        } else if (targetId) {
            // 타겟 요소에서 텍스트 추출
            const targetElement = document.getElementById(targetId);
            if (targetElement) {
                textToCopy = targetElement.textContent;
            }
        }

        if (textToCopy) {
            this.copyToClipboard(textToCopy, copyType);
        }
    },

    getKeyContent: function(keyType) {
        // 모달에서 키 타입별 콘텐츠 추출
        const elementId = `modal-${keyType}-key`;
        const element = document.getElementById(elementId);
        return element ? element.textContent : '';
    },

    getCopyTypeFromKeyType: function(keyType) {
        const types = {
            'public': '공개키',
            'pem': 'PEM 키',
            'ppk': 'PPK 키'
        };
        return types[keyType] || '키';
    },

    getCopyTypeFromId: function(targetId) {
        if (targetId.includes('public')) return '공개키';
        if (targetId.includes('pem')) return 'PEM 키';
        if (targetId.includes('ppk')) return 'PPK 키';
        if (targetId.includes('cmd')) return '명령어';
        return '텍스트';
    },

    copyToClipboard: async function(text, type = '텍스트') {
        if (!text) {
            this.showCopyMessage('복사할 내용이 없습니다', 'warning');
            return;
        }

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
        Object.assign(textArea.style, {
            position: "fixed",
            top: "0",
            left: "0",
            width: "2em",
            height: "2em",
            padding: "0",
            border: "none",
            outline: "none",
            boxShadow: "none",
            background: "transparent",
            zIndex: "-1"
        });

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
        // Utils 매니저를 사용하여 토스트 메시지 표시
        if (typeof Utils !== 'undefined' && Utils.showToast) {
            Utils.showToast(`✅ ${type}가 클립보드에 복사되었습니다!`, 'success');
        } else {
            console.log(`✅ ${type} 복사 성공!`);
        }
    },

    showCopyError: function(type) {
        // Utils 매니저를 사용하여 토스트 메시지 표시
        if (typeof Utils !== 'undefined' && Utils.showToast) {
            Utils.showToast(`❌ ${type} 복사에 실패했습니다.`, 'error');
        } else {
            console.log(`❌ ${type} 복사 실패!`);
        }
    },
    
    // showCopyMessage는 더 이상 사용하지 않고 Utils.showToast로 대체
    showCopyMessage: function(message, type) {
        if (typeof Utils !== 'undefined' && Utils.showToast) {
            Utils.showToast(message, type);
        } else {
            console.log(`[${type}] ${message}`);
        }
    },
    
    // removeCopyMessage는 더 이상 사용하지 않음
    removeCopyMessage: function() {
        // 기능 대체됨
    },

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

    copyText: function(text, type) {
        this.copyToClipboard(text, type);
    },

    isCopySupported: function() {
        return !!(navigator.clipboard || document.queryCommandSupported('copy'));
    },

    isSecureContext: function() {
        return window.isSecureContext || location.protocol === 'https:' || location.hostname === 'localhost';
    },

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

    logCopyStatus: function() {
        const status = this.getCopyStatus();
        console.log('복사 기능 상태:', status);
        
        if (!status.supported) {
            console.warn('복사 기능이 지원되지 않습니다.');
        } else if (!status.modern) {
            console.warn('Fallback 복사 방식을 사용합니다. HTTPS 환경에서 최적의 경험을 위해 사용하세요.');
        }
    },

    setupKeyboardShortcuts: function() {
        document.addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 'c') {
                const selection = window.getSelection();
                if (!selection.toString()) {
                    const activeKeyElement = document.querySelector('.key-content:focus, .command-display:focus');
                    if (activeKeyElement) {
                        e.preventDefault();
                        this.copyToClipboard(activeKeyElement.textContent, '키 정보');
                    }
                }
            }
        });
    },

    init: function() {
        this.logCopyStatus();
        this.setupKeyboardShortcuts();
        this.setupEventListeners(); // setupEventListeners 호출을 init에 포함
        console.log('CopyManager 초기화 완료');
    }
};