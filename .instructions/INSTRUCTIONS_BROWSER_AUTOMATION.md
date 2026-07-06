# Browser Automation

## Focus Policy

* Browser scripting and page inspection must be **no-focus by default**.
* Do not activate, raise, focus, or make Safari/Chrome/front browser windows frontmost unless the user explicitly asks to see or interact with the browser.
* Prefer background-capable tools and APIs such as `mac-safari-session open-bg`, `snapshot`, `run-js`, or equivalent browser-session helpers that preserve the user's active window.
* Do not use AppleScript patterns such as `tell application "Safari" to activate`, focused new windows/tabs, UI scripting clicks, or frontmost-window manipulation just to inspect DOM, links, text, downloads, or authenticated page state.
* If a browser action truly needs human interaction, for example passkey confirmation, SSO, CAPTCHA, file picker, or a permission prompt, say what is needed and let the user open or focus the browser unless they explicitly ask the agent to do it.

## Authenticated Sessions

* Reuse the user's authenticated browser session only through same-origin page context or approved browser-session tooling.
* Never export, print, persist, or bypass protections around cookies, local storage, session storage, authorization headers, tokens, or other browser secrets.
* Persist only sanitized page content, URLs, downloaded files, metadata, and notes needed for the task.
