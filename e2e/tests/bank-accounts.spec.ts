import { test, expect } from '@playwright/test';
import { register } from './helpers';

const timestamp = Date.now();
const user = {
  name: 'Bank Account Test User',
  username: `bank_user_${timestamp}`,
  email: `bank_${timestamp}@test.com`,
  password: 'bank_pass_123',
};

test.describe('Bank Accounts', () => {
  test.beforeEach(async ({ page }) => {
    await register(page, user);
  });

  test('creates a bank account with a name', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');

    await page.fill('[formControlName="name"]', 'Checking Account');
    await page.click('button[type="submit"]');

    await expect(page.locator('text=Checking Account')).toBeVisible();
  });

  test('creates a bank account with initial balance', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');

    await page.fill('[formControlName="name"]', 'Savings');
    await page.fill('[formControlName="initialBalance"]', '1500.50');
    await page.click('button[type="submit"]');

    await expect(page.locator('text=Savings')).toBeVisible();
    await expect(page.locator('text=1,500.50').or(page.locator('text=1500.50'))).toBeVisible();
  });

  test('rejects negative initial balance', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');

    await page.fill('[formControlName="name"]', 'Bad Account');
    await page.fill('[formControlName="initialBalance"]', '-100');
    await page.click('button[type="submit"]');

    await expect(page.locator('.error-msg, [class*="error"]')).toBeVisible();
  });

  test('edits a bank account name', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'Old Name');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Old Name')).toBeVisible();

    await page.click('text=Edit');
    await page.locator('[formControlName="name"]').last().fill('New Name');
    await page.click('text=Save');

    await expect(page.locator('text=New Name')).toBeVisible();
  });

  test('deletes a bank account', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'To Delete');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=To Delete')).toBeVisible();

    page.on('dialog', (dialog) => dialog.accept());
    await page.click('text=Delete');

    await expect(page.locator('text=To Delete')).not.toBeVisible();
  });

  test('navigates to bank account statement', async ({ page }) => {
    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'Statement Account');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Statement Account')).toBeVisible();

    await page.click('text=Statement');

    await expect(page).toHaveURL(/bank-accounts\/.+/);
    await expect(page.locator('text=Statement Account')).toBeVisible();
  });
});
