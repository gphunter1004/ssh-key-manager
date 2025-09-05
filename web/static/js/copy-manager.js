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
            opacity: '0',
            cursor: 'pointer'
        });
        
        // 타입별 색상 설정
        if (type === 'success') {
            messageEl.style.backgroundColor = '#27ae60';
        } else if (type === 'error') {
            messageEl.style.backgroundColor = '#e74c3c';
        } else if (type === 'warning') {
            messageEl.style.backgroundColor = '#f39c12';
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

    // 특정 요소의 텍스트 복사 (외부에서 호출용)
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

    // 키보드 단축키 지원 (Ctrl+C)
    setupKeyboardShortcuts: function() {
        document.addEventListener('keydown', (e) => {
            // Ctrl+C 또는 Cmd+C 감지
            if ((e.ctrlKey || e.metaKey) && e.key === 'c') {
                // 텍스트가 선택되지 않은 상태에서만 처리
                const selection = window.getSelection();
                if (!selection.toString()) {
                    // 현재 활성화된 키 영역이 있으면 복사
                    const activeKeyElement = document.querySelector('.key-content:focus, .command-display:focus');
                    if (activeKeyElement) {
                        e.preventDefault();
                        this.copyToClipboard(activeKeyElement.textContent, '키 정보');
                    }
                }
            }
        });
    },

    // 초기화
    init: function() {
        this.logCopyStatus();
        this.setupKeyboardShortcuts();
        console.log('CopyManager 초기화 완료');
    }
};