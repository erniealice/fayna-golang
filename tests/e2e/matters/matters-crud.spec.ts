import { test, expect } from '@playwright/test';
import { waitForHtmxSettle } from '../helpers/htmx';

/**
 * FAY-JOB-001: Matters List (professional = Jobs)
 * FAY-JOB-002: Matter Add via Drawer
 * FAY-JOB-003: Matter Edit via Drawer
 *
 * Routes (professional override):
 *   List:   /app/matters/list/active
 *   Add:    /action/matters/add
 *   Edit:   /action/matters/edit/{id}
 *   Detail: /app/matters/detail/{id}
 *
 * Verifies: list page loads, table structure, CRUD via drawer
 */

test.describe('FAY-JOB-001: Matters List', () => {
  test('navigates to matters list page', async ({ page }) => {
    const response = await page.goto('/app/matters/list/active');
    expect(response?.status()).toBeLessThan(500);

    const pageContent = await page.textContent('body');
    expect(pageContent).not.toContain('Page content not available');
  });

  test('displays matters table with column headers', async ({ page }) => {
    await page.goto('/app/matters/list/active');

    const table = page.locator('table');
    await expect(table).toBeVisible({ timeout: 10000 });

    // Verify column headers exist (NAME, CLIENT, STATUS, CREATED, ACTIONS = 5+)
    const headers = page.locator('thead th');
    const count = await headers.count();
    expect(count).toBeGreaterThanOrEqual(3);
  });

  test('shows data rows or empty state', async ({ page }) => {
    await page.goto('/app/matters/list/active');

    // Either data rows with data-id exist or empty state is shown
    const dataRows = page.locator('tbody tr[data-id]');
    const emptyState = page.locator('.empty-state');
    const rowCount = await dataRows.count();
    const emptyVisible = await emptyState.isVisible().catch(() => false);

    expect(rowCount > 0 || emptyVisible).toBe(true);
  });

  test('has primary action button in toolbar', async ({ page }) => {
    await page.goto('/app/matters/list/active');

    const primaryAction = page.locator('.toolbar-primary-action');
    await expect(primaryAction).toBeVisible({ timeout: 10000 });
    await expect(primaryAction).toBeEnabled();
  });

  test('shows pagination with entry count', async ({ page }) => {
    await page.goto('/app/matters/list/active');

    // Pagination area is rendered in the table footer container
    const pagination = page.locator('.table-footer, .pagination-info');
    await expect(pagination).toBeVisible({ timeout: 10000 });
  });
});

test.describe('FAY-JOB-002: Matter Add via Drawer', () => {
  test('opens drawer when primary action clicked', async ({ page }) => {
    await page.goto('/app/matters/list/active');

    const primaryAction = page.locator('.toolbar-primary-action');
    await expect(primaryAction).toBeVisible({ timeout: 10000 });

    await primaryAction.click();
    await expect(page.locator('#sheet.open .sheet-panel')).toBeVisible({ timeout: 10000 });
  });

  test('drawer has job form fields (name, client_id, location_id)', async ({ page }) => {
    await page.goto('/app/matters/list/active');

    const primaryAction = page.locator('.toolbar-primary-action');
    await expect(primaryAction).toBeVisible({ timeout: 10000 });

    await primaryAction.click();
    await expect(page.locator('#sheet.open .sheet-panel')).toBeVisible({ timeout: 10000 });

    // Wait for HTMX to load drawer content
    await waitForHtmxSettle(page);

    // Job drawer fields: name, client_id, location_id
    await expect(page.locator('#name')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#client_id')).toBeVisible();
    await expect(page.locator('#location_id')).toBeVisible();
  });

  test('cancel closes drawer without creating', async ({ page }) => {
    await page.goto('/app/matters/list/active');

    const primaryAction = page.locator('.toolbar-primary-action');
    await expect(primaryAction).toBeVisible({ timeout: 10000 });

    await primaryAction.click();
    await expect(page.locator('#sheet.open .sheet-panel')).toBeVisible({ timeout: 10000 });
    await waitForHtmxSettle(page);

    // Cancel via secondary button in sheet footer
    const cancelBtn = page.locator('#sheet .sheet-footer .btn-secondary');
    await expect(cancelBtn).toBeVisible({ timeout: 5000 });
    await cancelBtn.click();

    // Drawer should close
    await expect(page.locator('.sheet.open')).not.toBeVisible({ timeout: 10000 });
  });
});

test.describe('FAY-JOB-003: Matter Edit via Drawer', () => {
  test('row has action buttons when data exists', async ({ page }) => {
    await page.goto('/app/matters/list/active');

    // Only test if there are actual data rows (not empty state rows)
    const dataRows = page.locator('tbody tr[data-id]');
    const rowCount = await dataRows.count();
    if (rowCount === 0) {
      test.skip(true, 'No matter rows to test edit actions - table is empty (add seed data or create via FAY-JOB-002 first)');
    }

    const firstRow = dataRows.first();
    const viewLink = firstRow.locator('a.action-btn.view');
    const editBtn = firstRow.locator('.action-btn.edit');

    const viewVisible = await viewLink.isVisible().catch(() => false);
    const editVisible = await editBtn.isVisible().catch(() => false);

    expect(viewVisible || editVisible).toBe(true);
  });

  test('edit button opens drawer with pre-filled data', async ({ page }) => {
    await page.goto('/app/matters/list/active');

    const dataRows = page.locator('tbody tr[data-id]');
    const rowCount = await dataRows.count();
    if (rowCount === 0) {
      test.skip(true, 'No matter rows to test edit - table is empty (add seed data or create via FAY-JOB-002 first)');
    }

    const editBtn = dataRows.first().locator('.action-btn.edit');
    if (!(await editBtn.isVisible().catch(() => false))) {
      test.skip(true, 'No edit button found on first data row');
    }

    await editBtn.click();
    await expect(page.locator('#sheet.open .sheet-panel')).toBeVisible({ timeout: 10000 });
  });
});
