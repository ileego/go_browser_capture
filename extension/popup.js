const wsDot = document.getElementById('wsDot');
const wsStatus = document.getElementById('wsStatus');
const pluginIdEl = document.getElementById('pluginId');
const wsUrlInput = document.getElementById('wsUrlInput');
const urlInput = document.getElementById('urlInput');
const selectorInput = document.getElementById('selectorInput');
const timeoutInput = document.getElementById('timeoutInput');
const minDelayInput = document.getElementById('minDelayInput');
const maxDelayInput = document.getElementById('maxDelayInput');
const captureBtn = document.getElementById('captureBtn');
const reconnectBtn = document.getElementById('reconnectBtn');
const testWsBtn = document.getElementById('testWsBtn');
const saveConfigBtn = document.getElementById('saveConfigBtn');
const messageEl = document.getElementById('message');
const resultEl = document.getElementById('result');
const resultContent = document.getElementById('resultContent');

function showMessage(text, type = 'success', duration = 3000) {
    messageEl.textContent = text;
    messageEl.className = `message ${type} show`;
    
    if (duration > 0) {
        setTimeout(() => {
            messageEl.classList.remove('show');
        }, duration);
    }
}

function updateWSStatus(connected) {
    if (connected) {
        wsDot.className = 'status-dot connected';
        wsStatus.textContent = '已连接';
        wsStatus.style.color = '#28a745';
    } else {
        wsDot.className = 'status-dot disconnected';
        wsStatus.textContent = '未连接';
        wsStatus.style.color = '#dc3545';
        showMessage('后端服务未启动，请先启动Go后端服务', 'info', 5000);
    }
}

async function getWSStatus() {
    try {
        const response = await chrome.runtime.sendMessage({ type: 'get_ws_status' });
        updateWSStatus(response.connected);
        if (response.pluginId) {
            pluginIdEl.textContent = response.pluginId;
            pluginIdEl.title = response.pluginId;
        }
    } catch (error) {
        console.error('Failed to get WS status:', error);
        updateWSStatus(false);
        pluginIdEl.textContent = '-';
    }
}

async function loadConfig() {
    try {
        const result = await chrome.storage.local.get('settings');
        if (result.settings) {
            if (result.settings.wsUrl) {
                wsUrlInput.value = result.settings.wsUrl;
            }
            if (result.settings.minDelay !== undefined) {
                minDelayInput.value = result.settings.minDelay;
            }
            if (result.settings.maxDelay !== undefined) {
                maxDelayInput.value = result.settings.maxDelay;
            }
        }
    } catch (error) {
        console.error('Failed to load config:', error);
    }
}

async function testWsConnection() {
    const wsUrl = wsUrlInput.value.trim();
    
    if (!wsUrl) {
        showMessage('请输入WebSocket地址', 'error');
        return;
    }

    if (!wsUrl.startsWith('ws://') && !wsUrl.startsWith('wss://')) {
        showMessage('WebSocket地址必须以 ws:// 或 wss:// 开头', 'error');
        return;
    }

    testWsBtn.disabled = true;
    testWsBtn.textContent = '测试中...';
    showMessage('正在测试连接...', 'info', 0);

    try {
        const result = await new Promise((resolve, reject) => {
            const ws = new WebSocket(wsUrl);
            const timeout = setTimeout(() => {
                ws.close();
                reject(new Error('连接超时'));
            }, 3000);

            ws.onopen = () => {
                clearTimeout(timeout);
                ws.close();
                resolve(true);
            };

            ws.onerror = () => {
                clearTimeout(timeout);
                reject(new Error('连接失败'));
            };
        });

        showMessage('连接测试成功！', 'success');
    } catch (error) {
        showMessage('连接测试失败: ' + error.message, 'error');
    } finally {
        testWsBtn.disabled = false;
        testWsBtn.textContent = '测试连接';
    }
}

