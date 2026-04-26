import { test, expect } from '@playwright/test';
import { register } from './helpers';

const timestamp = Date.now();
const user = {
  name: 'Dashboard Test User',
  username: `dash_user_${timestamp}`,
  email: `dash_${timestamp}@test.com`,
  password: 'dash_pass_123',
};

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await register(page, user);
  });

  test('shows empty state when no bank accounts exist', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page.locator('text=No bank accounts').or(page.locator('[class*="empty"]'))).toBeVisible();
  });

  test('shows bank account summary after creating an account', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'My Main Account');
    await page.fill('[formControlName="initialBalance"]', '2500');
    await page.click('button[type="submit"]');

    await page.goto('/dashboard');
    await expect(page.locator('text=My Main Account')).toBeVisible();
  });

  test('shows latest transactions in bank account summary', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'Active Account');
    await page.fill('[formControlName="initialBalance"]', '1000');
    await page.click('button[type="submit"]');

    await page.goto('/transactions');
    await page.click('text=+ New Transaction');
    await page.fill('[formControlName="name"]', 'Coffee');
    await page.fill('[formControlName="value"]', '5');
    await page.selectOption('[formControlName="operation"]', 'subtract');
    await page.selectOption('[formControlName="bankAccountId"]', { label: 'Active Account' });
    await page.click('button[type="submit"]');

    await page.goto('/dashboard');
    await expect(page.locator('text=Coffee')).toBeVisible();
  });

  test('shows link to bank account statement', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'Linked Account');
    await page.click('button[type="submit"]');

    await page.goto('/dashboard');
    const statementLink = page.locator('a', { hasText: 'Statement' }).or(page.locator('a', { hasText: 'View Statement' })).first();
    await expect(statementLink).toBeVisible();
    await statementLink.click();
    await expect(page).toHaveURL(/bank-accounts\/.+/);
  });
});
