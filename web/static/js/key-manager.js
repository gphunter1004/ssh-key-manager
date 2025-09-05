// SSH 키 관리자 (최종 수정본)
window.KeyManager = {
    init: function() {
        console.log('🔧 KeyManager 초기화 시작');
        this.setupEventListeners();
        return true;
    },

    setupEventListeners: function() {
        document.addEventListener('click', (e) => {
            const action = e.target.dataset.action;
            if (!action || !action.startsWith('key-')) return;
            e.preventDefault();
            
            if (action === 'key-create') this.createKey();
            else if (action === 'key-view') this.viewKey();
            else if (action === 'key-delete') this.deleteKey();
        });
    },

    autoLoadKeys: async function() {
        await this.viewKey();
    },

    createKey: async function() {
        if (!confirm('새로운 SSH 키를 생성하시겠습니까?\n기존 키가 있다면 덮어쓰여집니다.')) return;
        Utils.setLoading(true, '키 생성 중...');
        try {
            const keyData = await AppUtils.apiFetch('/keys', 'POST');
            this.displayKeys(keyData);
            Utils.showToast('SSH 키가 성공적으로 생성되었습니다!', 'success');
        } catch (error) {
            this.hideKeys();
        } finally {
            Utils.setLoading(false);
        }
    },

    viewKey: async function() {
        Utils.setLoading(true, '키 정보 조회 중...');
        try {
            const keyData = await AppUtils.apiFetch('/keys', 'GET');
            this.displayKeys(keyData);
        } catch (error) {
            DOM.keyInfo.textContent = error.status === 404
                ? '생성된 SSH 키가 없습니다. 키를 생성해주세요.'
                : '키 정보를 불러오는 중 오류가 발생했습니다.';
            this.hideKeys();
        } finally {
            Utils.setLoading(false);
        }
    },

    deleteKey: async function() {
        if (!confirm('정말로 SSH 키를 삭제하시겠습니까?\n이 작업은 되돌릴 수 없습니다.')) return;
        Utils.setLoading(true, '키 삭제 중...');
        try {
            await AppUtils.apiFetch('/keys', 'DELETE');
            this.hideKeys();
            DOM.keyInfo.textContent = 'SSH 키가 삭제되었습니다. 필요 시 다시 생성하세요.';
            Utils.showToast('SSH 키가 성공적으로 삭제되었습니다.', 'success');
        } finally {
            Utils.setLoading(false);
        }
    },

    displayKeys: function(keyData) {
        DOM.keyInfo.textContent = `키 알고리즘: ${keyData.Algorithm || 'RSA'} / ${keyData.Bits || 4096} bits`;
        DOM.keyDisplayArea.innerHTML = `
            <div class="key-item">
                <h4>공개키 (Public Key)</h4>
                <div class="command-display">
                    <pre>echo '${keyData.PublicKey}' >> ~/.ssh/authorized_keys</pre>
                    <button class="copy-btn" data-target="#key-public-content">복사</button>
                </div>
                <div class="key-content"><pre id="key-public-content">${keyData.PublicKey}</pre></div>
            </div>
            <div class="key-item">
                <h4>개인키 (PEM Format)</h4>
                <div class="key-content">
                    <pre id="key-pem-content">${keyData.PEM}</pre>
                    <button class="copy-btn" data-target="#key-pem-content">복사</button>
                </div>
            </div>
            <div class="key-item">
                <h4>개인키 (PPK Format for PuTTY)</h4>
                <div class="key-content">
                    <pre id="key-ppk-content">${keyData.PPK}</pre>
                    <button class="copy-btn" data-target="#key-ppk-content">복사</button>
                </div>
            </div>
        `;
        DOM.keyDisplayArea.classList.remove('hidden');
    },

    hideKeys: function() {
        DOM.keyDisplayArea.classList.add('hidden');
        DOM.keyDisplayArea.innerHTML = '';
    }
};