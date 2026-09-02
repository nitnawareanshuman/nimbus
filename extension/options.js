const DEFAULT_API_BASE_URL =
    "https://nimbus-api-vqsz.onrender.com";

const STORAGE_KEY = "apiBaseUrl";

const apiBaseUrlInput =
    document.getElementById("apiBaseUrl");

const saveBtn =
    document.getElementById("saveBtn");

const status =
    document.getElementById("status");

const currentApiUrl =
    document.getElementById("currentApiUrl");

document.addEventListener(
    "DOMContentLoaded",
    loadSettings
);

saveBtn.addEventListener(
    "click",
    saveSettings
);

async function loadSettings() {
    const stored =
        await chrome.storage.local.get(
            STORAGE_KEY
        );

    const apiBaseUrl =
        normalizeBaseUrl(
            stored[STORAGE_KEY] ||
            DEFAULT_API_BASE_URL
        );

    apiBaseUrlInput.value = apiBaseUrl;
    currentApiUrl.textContent = apiBaseUrl;
}

async function saveSettings() {
    const value =
        normalizeBaseUrl(
            apiBaseUrlInput.value
        );

    clearStatus();

    if (!value) {
        showStatus(
            "Please enter an API URL.",
            "error"
        );
        return;
    }

    try {
        const parsed = new URL(value);

        if (!["http:", "https:"].includes(
            parsed.protocol
        )) {
            throw new Error();
        }
    } catch {
        showStatus(
            "Please enter a valid HTTP or HTTPS URL.",
            "error"
        );
        return;
    }

    saveBtn.disabled = true;
    saveBtn.textContent = "Saving...";

    try {
        await chrome.storage.local.set({
            [STORAGE_KEY]: value
        });

        apiBaseUrlInput.value = value;
        currentApiUrl.textContent = value;

        showStatus(
            "Settings saved successfully.",
            "success"
        );
    } catch (error) {
        console.error(
            "Failed to save settings:",
            error
        );

        showStatus(
            "Unable to save settings.",
            "error"
        );
    } finally {
        saveBtn.disabled = false;
        saveBtn.textContent = "Save settings";
    }
}

function normalizeBaseUrl(value) {
    return String(value || "")
        .trim()
        .replace(/\/+$/, "");
}

function showStatus(message, type = "") {
    status.textContent = message;
    status.className = `status ${type}`;
}

function clearStatus() {
    status.textContent = "";
    status.className = "status";
}