# Nimbus URL Shortener

Turn long links into clean, shareable URLs without leaving the page you are on.

Nimbus is a focused Chrome extension for one job: shortening URLs quickly. Open it from the toolbar to shorten the current tab or paste another HTTP/HTTPS link. You can also right-click a page or link and choose **Shorten with Nimbus**.

## Highlights

- **Current-tab shortening** — Nimbus can prefill the URL of the tab where you opened it.
- **Paste any web link** — Shorten any valid HTTP or HTTPS URL.
- **Right-click shortcut** — Shorten pages and links directly from Chrome's context menu.
- **One-click copy** — Copy a newly created short link from the popup.
- **Recent links** — Keep up to 20 recent results in local Chrome storage for quick reuse.
- **Minimal permissions** — Nimbus requests only the browser capabilities needed for its core shortening workflow.

## Privacy

Nimbus sends a URL to the Nimbus backend only when you explicitly ask to shorten it. Recent-link history is stored locally in your Chrome profile. Nimbus does not continuously monitor browsing activity and does not sell user data or use submitted URLs for targeted advertising.

Before publishing, host `privacy-policy.html` at a public HTTPS URL and enter that URL in the Chrome Web Store Privacy tab.

## Permissions explanation

- **Active tab:** reads the current tab URL when you invoke Nimbus.
- **Storage:** stores recent shortened links locally.
- **Context menus:** adds the “Shorten with Nimbus” right-click command.
- **Notifications:** confirms success or failure for right-click shortening.
- **Host permission:** communicates only with the Nimbus URL-shortening API.
