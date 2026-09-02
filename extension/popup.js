const DEFAULT_API_BASE_URL = "https://nimbus-api-vqsz.onrender.com";
const STORAGE_KEYS = {
    API_BASE_URL: "apiBaseUrl",
    HISTORY: "shortenedUrls"
};

const MAX_HISTORY_ITEMS = 20;

const urlInput = document.getElementById("urlInput");
const shortenBtn = document.getElementById("shortenBtn");
const status = document.getElementById("status");
const result = document.getElementById("result");
const shortUrl = document.getElementById("shortUrl");
const copyBtn = document.getElementById("copyBtn");
const historySection = document.getElementById("history");
const historyList = document.getElementById("historyList");
const clearHistoryBtn = document.getElementById("clearHistoryBtn");

document.addEventListener("DOMContentLoaded", async () => {
    await loadCurrentTabUrl();
    await loadHistory();
});

shortenBtn.addEventListener("click", shortenUrl);
copyBtn.addEventListener("click", copyShortUrl);
clearHistoryBtn.addEventListener("click", clearHistory);

async function loadCurrentTabUrl() {
    try {
        const tabs = await chrome.tabs.query({
            active: true,
            currentWindow: true
        });

        const currentTab = tabs[0];

        if (!currentTab || !currentTab.url) {
            return;
        }

        const url = currentTab.url;

        if (isUnsupportedUrl(url)) {
            urlInput.value = "";
            return;
        }

        urlInput.value = url;
    } catch (error) {
        console.error("Failed to get current tab URL:", error);
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

async function getApiBaseUrl() {
    const result = await chrome.storage.local.get(
        STORAGE_KEYS.API_BASE_URL
    );

    return normalizeBaseUrl(
        result[STORAGE_KEYS.API_BASE_URL] || DEFAULT_API_BASE_URL
    );
}

function normalizeBaseUrl(value) {
    return String(value || "")
        .trim()
        .replace(/\/+$/, "");
}

async function shortenUrl() {
    const url = urlInput.value.trim();

    clearStatus();
    result.classList.add("hidden");

    if (!url) {
        showStatus("Please enter a URL.", "error");
        return;
    }

    if (isUnsupportedUrl(url)) {
        showStatus(
            "Chrome internal pages cannot be shortened.",
            "error"
        );
        return;
    }

    try {
        const parsed = new URL(url);

        if (!["http:", "https:"].includes(parsed.protocol)) {
            throw new Error(
                "Only HTTP and HTTPS URLs are supported."
            );
        }
    } catch {
        showStatus(
            "Please enter a valid HTTP or HTTPS URL.",
            "error"
        );
        return;
    }

    setLoading(true);

    try {
        const apiBaseUrl = await getApiBaseUrl();

        const response = await fetch(
            `${apiBaseUrl}/shorten`,
            {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    url: url
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
                `Request failed with status ${response.status}.`
            );
        }

        if (!data.short_url) {
            throw new Error(
                "The server did not return a short URL."
            );
        }

        shortUrl.value = data.short_url;
        result.classList.remove("hidden");

        await saveToHistory({
            shortUrl: data.short_url,
            targetUrl: data.target_url || url,
            code: data.code || "",
            createdAt: Date.now()
        });

        showStatus(
            "URL shortened successfully.",
            "success"
        );

        await loadHistory();
    } catch (error) {
        console.error("Shorten request failed:", error);

        let message = error.message;

        if (
            error instanceof TypeError &&
            error.message.includes("fetch")
        ) {
            message =
                "Unable to connect to Nimbus. Check your API URL or internet connection.";
        }

        showStatus(
            message || "Unable to connect to Nimbus.",
            "error"
        );
    } finally {
        setLoading(false);
    }
}

function setLoading(isLoading) {
    shortenBtn.disabled = isLoading;

    if (isLoading) {
        shortenBtn.textContent = "Shortening...";
        showStatus("Contacting Nimbus...", "loading");
    } else {
        shortenBtn.textContent = "Shorten URL";
    }
}

async function copyShortUrl() {
    const value = shortUrl.value.trim();

    if (!value) {
        return;
    }

    try {
        await copyText(value);
        showCopyFeedback();
    } catch (error) {
        console.error("Copy failed:", error);

        showStatus(
            "Unable to copy the URL.",
            "error"
        );
    }
}

async function copyHistoryUrl(url, button) {
    try {
        await copyText(url);

        const originalText = button.textContent;

        button.textContent = "Copied!";
        button.disabled = true;

        setTimeout(() => {
            button.textContent = originalText;
            button.disabled = false;
        }, 1200);
    } catch (error) {
        console.error("History copy failed:", error);

        showStatus(
            "Unable to copy the URL.",
            "error"
        );
    }
}

async function copyText(value) {
    if (navigator.clipboard) {
        await navigator.clipboard.writeText(value);
        return;
    }

    shortUrl.select();

    const successful = document.execCommand("copy");

    if (!successful) {
        throw new Error("Copy command failed.");
    }
}

function showCopyFeedback() {
    const originalText = copyBtn.textContent;

    copyBtn.textContent = "Copied!";
    copyBtn.disabled = true;

    setTimeout(() => {
        copyBtn.textContent = originalText;
        copyBtn.disabled = false;
    }, 1500);
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

    const trimmedHistory = filteredHistory.slice(
        0,
        MAX_HISTORY_ITEMS
    );

    await chrome.storage.local.set({
        [STORAGE_KEYS.HISTORY]: trimmedHistory
    });
}

async function loadHistory() {
    const stored = await chrome.storage.local.get(
        STORAGE_KEYS.HISTORY
    );

    const history = Array.isArray(
        stored[STORAGE_KEYS.HISTORY]
    )
        ? stored[STORAGE_KEYS.HISTORY]
        : [];

    renderHistory(history);
}

function renderHistory(history) {
    historyList.innerHTML = "";

    if (history.length === 0) {
        historySection.classList.add("hidden");
        return;
    }

    historySection.classList.remove("hidden");

    history.forEach(item => {
        const historyItem = document.createElement("div");
        historyItem.className = "history-item";

        const content = document.createElement("div");
        content.className = "history-content";

        const shortLink = document.createElement("a");
        shortLink.className = "history-short-url";
        shortLink.href = item.shortUrl;
        shortLink.target = "_blank";
        shortLink.rel = "noopener noreferrer";
        shortLink.textContent = item.shortUrl;

        const targetUrl = document.createElement("div");
        targetUrl.className = "history-target-url";
        targetUrl.title = item.targetUrl;
        targetUrl.textContent = item.targetUrl;

        const date = document.createElement("div");
        date.className = "history-date";
        date.textContent = formatDate(item.createdAt);

        content.appendChild(shortLink);
        content.appendChild(targetUrl);
        content.appendChild(date);

        const copyHistoryButton = document.createElement("button");
        copyHistoryButton.className = "history-copy-btn";
        copyHistoryButton.type = "button";
        copyHistoryButton.textContent = "Copy";

        copyHistoryButton.addEventListener("click", () => {
            copyHistoryUrl(
                item.shortUrl,
                copyHistoryButton
            );
        });

        historyItem.appendChild(content);
        historyItem.appendChild(copyHistoryButton);

        historyList.appendChild(historyItem);
    });
}

async function clearHistory() {
    const confirmed = confirm(
        "Clear all Nimbus URL history?"
    );

    if (!confirmed) {
        return;
    }

    await chrome.storage.local.remove(
        STORAGE_KEYS.HISTORY
    );

    renderHistory([]);

    showStatus(
        "History cleared.",
        "success"
    );
}

function formatDate(timestamp) {
    if (!timestamp) {
        return "";
    }

    return new Date(timestamp).toLocaleString(
        undefined,
        {
            dateStyle: "short",
            timeStyle: "short"
        }
    );
}

function showStatus(message, type = "") {
    status.textContent = message;
    status.className = `status ${type}`;
}

function clearStatus() {
    status.textContent = "";
    status.className = "status";
}