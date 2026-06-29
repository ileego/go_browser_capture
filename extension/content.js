console.log('[Content Script] Loaded on:', window.location.href);

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    console.log('[Content Script] Received message:', request);

    switch (request.action) {
        case 'capture':
            handleCaptureRequest(request, sendResponse);
            return true;

        case 'check_element':
            handleCheckElement(request, sendResponse);
            return true;

        case 'highlight':
            handleHighlight(request);
            sendResponse({ success: true });
            break;
    }
});

function handleCaptureRequest(request, sendResponse) {
    const { selector, timeout = 5000 } = request;
    
    const startTime = Date.now();
    const pollInterval = 100;

    const checkElement = () => {
        const element = document.querySelector(selector);
        
        if (element) {
            const result = extractElementData(element);
            sendResponse({
                success: true,
                data: result
            });
            return;
        }

        if (Date.now() - startTime >= timeout) {
            sendResponse({
                success: false,
                error: 'Timeout: Element not found'
            });
            return;
        }

        setTimeout(checkElement, pollInterval);
    };

    checkElement();
}

function handleCheckElement(request, sendResponse) {
    const { selector } = request;
    
    const element = document.querySelector(selector);
    
    if (element) {
        sendResponse({
            found: true,
            tagName: element.tagName,
            id: element.id,
            className: element.className
        });
    } else {
        sendResponse({
            found: false
        });
    }
}

function handleHighlight(request) {
    const { selector } = request;
    
    try {
        document.querySelectorAll(selector).forEach(el => {
            el.style.outline = '2px solid #007bff';
            el.style.outlineOffset = '2px';
            
            setTimeout(() => {
                el.style.outline = '';
                el.style.outlineOffset = '';
            }, 3000);
        });
    } catch (error) {
        console.error('[Content Script] Highlight error:', error);
    }
}

function extractElementData(element) {
    const attributes = {};
    for (let i = 0; i < element.attributes.length; i++) {
        const attr = element.attributes[i];
        attributes[attr.name] = attr.value;
    }

    return {
        html: element.innerHTML,
        text: element.textContent,
        innerText: element.innerText,
        outerHTML: element.outerHTML,
        attributes: attributes,
        tagName: element.tagName,
        id: element.id,
        className: element.className,
        childElementCount: element.children.length,
        textContentLength: element.textContent.length
    };
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        console.log('[Content Script] DOM fully loaded');
    });
} else {
    console.log('[Content Script] DOM already loaded');
}

let lastUrl = location.href;
new MutationObserver(() => {
    const url = location.href;
    if (url !== lastUrl) {
        lastUrl = url;
        console.log('[Content Script] URL changed:', url);
    }
}).observe(document, { subtree: true, childList: true });