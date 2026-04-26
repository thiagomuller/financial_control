import { test, expect } from '@playwright/test';
import { register } from './helpers';

const timestamp = Date.now();
const user = {
  name: 'Txn Test User',
  username: `txn_user_${timestamp}`,
  email: `txn_${timestamp}@test.com`,
  password: 'txn_pass_123',
};

test.describe('Transactions', () => {
  test.beforeEach(async ({ page }) => {
    await register(page, user);

    await page.goto('/bank-accounts');
    await page.click('text=+ New Account');
    await page.fill('[formControlName="name"]', 'Txn Account');
    await page.fill('[formControlName="initialBalance"]', '500');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Txn Account')).toBeVisible();
  });

  test('creates an income transaction and increases balance', async ({ page }) => {
    await page.goto('/transactions');
    await page.click('text=+ New Transaction');

    await page.fill('[formControlName="name"]', 'Salary');
    await page.fill('[formControlName="value"]', '1000');
    await page.selectOption('[formControlName="operation"]', 'add');
    await page.selectOption('[formControlName="bankAccountId"]', { label: 'Txn Account' });
    await page.click('button[type="submit"]');

    await expect(page.locator('text=Salary')).toBeVisible();
  });

  test('creates an expense transaction', async ({ page }) => {
    await page.goto('/transactions');
    await page.click('text=+ New Transaction');

    await page.fill('[formControlName="name"]', 'Groceries');
    await page.fill('[formControlName="value"]', '50');
    await page.selectOption('[formControlName="operation"]', 'subtract');
    await page.selectOption('[formControlName="bankAccountId"]', { label: 'Txn Account' });
    await page.click('button[type="submit"]');

    await expect(page.locator('text=Groceries')).toBeVisible();
  });

  test('deletes a transaction', async ({ page }) => {
    await page.goto('/transactions');
    await page.click('text=+ New Transaction');
    await page.fill('[formControlName="name"]', 'Delete Me');
    await page.fill('[formControlName="value"]', '10');
    await page.selectOption('[formControlName="operation"]', 'add');
    await page.selectOption('[formControlName="bankAccountId"]', { label: 'Txn Account' });
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Delete Me')).toBeVisible();

    page.on('dialog', (dialog) => dialog.accept());
    await page.locator('tr', { hasText: 'Delete Me' }).locator('text=Delete').click();

    await expect(page.locator('text=Delete Me')).not.toBeVisible();
  });
});
