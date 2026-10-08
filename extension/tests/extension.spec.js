const { test, expect } = require('./fixtures');

test('extension background script is running and popup opens', async ({ page, extensionId }) => {
  // Wait for the background service worker or extension to be ready.
  // The fixture `extensionId` resolves after the service worker or background script is ready.
  expect(extensionId).toBeTruthy();

  // Navigate to the extension's popup HTML
  await page.goto(`chrome-extension://${extensionId}/options/options.html`);

  // Verify the page title or a specific element to ensure it loaded correctly
  await expect(page).toHaveTitle(/Snipo/i);
});
