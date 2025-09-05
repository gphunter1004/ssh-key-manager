// SSH 키 관리자 - 최종 완성본
window.KeyManager = {
    init: function() {
        console.log('🔧 KeyManager 초기화 시작');
        this.setupEventListeners();
        
        // DOM 요소 존재 확인
        this.checkRequiredElements();
        
        console.log('✅ KeyManager 초기화 완료');
        return true;
    },

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
        
        Utils.setLoading(false);
    },

    createKey: async function() {
        console.log('🔑 SSH 키 생성 요청');
        
        if (!confirm('SSH 키를 생성하시겠습니까?\n\n• 기존 키가 있다면 새 키로 교체됩니다.\n• 생성된 키는 서버에 자동으로 배포될 수 있습니다.')) {
            return;
        }

        try {
            Utils.setLoading(true, '키 생성 중...');
            
            const keyData = await AppUtils.apiFetch('/keys', 'POST');
            
            console.log('✅ SSH 키 생성 성공');
            this.displayKeys(keyData);
            
            Utils.showToast('SSH 키가 성공적으로 생성되었습니다!', 'success');
            
        } catch (error) {
            console.error('❌ SSH 키 생성 실패:', error.message);
            this.hideKeys();
        } finally {
            Utils.setLoading(false);
        }
    },

    viewKey: async function() {
        console.log('👀 SSH 키 조회 요청');
        
        try {
            Utils.setLoading(true, '키 조회 중...');
            
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
            
            throw error;
        } finally {
            Utils.setLoading(false);
        }
    },

    deleteKey: async function() {
        console.log('🗑️ SSH 키 삭제 요청');
        
        if (!confirm('정말로 SSH 키를 삭제하시겠습니까?\n\n⚠️ 주의사항:\n• 이 작업은 되돌릴 수 없습니다.\n• 서버에서도 자동으로 키가 제거됩니다.\n• 기존 SSH 연결이 불가능해질 수 있습니다.')) {
            return;
        }

        try {
            Utils.setLoading(true, '키 삭제 중...');
            
            await AppUtils.apiFetch('/keys', 'DELETE');
            
            console.log('✅ SSH 키 삭제 성공');
            this.hideKeys();
            
            Utils.showToast('SSH 키가 성공적으로 삭제되었습니다.', 'success');
            
        } catch (error) {
            console.error('❌ SSH 키 삭제 실패:', error.message);
        } finally {
            Utils.setLoading(false);
        }
    },

    displayKeys: function(keyData) {
        console.log('📋 키 데이터 표시:', keyData);
        
        const normalizedData = this.normalizeKeyData(keyData);
        
        this.setKeyInfoText(`Algorithm: ${normalizedData.algorithm} / Bits: ${normalizedData.bits}`);
        
        this.setElementText('key-public', normalizedData.publicKey);
        this.setElementText('key-pem', normalizedData.privateKeyPem);
        this.setElementText('key-ppk', normalizedData.privateKeyPpk);
        
        const commands = this.generateCommands(normalizedData);
        this.setElementText('cmd-public', commands.publicKey);
        this.setElementText('cmd-pem', commands.pem);
        this.setElementText('cmd-ppk', commands.ppk);
        
        if (normalizedData.publicKey) {
            this.setElementText('cmd-authorized-keys', 
                `echo '${this.escapeShell(normalizedData.publicKey)}' >> ~/.ssh/authorized_keys`);
        }
        
        const keyDisplayArea = document.getElementById('key-display-area');
        if (keyDisplayArea) {
            keyDisplayArea.classList.remove('hidden');
        }
        
        console.log('✅ 키 정보 표시 완료');
    },

    normalizeKeyData: function(keyData) {
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
    
    // 셸 명령어 이스케이프 헬퍼
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

    refresh: async function() {
        console.log('🔄 키 정보 새로고침');
        await this.viewKey();
    },

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
    }
};