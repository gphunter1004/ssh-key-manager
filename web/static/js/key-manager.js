// SSH 키 관리자 (복사 버그 수정 최종본)
window.KeyManager = {
    init: function() {
        document.addEventListener('click', (e) => {
            const action = e.target.dataset.action;
            if (!action || !action.startsWith('key-')) return;
            e.preventDefault();
            if (action === 'key-create') this.createKey();
            else if (action === 'key-view') this.viewKey();
            else if (action === 'key-delete') this.deleteKey();
        });
    },
    autoLoadKeys: async function() { await this.viewKey(); },
    createKey: async function() {
        if (!confirm('새로운 SSH 키를 생성하시겠습니까?\n기존 키가 있다면 덮어쓰여집니다.')) return;
        Utils.setLoading(true, '키 생성 중...');
        try {
            const keyData = await AppUtils.apiFetch('/keys', 'POST');
            this.displayKeys(keyData);
            Utils.showToast('SSH 키가 성공적으로 생성되었습니다!', 'success');
        } finally { Utils.setLoading(false); }
    },
    viewKey: async function() {
        Utils.setLoading(true, '키 정보 조회 중...');
        try {
            const keyData = await AppUtils.apiFetch('/keys');
            this.displayKeys(keyData);
        } catch (error) {
            if (DOM.keyInfo) {
                DOM.keyInfo.textContent = error.status === 404
                    ? '생성된 SSH 키가 없습니다. 키를 생성해주세요.'
                    : '키 정보를 불러오는 중 오류가 발생했습니다.';
            }
            this.hideKeys();
        } finally { Utils.setLoading(false); }
    },
    deleteKey: async function() {
        if (!confirm('정말로 SSH 키를 삭제하시겠습니까?\n이 작업은 되돌릴 수 없습니다.')) return;
        Utils.setLoading(true, '키 삭제 중...');
        try {
            await AppUtils.apiFetch('/keys', 'DELETE');
            this.hideKeys();
            if (DOM.keyInfo) DOM.keyInfo.textContent = 'SSH 키가 삭제되었습니다. 필요 시 다시 생성하세요.';
            Utils.showToast('SSH 키가 성공적으로 삭제되었습니다.', 'success');
        } finally { Utils.setLoading(false); }
    },
    displayKeys: function(keyData) {
        if (DOM.keyInfo) DOM.keyInfo.textContent = `키 알고리즘: ${keyData.Algorithm || 'RSA'} / ${keyData.Bits || 4096} bits`;
        if (DOM.keyDisplayArea) {
            DOM.keyDisplayArea.innerHTML = `
                <div class="key-item">
                    <h4>공개키 (Public Key)</h4>
                    <div class="command-display">
                        <pre id="cmd-public-content">echo '${keyData.PublicKey}' >> ~/.ssh/authorized_keys</pre>
                        <button class="copy-btn" data-target="#cmd-public-content">복사</button>
                    </div>
                    <div class="key-content">
                        <pre id="key-public-content">${keyData.PublicKey}</pre>
                        <button class="copy-btn" data-target="#key-public-content">복사</button>
                    </div>
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
        }
    },
    hideKeys: function() {
        if (DOM.keyDisplayArea) {
            DOM.keyDisplayArea.classList.add('hidden');
            DOM.keyDisplayArea.innerHTML = '';
        }
    }
};