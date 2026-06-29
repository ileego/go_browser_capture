let wsUrl = 'ws://localhost:8080/ws';
let minDelay = 0;
let maxDelay = 2000;

let ws = null;
let connectionAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 10;
const RECONNECT_DELAY = 5000;
const HEARTBEAT_INTERVAL = 15000;

let heartbeatTimer = null;
let pluginId = generatePluginId();

const pendingRequests = new Map();
const activeTabs = new Map();
const requestQueue = [];
let lastRequestTime = 0;
let isProcessingQueue = false;

function generatePluginId() {
    return 'plugin_' + Math.random().toString(36).substring(2, 15) + Date.now().toString(36);
}

async function loadSettings() {
    try {
        const result = await chrome.storage.local.get('settings');
        if (result.settings) {
            if (result.settings.wsUrl) {
                wsUrl = result.settings.wsUrl;
                console.log('[Config] Loaded wsUrl from storage:', wsUrl);
            }
            if (result.settings.minDelay !== undefined) {
                minDelay = result.settings.minDelay;
                console.log('[Config] Loaded minDelay:', minDelay);
            }
            if (result.settings.maxDelay !== undefined) {
                maxDelay = result.settings.maxDelay;
                console.log('[Config] Loaded maxDelay:', maxDelay);
            }
        }
    } catch (error) {
        console.error('[Config] Failed to load settings:', error);
    }
}

