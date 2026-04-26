import { test, expect } from '@playwright/test';
import { register } from './helpers';

const timestamp = Date.now();
const user = {
  name: 'Statement Test User',
  username: `stmt_user_${timestamp}`,
  email: `stmt_${timestamp}@test.com`,
  password: 'stmt_pass_123',
};

test.describe('Bank Account Statement', () => {
  test.beforeEach(async ({ page }) => {
    await register(page, user);

    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'Statement Account');
    await page.fill('[formControlName="initialBalance"]', '500');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Statement Account')).toBeVisible();
  });

  test('shows account balance on statement page', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=Statement');
    await expect(page).toHaveURL(/bank-accounts\/.+/);
    await expect(page.locator('text=Statement Account')).toBeVisible();
  });

  test('shows transactions in paginated list', async ({ page }) => {
    await page.goto('/transactions');
    await page.click('text=+ New Transaction');
    await page.fill('[formControlName="name"]', 'Rent');
    await page.fill('[formControlName="value"]', '800');
    await page.selectOption('[formControlName="operation"]', 'subtract');
    await page.selectOption('[formControlName="bankAccountId"]', { label: 'Statement Account' });
    await page.click('button[type="submit"]');

    await page.goto('/bank-accounts');
    await page.click('text=Statement');
    await expect(page.locator('text=Rent')).toBeVisible();
  });

  test('shows upcoming items section', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=Statement');
    await expect(page.locator('text=Upcoming').or(page.locator('[class*="upcoming"]'))).toBeVisible();
  });
});
