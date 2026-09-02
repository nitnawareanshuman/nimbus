const API_BASE_URL = "https://nimbus-api-vqsz.onrender.com";

chrome.runtime.onInstalled.addListener(() => {
    chrome.contextMenus.create({
        id: "shorten-with-nimbus",
        title: "Shorten with Nimbus",
        contexts: ["page", "link"]
    });
});

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
    let targetUrl = null;

    if (info.linkUrl) {
        targetUrl = info.linkUrl;
    } else if (tab && tab.url) {
        targetUrl = tab.url;
    }

    if (!targetUrl) {
        return;
    }

    try {
        const parsed = new URL(targetUrl);

        if (!["http:", "https:"].includes(parsed.protocol)) {
            throw new Error("Only HTTP and HTTPS URLs are supported.");
        }

        const response = await fetch(`${API_BASE_URL}/shorten`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                url: targetUrl
            })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || "Failed to shorten URL.");
        }

        if (!data.short_url) {
            throw new Error("Nimbus did not return a short URL.");
        }

        await navigator.clipboard.writeText(data.short_url);
    } catch (error) {
        console.error("Nimbus context menu error:", error);
    }
});