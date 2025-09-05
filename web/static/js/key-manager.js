// SSH 키 관리자
window.KeyManager = {
    setupEventListeners: function() {
        // 액션 기반 이벤트 위임 (통합 스크립트의 data-action 방식 사용)
        document.addEventListener('click', (e) => {
            const action = e.target.dataset.action;
            if (!action) return;

            switch (action) {
                case 'key-create':
                    e.preventDefault();
                    this.createKey();
                    break;
                case 'key-view':
                    e.preventDefault();
                    this.viewKey();
                    break;
                case 'key-delete':
                    e.preventDefault();
                    this.deleteKey();
                    break;
            }
        });
        
        console.log('KeyManager 이벤트 리스너 설정 완료');
    },

    // 로그인 후 키 자동 로드
    autoLoadKeys: async function() {
        console.log('🔍 키 자동 로드 시도');
        try {
            await this.viewKey();
        } catch (error) {
            // 키가 없는 경우는 정상적인 상황이므로 에러 로그만 출력
            console.log('ℹ️ 키가 없거나 로드 실패:', error.message);
            DOM.keyInfo.textContent = '생성된 SSH 키가 없습니다. 키를 생성해주세요.';
        }
    },

    createKey: async function() {
        console.log('SSH 키 생성 요청');
        
        // 사용자 확인
        const confirmCreate = confirm(
            'SSH 키를 생성하시겠습니까?\n\n' +
            '• 기존 키가 있다면 새 키로 교체됩니다.\n' +
            '• 생성된 키는 서버에 자동으로 설치될 수 있습니다.'
        );
        
        if (!confirmCreate) {
            console.log('키 생성 취소됨');
            return;
        }

        try {
            // 로딩 상태 표시
            this.setLoadingState('키 생성 중...');
            
            const keyData = await AppUtils.apiFetch('/keys', 'POST');
            
            console.log('SSH 키 생성 성공:', keyData);
            
            this.displayKeys(keyData);
            
            // 성공 메시지
            Utils.showToast('SSH 키가 성공적으로 생성되었습니다!', 'success');
            
        } catch (error) {
            console.error('SSH 키 생성 실패:', error.message);
            this.hideKeys();
            // 에러는 이미 AppUtils.apiFetch에서 표시됨
        } finally {
            this.clearLoadingState();
        }
    },

    viewKey: async function() {
        console.log('SSH 키 조회 요청');
        
        try {
            // 로딩 상태 표시
            this.setLoadingState('키 조회 중...');
            
            const keyData = await AppUtils.apiFetch('/keys', 'GET');
            
            console.log('SSH 키 조회 성공:', keyData);
            
            this.displayKeys(keyData);
            
        } catch (error) {
            console.error('SSH 키 조회 실패:', error.message);
            this.hideKeys();
            
            // 키가 없는 경우 특별 처리
            if (error.message.includes('키를 찾을 수 없습니다') || error.status === 404) {
                DOM.keyInfo.textContent = '생성된 SSH 키가 없습니다. 먼저 키를 생성해주세요.';
            } else {
                DOM.keyInfo.textContent = '키를 불러오는 중 오류가 발생했습니다.';
            }
            throw error; // autoLoadKeys에서 catch할 수 있도록
        } finally {
            this.clearLoadingState();
        }
    },

    deleteKey: async function() {
        console.log('SSH 키 삭제 요청');
        
        // 사용자 확인
        const confirmDelete = confirm(
            '정말로 SSH 키를 삭제하시겠습니까?\n\n' +
            '⚠️ 주의사항:\n' +
            '• 이 작업은 되돌릴 수 없습니다.\n' +
            '• 서버에서도 자동으로 키가 제거됩니다.\n' +
            '• 기존 SSH 연결이 불가능해질 수 있습니다.'
        );
        
        if (!confirmDelete) {
            console.log('키 삭제 취소됨');
            return;
        }

        try {
            // 로딩 상태 표시
            this.setLoadingState('키 삭제 중...');
            
            const result = await AppUtils.apiFetch('/keys', 'DELETE');
            
            console.log('SSH 키 삭제 성공');
            
            // 성공 메시지 및 UI 업데이트
            Utils.showToast('SSH 키가 성공적으로 삭제되었습니다.', 'success');
            this.hideKeys();
            
        } catch (error) {
            console.error('SSH 키 삭제 실패:', error.message);
            // 에러는 이미 AppUtils.apiFetch에서 표시됨
        } finally {
            this.clearLoadingState();
        }
    },

    displayKeys: function(keyData) {
        console.log('📋 키 데이터 수신:', keyData);
        
        // API 응답 구조 정규화 (통합 스크립트의 normalizeKeyData 로직 적용)
        const normalizedData = this.normalizeKeyData(keyData);
        
        // 키 정보 표시
        DOM.keyInfo.textContent = `Algorithm: ${normalizedData.algorithm} / Bits: ${normalizedData.bits}`;
        
        // 각 키 데이터 설정
        if (DOM.keyPublicPre && normalizedData.publicKey) {
            DOM.keyPublicPre.textContent = normalizedData.publicKey;
        }
        if (DOM.keyPemPre && normalizedData.privateKeyPem) {
            DOM.keyPemPre.textContent = normalizedData.privateKeyPem;
        }
        if (DOM.keyPpkPre && normalizedData.privateKeyPpk) {
            DOM.keyPpkPre.textContent = normalizedData.privateKeyPpk;
        }
        
        // 명령어 생성
        const commands = this.generateCommands(normalizedData);
        if (DOM.cmdPublicPre) {
            DOM.cmdPublicPre.textContent = commands.publicKey;
        }
        if (DOM.cmdAuthorizedKeysPre && normalizedData.publicKey) {
            DOM.cmdAuthorizedKeysPre.textContent = `echo '${this.escapeShell(normalizedData.publicKey)}' >> ~/.ssh/authorized_keys`;
        }
        if (DOM.cmdPemPre) {
            DOM.cmdPemPre.textContent = commands.pem;
        }
        if (DOM.cmdPpkPre) {
            DOM.cmdPpkPre.textContent = commands.ppk;
        }
        
        // 키 표시 영역 보이기
        if (DOM.keyDisplayArea) {
            DOM.keyDisplayArea.classList.remove('hidden');
        }
        
        console.log('✅ 키 정보 표시 완료');
    },

    // API 응답 데이터 정규화 (다양한 응답 형태를 표준화)
    normalizeKeyData: function(keyData) {
        console.log('🔧 정규화 전 데이터:', keyData);
        
        // 다양한 API 응답 형태를 표준화
        const normalized = {
            algorithm: keyData.Algorithm || keyData.algorithm || 'RSA',
            bits: keyData.Bits || keyData.bits || keyData.key_size || 4096,
            publicKey: keyData.PublicKey || keyData.public_key || keyData.publicKey || keyData.pub || '',
            privateKeyPem: keyData.PEM || keyData.private_key_pem || keyData.privateKeyPem || keyData.pem || '',
            privateKeyPpk: keyData.PPK || keyData.private_key_ppk || keyData.privateKeyPpk || keyData.ppk || ''
        };
        
        console.log('🔧 정규화 후 데이터:', normalized);
        
        // 빈 값 체크
        if (!normalized.publicKey) {
            console.warn('⚠️ 공개키가 비어있습니다');
        }
        if (!normalized.privateKeyPem) {
            console.warn('⚠️ PEM 개인키가 비어있습니다');
        }
        if (!normalized.privateKeyPpk) {
            console.warn('⚠️ PPK 개인키가 비어있습니다');
        }
        
        return normalized;
    },

    generateCommands: function(normalizedData) {
        return {
            publicKey: normalizedData.publicKey ? `echo '${this.escapeShell(normalizedData.publicKey)}' > id_rsa.pub` : '',
            pem: normalizedData.privateKeyPem ? `echo '${this.escapeShell(normalizedData.privateKeyPem)}' > id_rsa` : '',
            ppk: normalizedData.privateKeyPpk ? `echo '${this.escapeShell(normalizedData.privateKeyPpk)}' > id_rsa.ppk` : ''
        };
    },

    // 안전한 문자열 처리를 위한 이스케이프
    escapeShell: function(str) {
        if (!str) return '';
        return str.replace(/'/g, "'\"'\"'");
    },

    hideKeys: function() {
        console.log('키 정보 숨김');
        if (DOM.keyDisplayArea) {
            DOM.keyDisplayArea.classList.add('hidden');
        }
        if (DOM.keyInfo) {
            DOM.keyInfo.textContent = '';
        }
    },

    setLoadingState: function(message) {
        if (DOM.keyInfo) {
            DOM.keyInfo.textContent = message || '처리 중...';
            DOM.keyInfo.style.color = '#3498db';
            DOM.keyInfo.style.fontWeight = 'bold';
        }
        
        // 버튼 비활성화
        const buttons = document.querySelectorAll('[data-action^="key-"]');
        buttons.forEach(btn => {
            btn.disabled = true;
            btn.style.opacity = '0.6';
        });
    },

    clearLoadingState: function() {
        if (DOM.keyInfo) {
            DOM.keyInfo.style.color = '';
            DOM.keyInfo.style.fontWeight = '';
        }
        
        // 버튼 활성화
        const buttons = document.querySelectorAll('[data-action^="key-"]');
        buttons.forEach(btn => {
            btn.disabled = false;
            btn.style.opacity = '1';
        });
    },

    // 키 정보 검증
    validateKeyData: function(keyData) {
        if (!keyData) return false;
        
        const normalizedData = this.normalizeKeyData(keyData);
        
        // 기본 검증
        if (!normalizedData.publicKey) {
            console.error('유효하지 않은 공개키');
            return false;
        }
        
        if (normalizedData.publicKey && !normalizedData.publicKey.startsWith('ssh-rsa')) {
            console.error('유효하지 않은 공개키 형식');
            return false;
        }
        
        if (normalizedData.privateKeyPem && !normalizedData.privateKeyPem.includes('BEGIN RSA PRIVATE KEY')) {
            console.error('유효하지 않은 PEM 형식');
            return false;
        }
        
        if (normalizedData.privateKeyPpk && !normalizedData.privateKeyPpk.includes('PuTTY-User-Key-File')) {
            console.error('유효하지 않은 PPK 형식');
            return false;
        }
        
        console.log('✅ 키 데이터 검증 통과');
        return true;
    },

    // 새로고침 (현재 키 상태 다시 로드)
    refresh: async function() {
        console.log('키 정보 새로고침');
        await this.viewKey();
    }
};