// SSH 키 관리자
window.KeyManager = {
    setupEventListeners: function() {
        // 키 생성 버튼
        document.getElementById('create-btn').addEventListener('click', this.createKey);
        
        // 키 조회 버튼
        document.getElementById('view-btn').addEventListener('click', this.viewKey);
        
        // 키 삭제 버튼
        document.getElementById('delete-btn').addEventListener('click', this.deleteKey);
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
            KeyManager.setLoadingState('키 생성 중...');
            
            const keyData = await AppUtils.apiFetch('/keys', 'POST');
            
            console.log('SSH 키 생성 성공:', {
                algorithm: keyData.Algorithm,
                bits: keyData.Bits,
                created: new Date().toLocaleString()
            });
            
            KeyManager.displayKeys(keyData);
            
            // 성공 메시지
            const message = keyData.message || 'SSH 키가 성공적으로 생성되었습니다!';
            alert(message);
            
        } catch (error) {
            console.error('SSH 키 생성 실패:', error.message);
            KeyManager.hideKeys();
            // 에러는 이미 AppUtils.apiFetch에서 표시됨
        } finally {
            KeyManager.clearLoadingState();
        }
    },

    viewKey: async function() {
        console.log('SSH 키 조회 요청');
        
        try {
            // 로딩 상태 표시
            KeyManager.setLoadingState('키 조회 중...');
            
            const keyData = await AppUtils.apiFetch('/keys', 'GET');
            
            console.log('SSH 키 조회 성공:', {
                algorithm: keyData.Algorithm,
                bits: keyData.Bits,
                updated: new Date(keyData.UpdatedAt).toLocaleString()
            });
            
            KeyManager.displayKeys(keyData);
            
        } catch (error) {
            console.error('SSH 키 조회 실패:', error.message);
            KeyManager.hideKeys();
            
            // 키가 없는 경우 특별 처리
            if (error.message.includes('키를 찾을 수 없습니다')) {
                DOM.keyInfo.textContent = '생성된 SSH 키가 없습니다. 먼저 키를 생성해주세요.';
            }
        } finally {
            KeyManager.clearLoadingState();
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
            KeyManager.setLoadingState('키 삭제 중...');
            
            const result = await AppUtils.apiFetch('/keys', 'DELETE');
            
            console.log('SSH 키 삭제 성공');
            
            // 성공 메시지 및 UI 업데이트
            alert(result.message || 'SSH 키가 성공적으로 삭제되었습니다.');
            KeyManager.hideKeys();
            
        } catch (error) {
            console.error('SSH 키 삭제 실패:', error.message);
            // 에러는 이미 AppUtils.apiFetch에서 표시됨
        } finally {
            KeyManager.clearLoadingState();
        }
    },

    displayKeys: function(keyData) {
        console.log('키 정보 표시 중...');
        
        // 키 정보 표시
        DOM.keyInfo.textContent = `Algorithm: ${keyData.Algorithm} / Bits: ${keyData.Bits}`;
        
        // 각 키 데이터 설정
        DOM.keyPublicPre.textContent = keyData.PublicKey;
        DOM.keyPemPre.textContent = keyData.PEM;
        DOM.keyPpkPre.textContent = keyData.PPK;
        
        // 명령어 생성
        const commands = KeyManager.generateCommands(keyData);
        DOM.cmdPublicPre.textContent = commands.publicKey;
        DOM.cmdAuthorizedKeysPre.textContent = commands.authorizedKeys;
        DOM.cmdPemPre.textContent = commands.pem;
        DOM.cmdPpkPre.textContent = commands.ppk;
        
        // 키 표시 영역 보이기
        DOM.keyDisplayArea.classList.remove('hidden');
        
        console.log('키 정보 표시 완료');
    },

    generateCommands: function(keyData) {
        // 안전한 문자열 처리를 위한 이스케이프
        const escapeShell = (str) => {
            return str.replace(/'/g, "'\"'\"'");
        };

        return {
            publicKey: `echo '${escapeShell(keyData.PublicKey)}' > id_rsa.pub`,
            authorizedKeys: `echo '${escapeShell(keyData.PublicKey)}' >> ~/.ssh/authorized_keys`,
            pem: `echo '${escapeShell(keyData.PEM)}' > id_rsa`,
            ppk: `echo '${escapeShell(keyData.PPK)}' > id_rsa.ppk`
        };
    },

    hideKeys: function() {
        console.log('키 정보 숨김');
        DOM.keyDisplayArea.classList.add('hidden');
        DOM.keyInfo.textContent = '';
    },

    setLoadingState: function(message) {
        DOM.keyInfo.textContent = message || '처리 중...';
        DOM.keyInfo.style.color = '#3498db';
        DOM.keyInfo.style.fontWeight = 'bold';
        
        // 버튼 비활성화
        const buttons = document.querySelectorAll('#key-controls button');
        buttons.forEach(btn => {
            btn.disabled = true;
            btn.style.opacity = '0.6';
        });
    },

    clearLoadingState: function() {
        DOM.keyInfo.style.color = '';
        DOM.keyInfo.style.fontWeight = '';
        
        // 버튼 활성화
        const buttons = document.querySelectorAll('#key-controls button');
        buttons.forEach(btn => {
            btn.disabled = false;
            btn.style.opacity = '1';
        });
    },

    // 키 정보 검증
    validateKeyData: function(keyData) {
        const requiredFields = ['Algorithm', 'Bits', 'PublicKey', 'PEM', 'PPK'];
        const missingFields = requiredFields.filter(field => !keyData[field]);
        
        if (missingFields.length > 0) {
            console.error('키 데이터 검증 실패, 누락된 필드:', missingFields);
            return false;
        }
        
        // 키 형식 기본 검증
        if (!keyData.PublicKey.startsWith('ssh-rsa')) {
            console.error('유효하지 않은 공개키 형식');
            return false;
        }
        
        if (!keyData.PEM.includes('BEGIN RSA PRIVATE KEY')) {
            console.error('유효하지 않은 PEM 형식');
            return false;
        }
        
        if (!keyData.PPK.includes('PuTTY-User-Key-File')) {
            console.error('유효하지 않은 PPK 형식');
            return false;
        }
        
        console.log('키 데이터 검증 통과');
        return true;
    },

    // 키 통계 정보 생성
    getKeyStats: function(keyData) {
        if (!keyData) return null;
        
        const publicKeyParts = keyData.PublicKey.split(' ');
        const keyType = publicKeyParts[0] || 'Unknown';
        const keyDataLength = publicKeyParts[1] ? publicKeyParts[1].length : 0;
        const comment = publicKeyParts.slice(2).join(' ') || 'No comment';
        
        return {
            type: keyType,
            algorithm: keyData.Algorithm,
            bits: keyData.Bits,
            dataLength: keyDataLength,
            comment: comment,
            createdAt: keyData.CreatedAt,
            updatedAt: keyData.UpdatedAt
        };
    },

    // 새로고침 (현재 키 상태 다시 로드)
    refresh: async function() {
        console.log('키 정보 새로고침');
        await this.viewKey();
    }
};