window.ServerManager = {
    servers: [],
    init: function() {
        const serversView = document.getElementById('servers-view');
        if (!serversView) return;
        serversView.addEventListener('click', (e) => {
            const button = e.target.closest('button[data-action]');
            if (!button) return;
            e.preventDefault();
            const action = button.dataset.action;
            const serverId = button.dataset.serverId;
            switch (action) {
                case 'server-create-form': this.showServerForm(); break;
                case 'server-form-cancel': this.hideServerForm(); break;
                case 'server-delete': this.handleDeleteServer(serverId); break;
                case 'server-test': this.handleTestConnection(serverId); break;
                case 'server-deploy-key': this.handleDeployKey([serverId]); break;
                case 'deploy-all-keys': this.handleDeployToAll(); break;
            }
        });
        const serverForm = document.getElementById('server-form');
        if (serverForm) serverForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleCreateServer(e.target);
        });
    },
    loadServers: async function() {
        Utils.setLoading(true, '서버 목록 로딩 중...');
        try {
            const servers = await AppUtils.apiFetch('/servers');
            this.servers = servers || [];
            this.displayServers();
        } catch (error) {
            if (DOM.serversList) DOM.serversList.innerHTML = `<div class="error-message">서버 목록을 불러오는 데 실패했습니다.</div>`;
        } finally { Utils.setLoading(false); }
    },
    displayServers: function() {
        if (!DOM.serversList) return;
        DOM.serversList.innerHTML = '';
        if (this.servers.length === 0) {
            DOM.serversList.innerHTML = `<div class="empty-state">등록된 서버가 없습니다.</div>`;
            return;
        }
        this.servers.forEach(server => DOM.serversList.appendChild(this.createServerCard(server)));
    },
    createServerCard: function(server) {
        const card = document.createElement('div');
        card.className = 'server-card';
        card.innerHTML = `
            <div class="server-card-header"><span class="server-name">${Utils.escapeHtml(server.name)}</span><span class="server-host">${Utils.escapeHtml(server.username)}@${Utils.escapeHtml(server.host)}:${server.port}</span></div>
            <p class="server-description">${Utils.escapeHtml(server.description) || '설명이 없습니다.'}</p>
            <div class="server-actions">
                <button class="btn-secondary btn-sm" data-action="server-test" data-server-id="${server.ID}">연결 테스트</button>
                <button class="btn-success btn-sm" data-action="server-deploy-key" data-server-id="${server.ID}">키 배포</button>
                <button class="btn-danger btn-sm" data-action="server-delete" data-server-id="${server.ID}">삭제</button>
            </div>`;
        return card;
    },
    showServerForm: () => document.getElementById('server-form-modal')?.classList.remove('hidden'),
    hideServerForm: () => {
        const form = document.getElementById('server-form-modal');
        if (form) { form.classList.add('hidden'); form.querySelector('form').reset(); }
    },
    handleCreateServer: async function(form) {
        const data = {
            name: form.name.value, host: form.host.value, port: parseInt(form.port.value) || 22,
            username: form.username.value, description: form.description.value
        };
        Utils.setLoading(true, '서버 등록 중...');
        try {
            await AppUtils.apiFetch('/servers', 'POST', data);
            Utils.showToast('서버가 성공적으로 등록되었습니다.', 'success');
            this.hideServerForm();
            await this.loadServers();
        } finally { Utils.setLoading(false); }
    },
    handleDeleteServer: async function(serverId) {
        if (!confirm('정말로 이 서버를 삭제하시겠습니까?')) return;
        Utils.setLoading(true, '서버 삭제 중...');
        try {
            await AppUtils.apiFetch(`/servers/${serverId}`, 'DELETE');
            Utils.showToast('서버가 삭제되었습니다.', 'success');
            await this.loadServers();
        } finally { Utils.setLoading(false); }
    },
    handleTestConnection: async function(serverId) {
        Utils.setLoading(true, '서버 연결 테스트 중...');
        try {
            const result = await AppUtils.apiFetch(`/servers/${serverId}/test`, 'POST');
            if (result.success) Utils.showToast(`${result.server_name}: 연결 성공.`, 'success');
            else throw new Error(result.message || '연결 실패.');
        } catch (error) {
            Utils.showToast(error.message, 'error');
        } finally { Utils.setLoading(false); }
    },
    handleDeployKey: async function(serverIds) {
        if (!serverIds || serverIds.length === 0) return Utils.showToast('배포할 서버를 선택하세요.', 'warning');
        Utils.setLoading(true, 'SSH 키 배포 중...');
        try {
            const data = await AppUtils.apiFetch('/servers/deploy', 'POST', { server_ids: serverIds.map(id => parseInt(id)) });
            this.showDeploymentResults(data);
        } catch (error) {
            if (error.message.includes('SSH 키를 찾을 수 없습니다')) {
                 Utils.showToast('먼저 SSH 키를 생성해야 합니다.', 'error');
                 ViewManager.showView('keys');
            }
        } finally { Utils.setLoading(false); }
    },
    handleDeployToAll: function() {
        if (this.servers.length === 0) return Utils.showToast('등록된 서버가 없습니다.', 'info');
        if (!confirm(`모든 서버(${this.servers.length}개)에 키를 배포하시겠습니까?`)) return;
        this.handleDeployKey(this.servers.map(s => s.ID));
    },
    showDeploymentResults: function(data) {
        const results = data.results || [];
        let html = '<ul>';
        results.forEach(r => {
            const statusClass = r.status === 'success' ? 'deploy-result-success' : 'deploy-result-error';
            html += `<li class="${statusClass}"><strong>${Utils.escapeHtml(r.server_name)}:</strong> ${r.status === 'success' ? '성공' : `실패 - ${Utils.escapeHtml(r.error_message)}`}</li>`;
        });
        html += '</ul>';
        ModalManager.updateModalContent(`키 배포 결과 (${data.summary.success}/${data.summary.total})`, html);
        ModalManager.openModal('deploy-modal');
    }
};