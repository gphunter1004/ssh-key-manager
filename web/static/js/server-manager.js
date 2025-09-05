// 서버 관리자 (서버 수정 기능 추가 최종본)
window.ServerManager = {
    servers: [],

    init: function() {
        console.log('🖥️ ServerManager 초기화 시작');
        this.setupEventListeners();
        return true;
    },

    setupEventListeners: function() {
        const serversView = document.getElementById('servers-view');
        if (!serversView) return;

        // 서버 뷰 내에서 발생하는 클릭 이벤트를 위임하여 처리
        serversView.addEventListener('click', (e) => {
            const button = e.target.closest('button[data-action]');
            if (!button) return;

            e.preventDefault();
            const action = button.dataset.action;
            const serverId = button.dataset.serverId;

            switch (action) {
                case 'server-create-form':
                    this.showServerForm();
                    break;
                case 'server-form-cancel':
                    this.hideServerForm();
                    break;
                case 'server-edit-form':
                    this.showServerEditForm(serverId);
                    break;
                case 'server-delete':
                    this.handleDeleteServer(serverId);
                    break;
                case 'server-test':
                    this.handleTestConnection(serverId);
                    break;
                case 'server-deploy-key':
                    this.handleDeployKey([serverId]);
                    break;
                case 'deploy-all-keys':
                    this.handleDeployToAll();
                    break;
            }
        });

        // 서버 생성 폼 제출 이벤트
        const serverForm = document.getElementById('server-form');
        if (serverForm) {
            serverForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.handleCreateServer(e.target);
            });
        }

        // 수정 폼 제출 이벤트 리스너 추가 (이벤트 위임)
        document.body.addEventListener('submit', (e) => {
            if (e.target.id === 'server-edit-form') {
                e.preventDefault();
                const serverId = e.target.dataset.serverId;
                this.handleUpdateServer(serverId, e.target);
            }
        });
    },

    loadServers: async function() {
        Utils.setLoading(true, '서버 목록을 불러오는 중...');
        try {
            const servers = await AppUtils.apiFetch('/servers');
            this.servers = servers || [];
            this.displayServers();
        } catch (error) {
            if (DOM.serversList) {
                DOM.serversList.innerHTML = `<div class="error-message">서버 목록을 불러오는 데 실패했습니다.</div>`;
            }
        } finally {
            Utils.setLoading(false);
        }
    },

    displayServers: function() {
        if (!DOM.serversList) return;
        DOM.serversList.innerHTML = '';
        if (this.servers.length === 0) {
            DOM.serversList.innerHTML = `<div class="empty-state">등록된 서버가 없습니다. '서버 추가' 버튼을 눌러 시작하세요.</div>`;
            return;
        }
        this.servers.forEach(server => {
            const serverCard = this.createServerCard(server);
            DOM.serversList.appendChild(serverCard);
        });
    },

    createServerCard: function(server) {
        const card = document.createElement('div');
        card.className = 'server-card';
        card.innerHTML = `
            <div class="server-card-header">
                <span class="server-name">${Utils.escapeHtml(server.name)}</span>
                <span class="server-host">${Utils.escapeHtml(server.username)}@${Utils.escapeHtml(server.host)}:${server.port}</span>
            </div>
            <p class="server-description">
                <strong>경로:</strong> <code>${Utils.escapeHtml(server.target_directory)}</code><br>
                ${Utils.escapeHtml(server.description) || '설명이 없습니다.'}
            </p>
            <div class="server-actions">
                <button class="btn-secondary btn-sm" data-action="server-edit-form" data-server-id="${server.ID}">수정</button>
                <button class="btn-secondary btn-sm" data-action="server-test" data-server-id="${server.ID}">연결 테스트</button>
                <button class="btn-success btn-sm" data-action="server-deploy-key" data-server-id="${server.ID}">키 배포</button>
                <button class="btn-danger btn-sm" data-action="server-delete" data-server-id="${server.ID}">삭제</button>
            </div>
        `;
        return card;
    },

    showServerForm: function() {
        const form = document.getElementById('server-form-modal');
        if (form) form.classList.remove('hidden');
    },

    hideServerForm: function() {
        const form = document.getElementById('server-form-modal');
        if (form) {
            form.classList.add('hidden');
            form.querySelector('form').reset();
        }
    },

    handleCreateServer: async function(formElement) {
        const formData = new FormData(formElement);
        const serverData = {
            name: formData.get('name'),
            host: formData.get('host'),
            port: parseInt(formData.get('port'), 10) || 22,
            username: formData.get('username'),
            description: formData.get('description'),
            target_directory: formData.get('target_directory')
        };

        Utils.setLoading(true, '서버를 등록하는 중...');
        try {
            await AppUtils.apiFetch('/servers', 'POST', serverData);
            Utils.showToast('서버가 성공적으로 등록되었습니다.', 'success');
            this.hideServerForm();
            await this.loadServers();
        } catch (error) {
            // 에러 메시지는 apiFetch에서 자동으로 처리합니다.
        } finally {
            Utils.setLoading(false);
        }
    },

    handleDeleteServer: async function(serverId) {
        if (!confirm('정말로 이 서버를 삭제하시겠습니까?')) return;
        
        Utils.setLoading(true, '서버를 삭제하는 중...');
        try {
            await AppUtils.apiFetch(`/servers/${serverId}`, 'DELETE');
            Utils.showToast('서버가 삭제되었습니다.', 'success');
            await this.loadServers();
        } catch (error) {
             // 에러 메시지는 apiFetch에서 자동으로 처리합니다.
        } finally {
            Utils.setLoading(false);
        }
    },

    handleTestConnection: async function(serverId) {
        Utils.setLoading(true, '서버 연결을 테스트하는 중...');
        try {
            const result = await AppUtils.apiFetch(`/servers/${serverId}/test`, 'POST');
            if (result.success) {
                Utils.showToast(`${result.server_name}: 연결에 성공했습니다.`, 'success');
            } else {
                throw new Error(result.message || '연결에 실패했습니다.');
            }
        } catch (error) {
            Utils.showToast(error.message, 'error');
        } finally {
            Utils.setLoading(false);
        }
    },

    handleDeployKey: async function(serverIds) {
        if (!serverIds || serverIds.length === 0) {
            return Utils.showToast('배포할 서버를 선택하세요.', 'warning');
        }
        
        Utils.setLoading(true, 'SSH 키를 배포하는 중...');
        try {
            const data = await AppUtils.apiFetch('/servers/deploy', 'POST', { server_ids: serverIds.map(id => parseInt(id)) });
            this.showDeploymentResults(data);
        } catch (error) {
            if (error.message.includes('SSH 키를 찾을 수 없습니다')) {
                 Utils.showToast('먼저 SSH 키를 생성해야 합니다.', 'error');
                 ViewManager.showView('keys');
            }
        } finally {
            Utils.setLoading(false);
        }
    },

    handleDeployToAll: function() {
        if (this.servers.length === 0) {
            return Utils.showToast('등록된 서버가 없습니다.', 'info');
        }
        if (!confirm(`모든 서버(${this.servers.length}개)에 키를 배포하시겠습니까?`)) return;
        
        const allServerIds = this.servers.map(s => s.ID);
        this.handleDeployKey(allServerIds);
    },

    showDeploymentResults: function(data) {
        const results = data.results || [];
        let resultsHTML = '<ul>';
        results.forEach(result => {
            const statusClass = result.status === 'success' ? 'deploy-result-success' : 'deploy-result-error';
            resultsHTML += `<li class="${statusClass}"><strong>${Utils.escapeHtml(result.server_name)}:</strong> ${result.status === 'success' ? '성공' : `실패 - ${Utils.escapeHtml(result.error_message)}`}</li>`;
        });
        resultsHTML += '</ul>';

        ModalManager.updateModalContent(`키 배포 결과 (${data.summary.success}/${data.summary.total})`, resultsHTML);
        ModalManager.openModal('deploy-modal');
    },

    showServerEditForm: async function(serverId) {
        ModalManager.openModalWithSpinner('server-edit-modal', '서버 정보 불러오는 중...');
        try {
            const server = this.servers.find(s => s.ID == serverId);
            if (!server) throw new Error('서버 정보를 찾을 수 없습니다.');

            const formHTML = `
                <form id="server-edit-form" data-server-id="${server.ID}" class="form-container" style="border: none; padding: 0;">
                    <div class="form-row">
                        <div class="form-group"><label>서버명</label><input type="text" name="name" value="${Utils.escapeHtml(server.name)}" required></div>
                        <div class="form-group"><label>호스트</label><input type="text" name="host" value="${Utils.escapeHtml(server.host)}" required></div>
                    </div>
                    <div class="form-row">
                        <div class="form-group"><label>포트</label><input type="number" name="port" value="${server.port}"></div>
                        <div class="form-group"><label>사용자명</label><input type="text" name="username" value="${Utils.escapeHtml(server.username)}" required></div>
                    </div>
                    <div class="form-group"><label>설명</label><textarea name="description">${Utils.escapeHtml(server.description)}</textarea></div>
                    <div class="form-group"><label>대상 디렉토리</label><input type="text" name="target_directory" value="${Utils.escapeHtml(server.target_directory)}"></div>
                    <div class="form-actions">
                        <button type="submit" class="btn-primary">저장</button>
                        <button type="button" class="btn-secondary" data-action="close-modal">취소</button>
                    </div>
                </form>
            `;
            ModalManager.updateModalContent('서버 정보 수정', formHTML);
        } catch (error) {
            ModalManager.updateModalContent('오류', `<div class="error-message">${error.message}</div>`);
        }
    },

    handleUpdateServer: async function(serverId, formElement) {
        const formData = new FormData(formElement);
        const serverData = {
            name: formData.get('name'),
            host: formData.get('host'),
            port: parseInt(formData.get('port'), 10) || 22,
            username: formData.get('username'),
            description: formData.get('description'),
            target_directory: formData.get('target_directory')
        };

        Utils.setLoading(true, '서버 정보를 업데이트하는 중...');
        try {
            await AppUtils.apiFetch(`/servers/${serverId}`, 'PUT', serverData);
            Utils.showToast('서버 정보가 성공적으로 업데이트되었습니다.', 'success');
            ModalManager.closeActiveModal();
            await this.loadServers();
        } catch (error) {
            // 에러 메시지는 apiFetch에서 자동으로 표시됩니다.
        } finally {
            Utils.setLoading(false);
        }
    }
};