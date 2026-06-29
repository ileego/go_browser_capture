const API_BASE = '/api/v1';

let token = localStorage.getItem('domcapture_token');
let currentUser = JSON.parse(localStorage.getItem('domcapture_user') || 'null');

function showError(msg) {
    document.getElementById('errorMsg').textContent = msg;
    document.getElementById('errorMsg').style.display = 'block';
}

function hideError() {
    document.getElementById('errorMsg').style.display = 'none';
}

function showToast(msg, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = msg;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}

async function apiRequest(url, method = 'GET', data = null) {
    const options = {
        method,
        headers: { 'Content-Type': 'application/json' },
    };
    
    if (token) {
        options.headers['Authorization'] = `Bearer ${token}`;
    }
    
    if (data) {
        options.body = JSON.stringify(data);
    }
    
    try {
        const response = await fetch(API_BASE + url, options);
        const result = await response.json();
        
        if (!result.success && response.status === 401 && url !== '/auth/login') {
            logout();
            return null;
        }
        
        return result;
    } catch (error) {
        console.error('API请求失败:', error);
        showError('服务不可用，请检查后端服务是否正常运行');
        return null;
    }
}

function logout() {
    localStorage.removeItem('domcapture_token');
    localStorage.removeItem('domcapture_user');
    token = null;
    currentUser = null;
    window.location.href = '/admin/login.html';
}

document.addEventListener('DOMContentLoaded', () => {
    if (window.location.pathname.includes('login.html')) {
        initLogin();
    } else {
        initAdmin();
    }
});

function initLogin() {
    document.getElementById('loginForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        hideError();
        
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        
        const result = await apiRequest('/auth/login', 'POST', { username, password });
        
        if (!result || !result.success) {
            showError(result ? result.message : '登录失败');
            return;
        }
        
        token = result.token;
        localStorage.setItem('domcapture_token', token);
        localStorage.setItem('domcapture_user', JSON.stringify(result));
        
        window.location.href = '/admin/index.html';
    });
}

function initAdmin() {
    if (!token) {
        window.location.href = '/admin/login.html';
        return;
    }
    
    document.getElementById('currentUser').textContent = currentUser?.username || '管理员';
    
    document.getElementById('logoutBtn').addEventListener('click', logout);
    
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            switchPage(link.dataset.page);
        });
    });
    
    loadData();
}

function switchPage(page) {
    document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
    document.querySelector(`[data-page="${page}"]`).classList.add('active');
    
    document.querySelectorAll('.page-section').forEach(s => s.classList.add('hidden'));
    document.getElementById(page).classList.remove('hidden');
    
    const titles = {
        dashboard: '仪表盘',
        users: '用户管理',
        roles: '权限管理',
        configs: '配置管理',
        plugins: '插件列表',
        tasks: '任务列表',
    };
    document.getElementById('pageTitle').textContent = titles[page];
    
    if (page === 'dashboard') loadDashboard();
    else if (page === 'users') loadUsers();
    else if (page === 'roles') loadRoles();
    else if (page === 'configs') loadConfigs();
    else if (page === 'plugins') loadPlugins();
    else if (page === 'tasks') loadTasks();
}

async function loadData() {
    await loadDashboard();
    await loadUsers();
    await loadRoles();
    await loadConfigs();
}

async function loadDashboard() {
    const result = await apiRequest('/status/system');
    if (!result || !result.success) return;
    
    const { plugins, tasks } = result.data;
    
    document.getElementById('statPlugins').textContent = plugins.active || 0;
    document.getElementById('statTasks').textContent = tasks.total || 0;
    document.getElementById('statPending').textContent = tasks.pending || 0;
    document.getElementById('statFailed').textContent = tasks.failed || 0;
}

