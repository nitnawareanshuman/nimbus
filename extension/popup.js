const API_BASE_URL = "https://nimbus-api-vqsz.onrender.com";
const HISTORY_KEY = "shortenedUrls";
const MAX_HISTORY_ITEMS = 20;

const urlInput = document.getElementById("urlInput");
const useCurrentBtn = document.getElementById("useCurrentBtn");
const shortenBtn = document.getElementById("shortenBtn");
const status = document.getElementById("status");
const result = document.getElementById("result");
const shortUrl = document.getElementById("shortUrl");
const copyBtn = document.getElementById("copyBtn");
const historySection = document.getElementById("history");
const historyList = document.getElementById("historyList");
const clearHistoryBtn = document.getElementById("clearHistoryBtn");

window.addEventListener("DOMContentLoaded", async () => {
    await Promise.all([loadCurrentTabUrl(), loadHistory()]);
});

useCurrentBtn.addEventListener("click", loadCurrentTabUrl);
shortenBtn.addEventListener("click", shortenUrl);
copyBtn.addEventListener("click", copyShortUrl);
clearHistoryBtn.addEventListener("click", clearHistory);
urlInput.addEventListener("keydown", event => {
    if ((event.ctrlKey || event.metaKey) && event.key === "Enter") shortenUrl();
});

async function loadCurrentTabUrl() {
    try {
        const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
        if (!tab?.url || isUnsupportedUrl(tab.url)) {
            showStatus("This Chrome page cannot be shortened.", "error");
            return;
        }
        urlInput.value = tab.url;
        clearStatus();
    } catch (error) {
        console.error("Nimbus could not read the current tab:", error);
        showStatus("Could not read the current tab.", "error");
    }
}

function isUnsupportedUrl(url) {
    return /^(chrome|chrome-extension|edge|about|view-source):/i.test(url);
}

function validateHttpUrl(value) {
    try {
        return ["http:", "https:"].includes(new URL(value).protocol);
    } catch {
        return false;
    }
}

async function shortenUrl() {
    const url = urlInput.value.trim();
    result.classList.add("hidden");
    clearStatus();

    if (!validateHttpUrl(url) || isUnsupportedUrl(url)) {
        showStatus("Enter a valid HTTP or HTTPS URL.", "error");
        return;
    }

    setLoading(true);
    try {
        const response = await fetch(`${API_BASE_URL}/shorten`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ url })
        });

        const data = await response.json().catch(() => null);
        if (!response.ok) throw new Error(data?.error || `Nimbus returned ${response.status}.`);
        if (!data?.short_url) throw new Error("Nimbus did not return a short URL.");

        shortUrl.value = data.short_url;
        result.classList.remove("hidden");
        await saveToHistory({
            shortUrl: data.short_url,
            targetUrl: data.target_url || url,
            createdAt: Date.now()
        });
        await loadHistory();
        showStatus("Done — your short link is ready.", "success");
    } catch (error) {
        console.error("Nimbus shorten request failed:", error);
        showStatus(error?.message || "Unable to connect to Nimbus.", "error");
    } finally {
        setLoading(false);
    }
}

function setLoading(loading) {
    shortenBtn.disabled = loading;
    shortenBtn.querySelector("span").textContent = loading ? "Shortening…" : "Shorten URL";
    if (loading) showStatus("Creating your short link…", "loading");
}

async function copyShortUrl() {
    const value = shortUrl.value.trim();
    if (!value) return;
    try {
        await navigator.clipboard.writeText(value);
        copyBtn.disabled = true;
        copyBtn.setAttribute("aria-label", "Copied");
        showStatus("Copied to clipboard.", "success");
        setTimeout(() => {
            copyBtn.disabled = false;
            copyBtn.setAttribute("aria-label", "Copy short URL");
        }, 1000);
    } catch (error) {
        console.error("Nimbus copy failed:", error);
        showStatus("Could not copy the short URL.", "error");
    }
}

async function saveToHistory(item) {
    const stored = await chrome.storage.local.get(HISTORY_KEY);
    const history = Array.isArray(stored[HISTORY_KEY]) ? stored[HISTORY_KEY] : [];
    const next = history.filter(entry => entry.shortUrl !== item.shortUrl);
    next.unshift(item);
    await chrome.storage.local.set({ [HISTORY_KEY]: next.slice(0, MAX_HISTORY_ITEMS) });
}

async function loadHistory() {
    const stored = await chrome.storage.local.get(HISTORY_KEY);
    const history = Array.isArray(stored[HISTORY_KEY]) ? stored[HISTORY_KEY] : [];
    renderHistory(history);
}

function renderHistory(history) {
    historyList.replaceChildren();
    historySection.classList.toggle("hidden", history.length === 0);

    for (const item of history.slice(0, 5)) {
        const row = document.createElement("div");
        row.className = "history-item";

        const content = document.createElement("div");
        content.className = "history-content";

        const link = document.createElement("a");
        link.className = "history-short-url";
        link.href = item.shortUrl;
        link.target = "_blank";
        link.rel = "noopener noreferrer";
        link.textContent = item.shortUrl;

        const target = document.createElement("div");
        target.className = "history-target-url";
        target.title = item.targetUrl || "";
        target.textContent = item.targetUrl || "";

        const date = document.createElement("div");
        date.className = "history-date";
        date.textContent = item.createdAt ? new Date(item.createdAt).toLocaleString() : "";

        const copy = document.createElement("button");
        copy.className = "history-copy-btn";
        copy.type = "button";
        copy.textContent = "Copy";
        copy.addEventListener("click", async () => {
            try {
                await navigator.clipboard.writeText(item.shortUrl);
                copy.textContent = "Copied";
                copy.disabled = true;
                setTimeout(() => { copy.textContent = "Copy"; copy.disabled = false; }, 900);
            } catch {
                showStatus("Could not copy the short URL.", "error");
            }
        });

        content.append(link, target, date);
        row.append(content, copy);
        historyList.append(row);
    }
}

async function clearHistory() {
    await chrome.storage.local.remove(HISTORY_KEY);
    renderHistory([]);
    showStatus("Recent links cleared.", "success");
}

function showStatus(message, type = "") {
    status.textContent = message;
    status.className = `status ${type}`.trim();
}

function clearStatus() {
    status.textContent = "";
    status.className = "status";
}