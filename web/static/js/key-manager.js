// SSH 키 관리자 - 최종 완성본
window.KeyManager = {
    setupEventListeners: function() {
        // 키 관리 버튼 이벤트 설정
        document.addEventListener('click', (e) => {
            const action = e.target.dataset.action;
            if (!action || !action.startsWith('key-')) return;

            e.preventDefault();
            
            switch (action) {
                case 'key-create':
                    this.createKey();
                    break;
                case 'key-view':
                    this.viewKey();
                    break;
                case 'key-delete':
                    this.deleteKey();
                    break;
            }
        });
        
        console.log('✅ KeyManager 이벤트 리스너 설정 완료');
    },

    // 로그인 후 키 자동 로드
    autoLoadKeys: async function() {
        console.log('🔍 키 자동 로드 시도');
        try {
            await this.viewKey();
        } catch (error) {
            console.log('ℹ️ 키가 없거나 로드 실패:', error.message);
            this.setKeyInfoText('생성된 SSH 키가 없습니다. 키를 생성해주세요.');
        }
        
        // 🔥 성공/실패 관계없이 항상 버튼 활성화
        this.clearLoadingState();
    },

    createKey: async function() {
        console.log('🔑 SSH 키 생성 요청');
        
        if (!confirm('SSH 키를 생성하시겠습니까?\n\n• 기존 키가 있다면 새 키로 교체됩니다.\n• 생성된 키는 서버에 자동으로 배포될 수 있습니다.')) {
            return;
        }

        try {
            this.setLoadingState('키 생성 중...');
            
            const keyData = await AppUtils.apiFetch('/keys', 'POST');
            
            console.log('✅ SSH 키 생성 성공');
            this.displayKeys(keyData);
            
            if (Utils && Utils.showToast) {
                Utils.showToast('SSH 키가 성공적으로 생성되었습니다!', 'success');
            }
            
        } catch (error) {
            console.error('❌ SSH 키 생성 실패:', error.message);
            this.hideKeys();
        } finally {
            this.clearLoadingState();
        }
    },

    viewKey: async function() {
        console.log('👀 SSH 키 조회 요청');
        
        try {
            this.setLoadingState('키 조회 중...');
            
            const keyData = await AppUtils.apiFetch('/keys', 'GET');
            
            console.log('✅ SSH 키 조회 성공');
            this.displayKeys(keyData);
            
        } catch (error) {
            console.error('❌ SSH 키 조회 실패:', error.message);
            this.hideKeys();
            
            if (error.message.includes('키를 찾을 수 없습니다') || error.status === 404) {
                this.setKeyInfoText('생성된 SSH 키가 없습니다. 먼저 키를 생성해주세요.');
            } else {
                this.setKeyInfoText('키를 불러오는 중 오류가 발생했습니다.');
            }
            
            throw error; // autoLoadKeys에서 catch할 수 있도록
        } finally {
            this.clearLoadingState();
        }
    },

    deleteKey: async function() {
        console.log('🗑️ SSH 키 삭제 요청');
        
        if (!confirm('정말로 SSH 키를 삭제하시겠습니까?\n\n⚠️ 주의사항:\n• 이 작업은 되돌릴 수 없습니다.\n• 서버에서도 자동으로 키가 제거됩니다.\n• 기존 SSH 연결이 불가능해질 수 있습니다.')) {
            return;
        }

        try {
            this.setLoadingState('키 삭제 중...');
            
            await AppUtils.apiFetch('/keys', 'DELETE');
            
            console.log('✅ SSH 키 삭제 성공');
            this.hideKeys();
            
            if (Utils && Utils.showToast) {
                Utils.showToast('SSH 키가 성공적으로 삭제되었습니다.', 'success');
            }
            
        } catch (error) {
            console.error('❌ SSH 키 삭제 실패:', error.message);
        } finally {
            this.clearLoadingState();
        }
    },

    displayKeys: function(keyData) {
        console.log('📋 키 데이터 표시:', keyData);
        
        // 키 데이터 정규화
        const normalizedData = this.normalizeKeyData(keyData);
        
        // 키 정보 표시
        this.setKeyInfoText(`Algorithm: ${normalizedData.algorithm} / Bits: ${normalizedData.bits}`);
        
        // 키 내용 설정
        this.setElementText('key-public', normalizedData.publicKey);
        this.setElementText('key-pem', normalizedData.privateKeyPem);
        this.setElementText('key-ppk', normalizedData.privateKeyPpk);
        
        // 명령어 생성
        const commands = this.generateCommands(normalizedData);
        this.setElementText('cmd-public', commands.publicKey);
        this.setElementText('cmd-pem', commands.pem);
        this.setElementText('cmd-ppk', commands.ppk);
        
        if (normalizedData.publicKey) {
            this.setElementText('cmd-authorized-keys', 
                `echo '${this.escapeShell(normalizedData.publicKey)}' >> ~/.ssh/authorized_keys`);
        }
        
        // 키 표시 영역 보이기
        const keyDisplayArea = document.getElementById('key-display-area');
        if (keyDisplayArea) {
            keyDisplayArea.classList.remove('hidden');
        }
        
        console.log('✅ 키 정보 표시 완료');
    },

    normalizeKeyData: function(keyData) {
        // 다양한 API 응답 형태를 표준화
        const normalized = {
            algorithm: keyData.Algorithm || keyData.algorithm || 'RSA',
            bits: keyData.Bits || keyData.bits || keyData.key_size || 4096,
            publicKey: keyData.PublicKey || keyData.public_key || keyData.publicKey || keyData.pub || '',
            privateKeyPem: keyData.PEM || keyData.private_key_pem || keyData.privateKeyPem || keyData.pem || '',
            privateKeyPpk: keyData.PPK || keyData.private_key_ppk || keyData.privateKeyPpk || keyData.ppk || ''
        };
        
        return normalized;
    },

    generateCommands: function(normalizedData) {
        return {
            publicKey: normalizedData.publicKey ? `echo '${this.escapeShell(normalizedData.publicKey)}' > id_rsa.pub` : '',
            pem: normalizedData.privateKeyPem ? `echo '${this.escapeShell(normalizedData.privateKeyPem)}' > id_rsa` : '',
            ppk: normalizedData.privateKeyPpk ? `echo '${this.escapeShell(normalizedData.privateKeyPpk)}' > id_rsa.ppk` : ''
        };
    },

    escapeShell: function(str) {
        if (!str) return '';
        return str.replace(/'/g, "'\"'\"'");
    },

    hideKeys: function() {
        console.log('🙈 키 정보 숨김');
        const keyDisplayArea = document.getElementById('key-display-area');
        if (keyDisplayArea) {
            keyDisplayArea.classList.add('hidden');
        }
        this.setKeyInfoText('');
    },

    setLoadingState: function(message) {
        this.setKeyInfoText(message || '처리 중...', '#3498db', 'bold');
        
        // 키 관리 버튼들 비활성화
        const buttons = document.querySelectorAll('[data-action^="key-"]');
        buttons.forEach(btn => {
            btn.disabled = true;
            btn.style.opacity = '0.6';
        });
    },

    clearLoadingState: function() {
        console.log('🔧 KeyManager 로딩 상태 해제 및 버튼 활성화');
        
        // 🔥 키 관리 버튼들 항상 활성화
        const buttons = document.querySelectorAll('[data-action^="key-"]');
        buttons.forEach(btn => {
            btn.disabled = false;
            btn.style.opacity = '1';
            btn.style.pointerEvents = 'auto';
            btn.style.cursor = 'pointer';
        });
        
        console.log('✅ 모든 키 관련 버튼이 활성화되었습니다');
    },

    setKeyInfoText: function(text, color = '', fontWeight = '') {
        const keyInfo = document.getElementById('key-info');
        if (keyInfo) {
            keyInfo.textContent = text;
            keyInfo.style.color = color;
            keyInfo.style.fontWeight = fontWeight;
        }
    },

    setElementText: function(elementId, text) {
        const element = document.getElementById(elementId);
        if (element && text) {
            element.textContent = text;
        }
    },

    // 새로고침
    refresh: async function() {
        console.log('🔄 키 정보 새로고침');
        await this.viewKey();
    },

    // DOM 요소 존재 확인
    checkRequiredElements: function() {
        const requiredElements = [
            'key-info', 'key-display-area', 'key-public', 'key-pem', 'key-ppk',
            'cmd-public', 'cmd-authorized-keys', 'cmd-pem', 'cmd-ppk'
        ];

        const missingElements = requiredElements.filter(id => !document.getElementById(id));
        
        if (missingElements.length > 0) {
            console.warn('⚠️ 누락된 DOM 요소들:', missingElements);
            return false;
        }

        console.log('✅ 모든 필수 DOM 요소가 존재합니다');
        return true;
    },

    // 초기화
    init: function() {
        console.log('🔧 KeyManager 초기화 시작');
        
        // DOM 요소 존재 확인
        if (!this.checkRequiredElements()) {
            console.error('❌ 필수 DOM 요소가 누락되어 KeyManager 초기화 실패');
            return false;
        }

        // 이벤트 리스너 설정
        this.setupEventListeners();

        console.log('✅ KeyManager 초기화 완료');
        return true;
    }
};