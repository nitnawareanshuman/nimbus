const API_BASE_URL = "https://nimbus-api-vqsz.onrender.com";
const HISTORY_KEY = "shortenedUrls";
const MAX_HISTORY_ITEMS = 20;
const CONTEXT_MENU_ID = "shorten-with-nimbus";

chrome.runtime.onInstalled.addListener(createContextMenu);
chrome.runtime.onStartup.addListener(createContextMenu);

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
    if (info.menuItemId !== CONTEXT_MENU_ID) return;

    const targetUrl = info.linkUrl || tab?.url;
    if (!isValidHttpUrl(targetUrl)) {
        await notify("Nimbus", "This page or link cannot be shortened.");
        return;
    }

    try {
        setBadge(tab?.id, "…", "#1976F3");
        const response = await fetch(`${API_BASE_URL}/shorten`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ url: targetUrl })
        });

        const data = await response.json().catch(() => null);
        if (!response.ok) throw new Error(data?.error || `Nimbus returned ${response.status}.`);
        if (!data?.short_url) throw new Error("Nimbus did not return a short URL.");

        await saveToHistory({
            shortUrl: data.short_url,
            targetUrl: data.target_url || targetUrl,
            createdAt: Date.now()
        });

        setBadge(tab?.id, "✓", "#178650");
        await notify("Nimbus link created", `${data.short_url}\nOpen Nimbus to copy it.`);
    } catch (error) {
        console.error("Nimbus context-menu request failed:", error);
        setBadge(tab?.id, "!", "#BD3948");
        await notify("Nimbus couldn't shorten this link", error?.message || "Please try again.");
    } finally {
        if (tab?.id) setTimeout(() => chrome.action.setBadgeText({ tabId: tab.id, text: "" }), 4500);
    }
});

async function createContextMenu() {
    await chrome.contextMenus.removeAll();
    chrome.contextMenus.create({
        id: CONTEXT_MENU_ID,
        title: "Shorten with Nimbus",
        contexts: ["page", "link"]
    });
}

function isValidHttpUrl(value) {
    if (!value) return false;
    try {
        return ["http:", "https:"].includes(new URL(value).protocol);
    } catch {
        return false;
    }
}

async function saveToHistory(item) {
    const stored = await chrome.storage.local.get(HISTORY_KEY);
    const history = Array.isArray(stored[HISTORY_KEY]) ? stored[HISTORY_KEY] : [];
    const next = history.filter(entry => entry.shortUrl !== item.shortUrl);
    next.unshift(item);
    await chrome.storage.local.set({ [HISTORY_KEY]: next.slice(0, MAX_HISTORY_ITEMS) });
}

function setBadge(tabId, text, color) {
    if (!tabId) return;
    chrome.action.setBadgeBackgroundColor({ tabId, color });
    chrome.action.setBadgeText({ tabId, text });
}

async function notify(title, message) {
    try {
        await chrome.notifications.create({
            type: "basic",
            iconUrl: "icons/nimbus-128.png",
            title,
            message
        });
    } catch (error) {
        console.error("Nimbus notification failed:", error);
    }
}