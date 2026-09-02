const DEFAULT_API_BASE_URL =
    "https://nimbus-api.onrender.com";

const STORAGE_KEYS = {
    API_BASE_URL: "apiBaseUrl",
    HISTORY: "shortenedUrls"
};

const MAX_HISTORY_ITEMS = 20;

chrome.runtime.onInstalled.addListener(() => {
    chrome.contextMenus.create({
        id: "shorten-with-nimbus",
        title: "Shorten with Nimbus",
        contexts: ["page", "link"]
    });
});

chrome.runtime.onStartup.addListener(() => {
    createContextMenu();
});

chrome.contextMenus.onClicked.addListener(
    async (info, tab) => {
        let targetUrl = null;

        if (info.linkUrl) {
            targetUrl = info.linkUrl;
        } else if (tab && tab.url) {
            targetUrl = tab.url;
        }

        if (!targetUrl) {
            return;
        }

        if (isUnsupportedUrl(targetUrl)) {
            console.error(
                "Nimbus cannot shorten this URL."
            );
            return;
        }

        try {
            const parsed = new URL(targetUrl);

            if (
                !["http:", "https:"].includes(
                    parsed.protocol
                )
            ) {
                throw new Error(
                    "Only HTTP and HTTPS URLs are supported."
                );
            }

            const apiBaseUrl =
                await getApiBaseUrl();

            const response = await fetch(
                `${apiBaseUrl}/shorten`,
                {
                    method: "POST",
                    headers: {
                        "Content-Type":
                            "application/json"
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
                    data.error ||
                    "Failed to shorten URL."
                );
            }

            if (!data.short_url) {
                throw new Error(
                    "Nimbus did not return a short URL."
                );
            }

            await saveToHistory({
                shortUrl: data.short_url,
                targetUrl:
                    data.target_url ||
                    targetUrl,
                code: data.code || "",
                createdAt: Date.now()
            });

            await copyToClipboard(
                data.short_url
            );
        } catch (error) {
            console.error(
                "Nimbus context menu error:",
                error
            );
        }
    }
);

async function createContextMenu() {
    try {
        await chrome.contextMenus.removeAll();

        chrome.contextMenus.create({
            id: "shorten-with-nimbus",
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
    const stored =
        await chrome.storage.local.get(
            STORAGE_KEYS.API_BASE_URL
        );

    return normalizeBaseUrl(
        stored[STORAGE_KEYS.API_BASE_URL] ||
        DEFAULT_API_BASE_URL
    );
}

async function saveToHistory(item) {
    const stored =
        await chrome.storage.local.get(
            STORAGE_KEYS.HISTORY
        );

    const history = Array.isArray(
        stored[STORAGE_KEYS.HISTORY]
    )
        ? stored[STORAGE_KEYS.HISTORY]
        : [];

    const filteredHistory =
        history.filter(
            entry =>
                entry.shortUrl !==
                item.shortUrl
        );

    filteredHistory.unshift(item);

    await chrome.storage.local.set({
        [STORAGE_KEYS.HISTORY]:
            filteredHistory.slice(
                0,
                MAX_HISTORY_ITEMS
            )
    });
}

async function copyToClipboard(text) {
    try {
        const permission =
            await navigator.permissions.query({
                name: "clipboard-write"
            });

        if (
            permission.state === "denied"
        ) {
            throw new Error(
                "Clipboard permission denied."
            );
        }

        await navigator.clipboard.writeText(
            text
        );
    } catch (error) {
        console.error(
            "Clipboard copy failed:",
            error
        );
    }
}

function isUnsupportedUrl(url) {
    return (
        url.startsWith("chrome://") ||
        url.startsWith(
            "chrome-extension://"
        ) ||
        url.startsWith("edge://") ||
        url.startsWith("about:")
    );
}

function normalizeBaseUrl(value) {
    return String(value || "")
        .trim()
        .replace(/\/+$/, "");
}