function connectWebSocket() {
    if (ws) {
        ws.close();
    }

    console.log('[WebSocket] Connecting to:', wsUrl);
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        console.log('[WebSocket] Connected to server');
        connectionAttempts = 0;
        sendStatusMessage('connected', 'WebSocket connection established');
        
        sendRegistration();
        startHeartbeat();
    };

    ws.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data);
            console.log('[WebSocket] Received message:', message);

            switch (message.action) {
                case 'capture':
                    handleCaptureRequest(message);
                    break;
                case 'heartbeat':
                    handleHeartbeatRequest(message);
                    break;
                case 'registered':
                    handleRegistered(message);
                    break;
            }
        } catch (error) {
            console.error('[WebSocket] Error parsing message:', error);
        }
    };

    ws.onerror = (error) => {
        console.error('[WebSocket] Error:', error);
        sendStatusMessage('error', 'Connection error');
    };

    ws.onclose = () => {
        console.log('[WebSocket] Disconnected');
        sendStatusMessage('disconnected', 'WebSocket connection closed');
        stopHeartbeat();

        if (connectionAttempts < MAX_RECONNECT_ATTEMPTS) {
            connectionAttempts++;
            const delay = RECONNECT_DELAY * connectionAttempts;
            console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${connectionAttempts}/${MAX_RECONNECT_ATTEMPTS})`);
            setTimeout(connectWebSocket, delay);
        } else {
            console.error('[WebSocket] Max reconnection attempts reached');
        }
    };
}

function sendRegistration() {
    const message = {
        action: 'register',
        pluginId: pluginId,
        timestamp: new Date().toISOString()
    };
    sendWebSocketMessage(message);
}

function startHeartbeat() {
    stopHeartbeat();
    heartbeatTimer = setInterval(() => {
        const message = {
            action: 'heartbeat',
            pluginId: pluginId,
            timestamp: new Date().toISOString()
        };
        sendWebSocketMessage(message);
    }, HEARTBEAT_INTERVAL);
}

function stopHeartbeat() {
    if (heartbeatTimer) {
        clearInterval(heartbeatTimer);
        heartbeatTimer = null;
    }
}

function handleHeartbeatRequest(message) {
    const response = {
        action: 'heartbeat_response',
        pluginId: pluginId,
        requestId: message.id,
        timestamp: new Date().toISOString()
    };
    sendWebSocketMessage(response);
}

function handleRegistered(message) {
    console.log('[WebSocket] Registered successfully:', message);
}

function sendStatusMessage(status, message) {
    chrome.runtime.sendMessage({
        type: 'ws_status',
        status,
        message
    }).catch(() => {});
}

function sendWebSocketMessage(message) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(message));
        return true;
    }
    console.error('[WebSocket] Connection not open');
    return false;
}

async function handleCaptureRequest(request) {
    return new Promise((resolve) => {
        requestQueue.push({ request, resolve });
        processRequestQueue();
    });
}

async function processRequestQueue() {
    if (isProcessingQueue) {
        return;
    }
    isProcessingQueue = true;

    while (requestQueue.length > 0) {
        const now = Date.now();
        const timeSinceLastRequest = now - lastRequestTime;

        if (lastRequestTime > 0 && timeSinceLastRequest < 1000) {
            const delay = Math.floor(Math.random() * (maxDelay - minDelay + 1)) + minDelay;
            console.log(`[Queue] Adding random delay of ${delay}ms (requests within 1s)`);
            await new Promise(resolve => setTimeout(resolve, delay));
        }

        const { request, resolve } = requestQueue.shift();
        lastRequestTime = Date.now();

        const result = await executeCapture(request);
        resolve(result);
    }

    isProcessingQueue = false;
}

async function executeCapture(request) {
    const { id, url, selector, timeout = 30000 } = request;

    console.log(`[Capture] Starting capture request: ${id}`);
    console.log(`[Capture] URL: ${url}`);
    console.log(`[Capture] Selector: ${selector}`);

    try {
        const tab = await chrome.tabs.create({
            url: url,
            active: false
        });

        const startTime = Date.now();
        const maxWaitTime = timeout;
        const pollInterval = 500;
        let found = false;
        let domContent = null;

        const checkAndCapture = async () => {
            try {
                const result = await chrome.scripting.executeScript({
                    target: { tabId: tab.id },
                    func: captureDOM,
                    args: [selector]
                });

                if (result && result[0] && result[0].result) {
                    const data = result[0].result;
                    if (data.success && data.content) {
                        return data.content;
                    }
                }
            } catch (error) {
                console.warn('[Capture] Error checking DOM:', error);
            }
            return null;
        };

        while (Date.now() - startTime < maxWaitTime && !found) {
            domContent = await checkAndCapture();
            
            if (domContent) {
                found = true;
                break;
            }

            await new Promise(resolve => setTimeout(resolve, pollInterval));
        }

        await chrome.tabs.remove(tab.id);

        if (found && domContent) {
            const response = {
                action: 'response',
                id: id,
                pluginId: pluginId,
                success: true,
                data: {
                    id: id,
                    html: domContent.html,
                    text: domContent.text,
                    timestamp: new Date().toISOString()
                },
                error: null
            };
            sendWebSocketMessage(response);
            console.log(`[Capture] Successfully captured: ${id}`);
            console.log('[Capture] Result:', JSON.stringify(response, null, 2));
            return response;
        } else {
            const response = {
                action: 'response',
                id: id,
                pluginId: pluginId,
                success: false,
                data: null,
                error: found ? 'DOM element not found' : 'Timeout waiting for DOM element'
            };
            sendWebSocketMessage(response);
            console.log(`[Capture] Failed to capture: ${id}`);
            return response;
        }

    } catch (error) {
        console.error('[Capture] Error:', error);
        const response = {
            action: 'response',
            id: id,
            pluginId: pluginId,
            success: false,
            data: null,
            error: error.message
        };
        sendWebSocketMessage(response);
        return response;
    }
}

function captureDOM(selector) {
    try {
        const element = document.querySelector(selector);
        
        if (!element) {
            return {
                success: false,
                content: null,
                error: 'Element not found'
            };
        }

        return {
            success: true,
            content: {
                html: element.innerHTML,
                text: element.textContent
            },
            error: null
        };
    } catch (error) {
        return {
            success: false,
            content: null,
            error: error.message
        };
    }
}

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    switch (request.type) {
        case 'manual_capture':
            handleCaptureRequest(request.data).then((result) => {
                sendResponse({ success: true, result: result });
            }).catch(error => {
                sendResponse({ success: false, error: error.message });
            });
            return true;

        case 'get_ws_status':
            sendResponse({
                connected: ws && ws.readyState === WebSocket.OPEN,
                attempts: connectionAttempts,
                pluginId: pluginId
            });
            break;

        case 'reconnect_ws':
            connectWebSocket();
            sendResponse({ success: true });
            break;

        case 'update_ws_url':
            wsUrl = request.wsUrl;
            console.log('[Config] Updated wsUrl:', wsUrl);
            if (ws) {
                ws.close();
            }
            connectionAttempts = 0;
            setTimeout(connectWebSocket, 500);
            sendResponse({ success: true });
            break;

        case 'update_delay_settings':
            if (request.minDelay !== undefined) {
                minDelay = request.minDelay;
                console.log('[Config] Updated minDelay:', minDelay);
            }
            if (request.maxDelay !== undefined) {
                maxDelay = request.maxDelay;
                console.log('[Config] Updated maxDelay:', maxDelay);
            }
            sendResponse({ success: true });
            break;
    }
});

chrome.runtime.onInstalled.addListener((details) => {
    console.log('[Extension] Installed:', details.reason);
    
    if (details.reason === 'install') {
        chrome.storage.local.set({
            settings: {
                apiUrl: 'http://localhost:8080/api/v1',
                wsUrl: 'ws://localhost:8080/ws',
                autoReconnect: true,
                defaultTimeout: 30000,
                minDelay: 0,
                maxDelay: 2000
            }
        });
    }

    loadSettings().then(() => {
        connectWebSocket();
    });
});

chrome.runtime.onStartup.addListener(() => {
    console.log('[Extension] Startup');
    loadSettings().then(() => {
        connectWebSocket();
    });
});

console.log('[Extension] Background service worker loaded');
console.log('[Extension] Plugin ID:', pluginId);