async function saveConfig() {
    const wsUrl = wsUrlInput.value.trim();
    const minDelay = parseInt(minDelayInput.value) || 0;
    const maxDelay = parseInt(maxDelayInput.value) || 2000;
    
    if (!wsUrl) {
        showMessage('请输入WebSocket地址', 'error');
        return;
    }

    if (!wsUrl.startsWith('ws://') && !wsUrl.startsWith('wss://')) {
        showMessage('WebSocket地址必须以 ws:// 或 wss:// 开头', 'error');
        return;
    }

    if (minDelay < 0 || maxDelay < 0) {
        showMessage('延迟时间不能为负数', 'error');
        return;
    }

    if (minDelay > maxDelay) {
        showMessage('最小延迟不能大于最大延迟', 'error');
        return;
    }

    saveConfigBtn.disabled = true;
    saveConfigBtn.textContent = '保存中...';

    try {
        const result = await chrome.storage.local.get('settings');
        const settings = result.settings || {};
        settings.wsUrl = wsUrl;
        settings.minDelay = minDelay;
        settings.maxDelay = maxDelay;
        
        await chrome.storage.local.set({ settings });
        
        await chrome.runtime.sendMessage({ 
            type: 'update_ws_url', 
            wsUrl: wsUrl 
        });
        
        await chrome.runtime.sendMessage({ 
            type: 'update_delay_settings', 
            minDelay: minDelay, 
            maxDelay: maxDelay 
        });
        
        showMessage('配置已保存', 'success');
        
        setTimeout(() => {
            getWSStatus();
        }, 1000);
    } catch (error) {
        showMessage('保存失败: ' + error.message, 'error');
    } finally {
        saveConfigBtn.disabled = false;
        saveConfigBtn.textContent = '保存配置';
    }
}

async function handleCapture() {
    const url = urlInput.value.trim();
    const selector = selectorInput.value.trim();
    const timeout = parseInt(timeoutInput.value) || 30000;

    if (!url) {
        showMessage('请输入目标URL', 'error');
        return;
    }

    if (!selector) {
        showMessage('请输入DOM选择器', 'error');
        return;
    }

    captureBtn.disabled = true;
    showMessage('正在捕捉中...', 'info');
    resultEl.classList.remove('show');

    try {
        const requestId = 'manual_' + Date.now();
        
        const response = await chrome.runtime.sendMessage({
            type: 'manual_capture',
            data: {
                id: requestId,
                url: url,
                selector: selector,
                timeout: timeout,
                action: 'capture'
            }
        });

        if (response && response.success) {
            if (response.result) {
                if (response.result.success) {
                    showMessage('捕捉成功！', 'success');
                    displayResult(response.result);
                } else {
                    showMessage('捕捉失败: ' + response.result.error, 'error');
                    displayResult(response.result);
                }
            } else {
                showMessage('捕捉任务已提交', 'success');
            }
        } else {
            showMessage('捕捉任务提交失败', 'error');
        }
    } catch (error) {
        showMessage('捕捉失败: ' + error.message, 'error');
    } finally {
        captureBtn.disabled = false;
    }
}

function displayResult(result) {
    resultEl.classList.add('show');
    
    let content = '';
    if (result.success && result.data) {
        const data = result.data;
        content = JSON.stringify({
            html: data.html || '',
            text: data.text || ''
        }, null, 2);
    } else {
        content = JSON.stringify({
            success: result.success,
            error: result.error || 'Unknown error'
        }, null, 2);
    }
    
    resultContent.textContent = content;
}

async function handleReconnect() {
    try {
        await chrome.runtime.sendMessage({ type: 'reconnect_ws' });
        showMessage('正在重新连接...', 'info');
        
        setTimeout(() => {
            getWSStatus();
        }, 1000);
    } catch (error) {
        showMessage('重新连接失败: ' + error.message, 'error');
    }
}

chrome.runtime.onMessage.addListener((request) => {
    if (request.type === 'ws_status') {
        updateWSStatus(request.status === 'connected');
    }
});

captureBtn.addEventListener('click', handleCapture);
reconnectBtn.addEventListener('click', handleReconnect);
testWsBtn.addEventListener('click', testWsConnection);
saveConfigBtn.addEventListener('click', saveConfig);

document.addEventListener('DOMContentLoaded', () => {
    loadConfig();
    getWSStatus();
});