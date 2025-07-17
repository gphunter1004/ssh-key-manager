document.addEventListener('DOMContentLoaded', () => {
    // DOM 요소 가져오기 (기존과 동일)
    const container = document.querySelector('.container');
    const authSection = document.getElementById('auth-section');
    const keySection = document.getElementById('key-section');
    const loginView = document.getElementById('login-view');
    const registerView = document.getElementById('register-view');
    const errorDisplay = document.getElementById('error-display');
    const keyDisplayArea = document.getElementById('key-display-area');
    const keyInfo = document.getElementById('key-info');
    const keyPublicPre = document.getElementById('key-public');
    const keyPemPre = document.getElementById('key-pem');
    const keyPpkPre = document.getElementById('key-ppk');
    const cmdPublicPre = document.getElementById('cmd-public');
    const cmdAuthorizedKeysPre = document.getElementById('cmd-authorized-keys');
    const cmdPemPre = document.getElementById('cmd-pem');
    const cmdPpkPre = document.getElementById('cmd-ppk');
    const showRegisterLink = document.getElementById('show-register');
    const showLoginLink = document.getElementById('show-login');

    const API_BASE_URL = '/api';
    let jwtToken = localStorage.getItem('jwtToken') || null;

    function showError(message) { errorDisplay.textContent = message; errorDisplay.style.display = 'block'; }
    function clearError() { errorDisplay.textContent = ''; errorDisplay.style.display = 'none'; }

    async function apiFetch(endpoint, method = 'GET', body = null) {
        clearError();
        const headers = { 'Content-Type': 'application/json' };
        if (jwtToken) { headers['Authorization'] = `Bearer ${jwtToken}`; }
        const options = { method, headers };
        if (body) { options.body = JSON.stringify(body); }
        try {
            const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
            const data = await response.json();
            if (!response.ok) { throw new Error(data.error || data.message || 'An unknown error occurred.'); }
            return data;
        } catch (error) { showError(error.message); throw error; }
    }

    function updateUI() {
        if (jwtToken) {
            authSection.classList.add('hidden');
            keySection.classList.remove('hidden');
            container.classList.add('container-wide');
        } else {
            authSection.classList.remove('hidden');
            keySection.classList.add('hidden');
            container.classList.remove('container-wide');
            registerView.classList.add('hidden');
            loginView.classList.remove('hidden');
        }
    }

    showRegisterLink.addEventListener('click', (e) => { e.preventDefault(); clearError(); loginView.classList.add('hidden'); registerView.classList.remove('hidden'); });
    showLoginLink.addEventListener('click', (e) => { e.preventDefault(); clearError(); registerView.classList.add('hidden'); loginView.classList.remove('hidden'); });

    document.getElementById('login-form').addEventListener('submit', async (e) => { e.preventDefault(); try { const data = await apiFetch('/login', 'POST', { username: e.target.elements['login-username'].value, password: e.target.elements['login-password'].value }); jwtToken = data.token; localStorage.setItem('jwtToken', jwtToken); updateUI(); hideKeys(); } catch (error) {} });
    document.getElementById('register-form').addEventListener('submit', async (e) => { e.preventDefault(); try { await apiFetch('/register', 'POST', { username: e.target.elements['register-username'].value, password: e.target.elements['register-password'].value }); alert('User registered successfully! Please log in.'); e.target.reset(); registerView.classList.add('hidden'); loginView.classList.remove('hidden'); } catch (error) {} });
    document.getElementById('logout-btn').addEventListener('click', () => { jwtToken = null; localStorage.removeItem('jwtToken'); updateUI(); clearError(); hideKeys(); });
    
    function displayKeys(keyData) {
        keyInfo.textContent = `Algorithm: ${keyData.Algorithm} / Bits: ${keyData.Bits}`;
        keyPublicPre.textContent = keyData.PublicKey;
        keyPemPre.textContent = keyData.PEM;
        keyPpkPre.textContent = keyData.PPK;
        cmdPublicPre.textContent = `echo '${keyData.PublicKey}' > id_rsa.pub`;
        cmdAuthorizedKeysPre.textContent = `echo '${keyData.PublicKey}' >> ~/.ssh/authorized_keys`;
        cmdPemPre.textContent = `echo '${keyData.PEM}' > id_rsa`;
        cmdPpkPre.textContent = `echo '${keyData.PPK}' > id_rsa.ppk`;
        keyDisplayArea.classList.remove('hidden');
    }

    function hideKeys() { keyDisplayArea.classList.add('hidden'); keyInfo.textContent = ''; }
    
    document.getElementById('create-btn').addEventListener('click', async () => { try { const data = await apiFetch('/keys', 'POST'); displayKeys(data); } catch (error) {} });
    document.getElementById('view-btn').addEventListener('click', async () => { try { const data = await apiFetch('/keys', 'GET'); displayKeys(data); } catch (error) { hideKeys(); } });
    document.getElementById('delete-btn').addEventListener('click', async () => { if (!confirm('Are you sure you want to delete your key? This cannot be undone.')) return; try { const data = await apiFetch('/keys', 'DELETE'); alert(data.message); hideKeys(); } catch (error) {} });

    // ✅ HTTP 환경을 위한 복사 fallback 함수
    function fallbackCopyTextToClipboard(text) {
        const textArea = document.createElement("textarea");
        textArea.value = text;
        
        // 화면에 보이지 않도록 스타일 설정
        textArea.style.position = "fixed";
        textArea.style.top = 0;
        textArea.style.left = 0;
        textArea.style.width = "2em";
        textArea.style.height = "2em";
        textArea.style.padding = 0;
        textArea.style.border = "none";
        textArea.style.outline = "none";
        textArea.style.boxShadow = "none";
        textArea.style.background = "transparent";

        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        try {
            const successful = document.execCommand('copy');
            if (successful) {
                alert('Copied to clipboard!');
            } else {
                throw new Error('Copy command was not successful');
            }
        } catch (err) {
            showError('Oops, unable to copy');
        }

        document.body.removeChild(textArea);
    }

    // ✅ 복사 이벤트 리스너 수정
    keyDisplayArea.addEventListener('click', (e) => {
        const target = e.target;
        let textToCopy = '';

        if (target.classList.contains('copy-key-btn') || target.classList.contains('copy-cmd-btn')) {
            const targetId = target.dataset.targetId;
            textToCopy = document.getElementById(targetId).textContent;
        }

        if (textToCopy) {
            // 최신 API가 사용 불가능하면 fallback 함수 호출
            if (!navigator.clipboard) {
                fallbackCopyTextToClipboard(textToCopy);
                return;
            }
            // 최신 API 사용
            navigator.clipboard.writeText(textToCopy)
                .then(() => alert('Copied to clipboard!'))
                .catch(err => showError('Failed to copy.'));
        }
    });

    updateUI();
});