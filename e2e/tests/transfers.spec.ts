import { test, expect } from '@playwright/test';
import { register } from './helpers';

const timestamp = Date.now();
const user = {
  name: 'Transfer Test User',
  username: `transfer_user_${timestamp}`,
  email: `transfer_${timestamp}@test.com`,
  password: 'transfer_pass_123',
};

test.describe('Transfers', () => {
  test.beforeEach(async ({ page }) => {
    await register(page, user);

    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'Source Account');
    await page.fill('[formControlName="initialBalance"]', '1000');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Source Account')).toBeVisible();

    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'Target Account');
    await page.fill('[formControlName="initialBalance"]', '0');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Target Account')).toBeVisible();
  });

  test('creates a transfer between accounts', async ({ page }) => {
    await page.goto('/transfers');
    await page.click('text=+ New Transfer');

    await page.fill('[formControlName="name"]', 'Monthly Transfer');
    await page.fill('[formControlName="value"]', '200');
    await page.selectOption('[formControlName="sourceAccountId"]', { label: 'Source Account' });
    await page.selectOption('[formControlName="targetAccountId"]', { label: 'Target Account' });
    await page.click('button[type="submit"]');

    await expect(page.locator('text=Monthly Transfer')).toBeVisible();
  });

  test('rejects transfer with insufficient balance', async ({ page }) => {
    await page.goto('/transfers');
    await page.click('text=+ New Transfer');

    await page.fill('[formControlName="name"]', 'Overdraft Transfer');
    await page.fill('[formControlName="value"]', '9999');
    await page.selectOption('[formControlName="sourceAccountId"]', { label: 'Source Account' });
    await page.selectOption('[formControlName="targetAccountId"]', { label: 'Target Account' });
    await page.click('button[type="submit"]');

    await expect(page.locator('.error-msg, [class*="error"]')).toBeVisible();
    await expect(page.locator('text=Overdraft Transfer')).not.toBeVisible();
  });

  test('deletes a transfer', async ({ page }) => {
    await page.goto('/transfers');
    await page.click('text=+ New Transfer');
    await page.fill('[formControlName="name"]', 'Delete Transfer');
    await page.fill('[formControlName="value"]', '100');
    await page.selectOption('[formControlName="sourceAccountId"]', { label: 'Source Account' });
    await page.selectOption('[formControlName="targetAccountId"]', { label: 'Target Account' });
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Delete Transfer')).toBeVisible();

    page.on('dialog', (dialog) => dialog.accept());
    await page.locator('tr', { hasText: 'Delete Transfer' }).locator('text=Delete').click();

    await expect(page.locator('text=Delete Transfer')).not.toBeVisible();
  });
});
