const DEFAULT_API_BASE_URL =
    "https://nimbus-api-vqsz.onrender.com";

const STORAGE_KEYS = {
    API_BASE_URL: "apiBaseUrl",
    HISTORY: "shortenedUrls"
};

const MAX_HISTORY_ITEMS = 20;
const CONTEXT_MENU_ID = "shorten-with-nimbus";

chrome.runtime.onInstalled.addListener(async () => {
    await createContextMenu();
});

chrome.runtime.onStartup.addListener(async () => {
    await createContextMenu();
});

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
    if (info.menuItemId !== CONTEXT_MENU_ID) {
        return;
    }

    let targetUrl = null;

    if (info.linkUrl) {
        targetUrl = info.linkUrl;
    } else if (tab?.url) {
        targetUrl = tab.url;
    }

    if (!targetUrl) {
        await showNotification(
            "Nimbus",
            "Could not determine which URL to shorten."
        );
        return;
    }

    if (isUnsupportedUrl(targetUrl)) {
        await showNotification(
            "Nimbus",
            "This page cannot be shortened."
        );
        return;
    }

    try {
        const parsed = new URL(targetUrl);

        if (!["http:", "https:"].includes(parsed.protocol)) {
            throw new Error(
                "Only HTTP and HTTPS URLs are supported."
            );
        }

        const apiBaseUrl = await getApiBaseUrl();

        const response = await fetch(
            `${apiBaseUrl}/shorten`,
            {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    url: targetUrl
                })
            }
        );

        let data;

        try {
            data = await response.json();
        } catch {
            throw new Error(
                "The server returned an invalid response."
            );
        }

        if (!response.ok) {
            throw new Error(
                data.error || "Failed to shorten URL."
            );
        }

        if (!data.short_url) {
            throw new Error(
                "Nimbus did not return a short URL."
            );
        }

        await saveToHistory({
            shortUrl: data.short_url,
            targetUrl: data.target_url || targetUrl,
            code: data.code || "",
            createdAt: Date.now()
        });

        const copied = await copyUsingCurrentTab(
            tab?.id,
            data.short_url
        );

        if (copied) {
            await showNotification(
                "Nimbus",
                `Short URL copied: ${data.short_url}`
            );
        } else {
            await showNotification(
                "Nimbus URL created",
                data.short_url
            );
        }
    } catch (error) {
        console.error(
            "Nimbus context menu error:",
            error
        );

        await showNotification(
            "Nimbus error",
            error instanceof Error
                ? error.message
                : "Failed to shorten URL."
        );
    }
});

async function createContextMenu() {
    try {
        await chrome.contextMenus.removeAll();

        chrome.contextMenus.create({
            id: CONTEXT_MENU_ID,
            title: "Shorten with Nimbus",
            contexts: ["page", "link"]
        });
    } catch (error) {
        console.error(
            "Failed to create context menu:",
            error
        );
    }
}

async function getApiBaseUrl() {
    const stored = await chrome.storage.local.get(
        STORAGE_KEYS.API_BASE_URL
    );

    const value =
        stored[STORAGE_KEYS.API_BASE_URL] ||
        DEFAULT_API_BASE_URL;

    return normalizeBaseUrl(value);
}

async function saveToHistory(item) {
    const stored = await chrome.storage.local.get(
        STORAGE_KEYS.HISTORY
    );

    const history = Array.isArray(
        stored[STORAGE_KEYS.HISTORY]
    )
        ? stored[STORAGE_KEYS.HISTORY]
        : [];

    const filteredHistory = history.filter(
        entry => entry.shortUrl !== item.shortUrl
    );

    filteredHistory.unshift(item);

    await chrome.storage.local.set({
        [STORAGE_KEYS.HISTORY]:
            filteredHistory.slice(0, MAX_HISTORY_ITEMS)
    });
}

async function copyUsingCurrentTab(tabId, text) {
    if (!tabId) {
        return false;
    }

    try {
        const results =
            await chrome.scripting.executeScript({
                target: {
                    tabId
                },

                func: async value => {
                    try {
                        await navigator.clipboard.writeText(
                            value
                        );

                        return true;
                    } catch {
                        const textarea =
                            document.createElement("textarea");

                        textarea.value = value;

                        textarea.style.position = "fixed";
                        textarea.style.opacity = "0";

                        document.body.appendChild(textarea);

                        textarea.focus();
                        textarea.select();

                        const copied =
                            document.execCommand("copy");

                        textarea.remove();

                        return copied;
                    }
                },

                args: [text]
            });

        return results?.[0]?.result === true;
    } catch (error) {
        console.error(
            "Failed to copy shortened URL:",
            error
        );

        return false;
    }
}

async function showNotification(title, message) {
    try {
        await chrome.notifications.create({
            type: "basic",
            iconUrl: "icons/nimbus-128.png",
            title,
            message
        });
    } catch (error) {
        console.error(
            "Failed to show notification:",
            error
        );
    }
}

function isUnsupportedUrl(url) {
    return (
        url.startsWith("chrome://") ||
        url.startsWith("chrome-extension://") ||
        url.startsWith("edge://") ||
        url.startsWith("about:")
    );
}

function normalizeBaseUrl(value) {
    return String(value || "")
        .trim()
        .replace(/\/+$/, "");
}