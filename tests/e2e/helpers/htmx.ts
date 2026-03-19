import { Page } from '@playwright/test';

/** Wait for all HTMX requests to complete */
export async function waitForHtmx(page: Page, timeout = 5000) {
  await page.waitForFunction(() => {
    return document.querySelectorAll('.htmx-request').length === 0;
  }, { timeout });
}

/** Wait for HTMX request + settle + swap to complete */
export async function waitForHtmxSettle(page: Page, timeout = 5000) {
  await page.waitForFunction(() => {
    const busy = document.querySelectorAll(
      '.htmx-request, .htmx-settling, .htmx-swapping'
    );
    return busy.length === 0;
  }, { timeout });
}

/** Click an element and wait for HTMX to settle */
export async function clickAndWaitHtmx(page: Page, selector: string) {
  await page.click(selector);
  await waitForHtmxSettle(page);
}

/** Ensure htmx is initialized on the page */
export async function ensureHtmxReady(page: Page) {
  await page.waitForFunction(() => !!(window as any).htmx);
}
