const API_BASE_URL = "https://nimbus-api-vqsz.onrender.com";

const urlInput = document.getElementById("urlInput");
const shortenBtn = document.getElementById("shortenBtn");
const status = document.getElementById("status");
const result = document.getElementById("result");
const shortUrl = document.getElementById("shortUrl");
const copyBtn = document.getElementById("copyBtn");

document.addEventListener("DOMContentLoaded", async () => {
    await loadCurrentTabUrl();
});

shortenBtn.addEventListener("click", shortenUrl);
copyBtn.addEventListener("click", copyShortUrl);

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

        // Chrome internal pages cannot be shortened through the API.
        if (
            url.startsWith("chrome://") ||
            url.startsWith("chrome-extension://") ||
            url.startsWith("edge://") ||
            url.startsWith("about:")
        ) {
            urlInput.value = "";
            return;
        }

        urlInput.value = url;
    } catch (error) {
        console.error("Failed to get current tab URL:", error);
    }
}

async function shortenUrl() {
    const url = urlInput.value.trim();

    clearStatus();
    result.classList.add("hidden");

    if (!url) {
        showStatus("Please enter a URL.", "error");
        return;
    }

    try {
        const parsed = new URL(url);

        if (!["http:", "https:"].includes(parsed.protocol)) {
            throw new Error("Only HTTP and HTTPS URLs are supported.");
        }
    } catch {
        showStatus("Please enter a valid HTTP or HTTPS URL.", "error");
        return;
    }

    shortenBtn.disabled = true;
    shortenBtn.textContent = "Shortening...";

    try {
        const response = await fetch(`${API_BASE_URL}/shorten`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                url: url
            })
        });

        let data;

        try {
            data = await response.json();
        } catch {
            throw new Error("The server returned an invalid response.");
        }

        if (!response.ok) {
            throw new Error(
                data.error || `Request failed with status ${response.status}.`
            );
        }

        if (!data.short_url) {
            throw new Error("The server did not return a short URL.");
        }

        shortUrl.value = data.short_url;
        result.classList.remove("hidden");

        showStatus("URL shortened successfully.", "success");
    } catch (error) {
        console.error("Shorten request failed:", error);

        showStatus(
            error.message || "Unable to connect to Nimbus.",
            "error"
        );
    } finally {
        shortenBtn.disabled = false;
        shortenBtn.textContent = "Shorten URL";
    }
}

async function copyShortUrl() {
    const value = shortUrl.value.trim();

    if (!value) {
        return;
    }

    try {
        await navigator.clipboard.writeText(value);

        const originalText = copyBtn.textContent;
        copyBtn.textContent = "Copied!";

        setTimeout(() => {
            copyBtn.textContent = originalText;
        }, 1500);
    } catch (error) {
        console.error("Copy failed:", error);

        shortUrl.select();
        document.execCommand("copy");

        copyBtn.textContent = "Copied!";

        setTimeout(() => {
            copyBtn.textContent = "Copy";
        }, 1500);
    }
}

function showStatus(message, type = "") {
    status.textContent = message;
    status.className = `status ${type}`;
}

function clearStatus() {
    status.textContent = "";
    status.className = "status";
}