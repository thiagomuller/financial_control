import { test, expect } from '@playwright/test';
import { register } from './helpers';

const timestamp = Date.now();
const user = {
  name: 'Tags Test User',
  username: `tags_user_${timestamp}`,
  email: `tags_${timestamp}@test.com`,
  password: 'tags_pass_123',
};

test.describe('Tags', () => {
  test.beforeEach(async ({ page }) => {
    await register(page, user);
  });

  test('creates a tag', async ({ page }) => {
    await page.goto('/tags');
    await page.click('text=+ New Tag');

    await page.fill('[formControlName="name"]', 'Food');
    await page.click('button[type="submit"]');

    await expect(page.locator('text=Food')).toBeVisible();
  });

  test('edits a tag', async ({ page }) => {
    await page.goto('/tags');
    await page.click('text=+ New Tag');
    await page.fill('[formControlName="name"]', 'Old Tag');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Old Tag')).toBeVisible();

    await page.click('text=Edit');
    await page.locator('[formControlName="name"]').last().fill('New Tag');
    await page.click('text=Save');

    await expect(page.locator('text=New Tag')).toBeVisible();
  });

  test('deletes a tag', async ({ page }) => {
    await page.goto('/tags');
    await page.click('text=+ New Tag');
    await page.fill('[formControlName="name"]', 'Delete Tag');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Delete Tag')).toBeVisible();

    page.on('dialog', (dialog) => dialog.accept());
    await page.locator('[class*="tag"], tr', { hasText: 'Delete Tag' }).locator('text=Delete').click();

    await expect(page.locator('text=Delete Tag')).not.toBeVisible();
  });
});