async function loadUsers() {
    const result = await apiRequest('/users');
    if (!result || !result.success) return;
    
    const tbody = document.getElementById('usersTableBody');
    tbody.innerHTML = '';
    
    result.data.forEach(user => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${user.id.slice(0, 8)}...</td>
            <td>${user.username}</td>
            <td>${user.email || '-'}</td>
            <td>${user.role?.name || '-'}</td>
            <td><span class="status-badge ${user.is_active ? 'status-active' : 'status-inactive'}">${user.is_active ? '活跃' : '禁用'}</span></td>
            <td>${formatDate(user.created_at)}</td>
            <td>
                <div class="action-buttons">
                    <button class="btn btn-outline" onclick="showModal('editUser', '${user.id}')">编辑</button>
                    <button class="btn btn-danger" onclick="deleteUser('${user.id}')">删除</button>
                </div>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

async function loadRoles() {
    const result = await apiRequest('/roles');
    if (!result || !result.success) return;
    
    const tbody = document.getElementById('rolesTableBody');
    tbody.innerHTML = '';
    
    result.data.forEach(role => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${role.id}</td>
            <td>${role.name}</td>
            <td>${role.description}</td>
            <td>${role.permissions?.join(', ') || '-'}</td>
            <td>${formatDate(role.created_at)}</td>
        `;
        tbody.appendChild(tr);
    });
}

async function loadConfigs() {
    const result = await apiRequest('/selector-configs');
    if (!result || !result.success) return;
    
    const tbody = document.getElementById('configsTableBody');
    tbody.innerHTML = '';
    
    result.data.forEach(config => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${config.name}</td>
            <td>${config.domain}</td>
            <td title="${config.selector}">${config.selector.length > 30 ? config.selector.slice(0, 30) + '...' : config.selector}</td>
            <td><span class="status-badge ${config.is_active ? 'status-active' : 'status-inactive'}">${config.is_active ? '启用' : '禁用'}</span></td>
            <td>${formatDate(config.created_at)}</td>
            <td>
                <div class="action-buttons">
                    <button class="btn btn-outline" onclick="showModal('editConfig', '${config.id}')">编辑</button>
                    <button class="btn btn-danger" onclick="deleteConfig('${config.id}')">删除</button>
                </div>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

async function loadPlugins() {
    const result = await apiRequest('/status/plugins');
    if (!result || !result.success) return;
    
    const tbody = document.getElementById('pluginsTableBody');
    tbody.innerHTML = '';
    
    result.data.forEach(plugin => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${plugin.id.slice(0, 8)}...</td>
            <td><span class="status-badge ${plugin.status === 'IDLE' ? 'status-idle' : 'status-busy'}">${plugin.status === 'IDLE' ? '空闲' : '忙碌'}</span></td>
            <td>${plugin.taskCount}</td>
            <td>${plugin.totalTasksProcessed}</td>
            <td>${plugin.avgResponseTime}</td>
            <td>${formatDate(plugin.lastHeartbeat)}</td>
            <td>${formatDate(plugin.connectedAt)}</td>
        `;
        tbody.appendChild(tr);
    });
}

async function loadTasks() {
    const statusFilter = document.getElementById('taskStatusFilter')?.value || '';
    const url = statusFilter ? `/status/tasks?status=${encodeURIComponent(statusFilter)}` : '/status/tasks';
    const result = await apiRequest(url);
    if (!result || !result.success) return;
    
    const tbody = document.getElementById('tasksTableBody');
    tbody.innerHTML = '';
    
    const statsDiv = document.getElementById('tasksStats');
    statsDiv.innerHTML = `
        <div style="display: flex; gap: 16px; margin-bottom: 16px;">
            <span>待处理: ${result.stats.pending || 0}</span>
            <span>处理中: ${result.stats.assigned || 0}</span>
            <span>已完成: ${result.stats.completed || 0}</span>
            <span>失败: ${result.stats.failed || 0}</span>
        </div>
    `;
    
    if (result.data.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty-state">暂无任务</td></tr>';
        return;
    }
    
    result.data.forEach(task => {
        let statusClass = 'status-pending';
        let statusText = '待处理';
        if (task.status === 'COMPLETED') { statusClass = 'status-completed'; statusText = '已完成'; }
        else if (task.status === 'FAILED') { statusClass = 'status-failed'; statusText = '失败'; }
        else if (task.status === 'TIMEOUT') { statusClass = 'status-timeout'; statusText = '超时'; }
        else if (task.status === 'ASSIGNED') { statusClass = 'status-busy'; statusText = '处理中'; }
        else if (task.status === 'RUNNING') { statusClass = 'status-busy'; statusText = '运行中'; }
        
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${task.id.slice(0, 8)}...</td>
            <td title="${task.url}">${task.url.length > 40 ? task.url.slice(0, 40) + '...' : task.url}</td>
            <td title="${task.selector}">${task.selector.length > 30 ? task.selector.slice(0, 30) + '...' : task.selector}</td>
            <td><span class="status-badge ${statusClass}">${statusText}</span></td>
            <td>${task.pluginID ? task.pluginID.slice(0, 8) + '...' : '-'}</td>
            <td>${task.error || '-'}</td>
            <td>${formatDate(task.createdAt)}</td>
        `;
        tbody.appendChild(tr);
    });
}

function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
}

function showModal(type, id = null) {
    const modal = document.getElementById('modalOverlay');
    const title = document.getElementById('modalTitle');
    const body = document.getElementById('modalBody');
    const submit = document.getElementById('modalSubmit');
    
    modal.classList.remove('hidden');
    
    if (type === 'createUser') {
        title.textContent = '添加用户';
        body.innerHTML = `
            <div class="form-group">
                <label>用户名</label>
                <input type="text" id="modalUsername" required>
            </div>
            <div class="form-group">
                <label>密码</label>
                <input type="password" id="modalPassword" required>
            </div>
            <div class="form-group">
                <label>邮箱</label>
                <input type="email" id="modalEmail">
            </div>
            <div class="form-group">
                <label>角色</label>
                <select id="modalRole" class="form-control">
                    <option value="admin">管理员</option>
                    <option value="editor">编辑</option>
                    <option value="viewer">查看者</option>
                </select>
            </div>
        `;
        submit.onclick = async () => {
            const result = await apiRequest('/auth/register', 'POST', {
                username: document.getElementById('modalUsername').value,
                password: document.getElementById('modalPassword').value,
                email: document.getElementById('modalEmail').value,
                role_name: document.getElementById('modalRole').value,
            });
            if (result?.success) {
                showToast('用户添加成功');
                hideModal();
                loadUsers();
            } else {
                showToast(result?.message || '添加失败', 'error');
            }
        };
    } else if (type === 'editUser') {
        title.textContent = '编辑用户';
        body.innerHTML = `
            <input type="hidden" id="modalUserId" value="${id}">
            <div class="form-group">
                <label>用户名</label>
                <input type="text" id="modalUsername">
            </div>
            <div class="form-group">
                <label>邮箱</label>
                <input type="email" id="modalEmail">
            </div>
            <div class="form-group">
                <label>是否活跃</label>
                <select id="modalIsActive" class="form-control">
                    <option value="true">是</option>
                    <option value="false">否</option>
                </select>
            </div>
        `;
        apiRequest('/users/' + id).then(result => {
            if (result?.success) {
                document.getElementById('modalUsername').value = result.data.username;
                document.getElementById('modalEmail').value = result.data.email;
                document.getElementById('modalIsActive').value = String(result.data.is_active);
            }
        });
        submit.onclick = async () => {
            const result = await apiRequest('/users/' + id, 'PUT', {
                username: document.getElementById('modalUsername').value,
                email: document.getElementById('modalEmail').value,
                is_active: document.getElementById('modalIsActive').value === 'true',
            });
            if (result?.success) {
                showToast('用户更新成功');
                hideModal();
                loadUsers();
            } else {
                showToast(result?.message || '更新失败', 'error');
            }
        };
    } else if (type === 'createConfig') {
        title.textContent = '添加配置';
        body.innerHTML = `
            <div class="form-group">
                <label>名称</label>
                <input type="text" id="modalConfigName" required>
            </div>
            <div class="form-group">
                <label>域名</label>
                <input type="text" id="modalConfigDomain" required placeholder="如: item.jd.com">
            </div>
            <div class="form-group">
                <label>选择器</label>
                <input type="text" id="modalConfigSelector" required placeholder="如: .price">
            </div>
            <div class="form-group">
                <label>正则表达式</label>
                <input type="text" id="modalConfigRegex" placeholder="可选">
            </div>
        `;
        submit.onclick = async () => {
            const result = await apiRequest('/selector-configs', 'POST', {
                name: document.getElementById('modalConfigName').value,
                domain: document.getElementById('modalConfigDomain').value,
                selector: document.getElementById('modalConfigSelector').value,
                regex: document.getElementById('modalConfigRegex').value,
                is_active: true,
            });
            if (result?.success) {
                showToast('配置添加成功');
                hideModal();
                loadConfigs();
            } else {
                showToast(result?.message || '添加失败', 'error');
            }
        };
    } else if (type === 'editConfig') {
        title.textContent = '编辑配置';
        body.innerHTML = `
            <input type="hidden" id="modalConfigId" value="${id}">
            <div class="form-group">
                <label>名称</label>
                <input type="text" id="modalConfigName" required>
            </div>
            <div class="form-group">
                <label>域名</label>
                <input type="text" id="modalConfigDomain" required>
            </div>
            <div class="form-group">
                <label>选择器</label>
                <input type="text" id="modalConfigSelector" required>
            </div>
            <div class="form-group">
                <label>正则表达式</label>
                <input type="text" id="modalConfigRegex">
            </div>
            <div class="form-group">
                <label>是否启用</label>
                <select id="modalConfigActive" class="form-control">
                    <option value="true">是</option>
                    <option value="false">否</option>
                </select>
            </div>
        `;
        apiRequest('/selector-configs/' + id).then(result => {
            if (result?.success) {
                document.getElementById('modalConfigName').value = result.data.name;
                document.getElementById('modalConfigDomain').value = result.data.domain;
                document.getElementById('modalConfigSelector').value = result.data.selector;
                document.getElementById('modalConfigRegex').value = result.data.regex || '';
                document.getElementById('modalConfigActive').value = String(result.data.is_active);
            }
        });
        submit.onclick = async () => {
            const result = await apiRequest('/selector-configs/' + id, 'PUT', {
                name: document.getElementById('modalConfigName').value,
                domain: document.getElementById('modalConfigDomain').value,
                selector: document.getElementById('modalConfigSelector').value,
                regex: document.getElementById('modalConfigRegex').value,
                is_active: document.getElementById('modalConfigActive').value === 'true',
            });
            if (result?.success) {
                showToast('配置更新成功');
                hideModal();
                loadConfigs();
            } else {
                showToast(result?.message || '更新失败', 'error');
            }
        };
    }
}

function hideModal() {
    document.getElementById('modalOverlay').classList.add('hidden');
}

async function deleteUser(id) {
    if (!confirm('确定要删除该用户吗？')) return;
    const result = await apiRequest('/users/' + id, 'DELETE');
    if (result?.success) {
        showToast('用户删除成功');
        loadUsers();
    } else {
        showToast(result?.message || '删除失败', 'error');
    }
}

async function deleteConfig(id) {
    if (!confirm('确定要删除该配置吗？')) return;
    const result = await apiRequest('/selector-configs/' + id, 'DELETE');
    if (result?.success) {
        showToast('配置删除成功');
        loadConfigs();
    } else {
        showToast(result?.message || '删除失败', 'error');
    }
}