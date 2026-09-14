# Nimbus — Chrome Web Store publishing checklist

## Package
- Upload `nimbus-extension-1.4.0.zip`.
- The package uses Manifest V3.
- Runtime files only are included in the upload ZIP; store/privacy assets are kept outside it.
- Version is `1.4.0`.

## Store listing
- Name: **Nimbus URL Shortener**
- Summary: **Shorten long links from your toolbar or right-click menu and keep recent short URLs handy.**
- Category: **Productivity**
- Language: **English**
- Use `store-description.md` for the long description.
- Upload `screenshot-1280x800.png` as a store screenshot.
- Use `nimbus-icon-master.png` as the source for store branding if needed.

## Privacy tab
- Single purpose: **Nimbus shortens URLs that the user explicitly selects or enters.**
- Data usage: disclose that destination URLs are transmitted to the Nimbus backend to create and operate short-link mappings.
- Declare that recent-link history is stored locally in Chrome storage.
- Host `privacy-policy.html` on a public HTTPS URL and paste that URL into the Privacy Policy field.
- Do not claim that no browsing data is processed; a URL chosen for shortening is website/browsing data.

## Permissions justification
- `activeTab`: obtain the current tab URL only after the user opens Nimbus.
- `storage`: retain up to 20 recent short links locally.
- `contextMenus`: provide the user-invoked right-click shortening command.
- `notifications`: show the result of a right-click shortening action.
- `https://nimbus-api-vqsz.onrender.com/*`: call the Nimbus shortening backend.

## Final manual checks
- Verify the production API is online and uses HTTPS.
- Test toolbar shortening, copy, recent history, history clear, and right-click shortening on a normal HTTPS page.
- Confirm notifications appear when enabled; the toolbar badge also shows success/error feedback.
- Verify the privacy-policy URL is public and accessible without sign-in.
- Enable 2-Step Verification on the publishing Google account.
- Complete Store listing and Privacy fields in the Developer Dashboard before submission.
