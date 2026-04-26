import { test, expect } from '@playwright/test';

const timestamp = Date.now();
const user = {
  name: 'Auth Test User',
  username: `auth_user_${timestamp}`,
  email: `auth_${timestamp}@test.com`,
  password: 'auth_pass_123',
};

test.describe('Authentication', () => {
  test('register creates an account and redirects to dashboard', async ({ page }) => {
    await page.goto('/register');

    await page.fill('[formControlName="name"]', user.name);
    await page.fill('[formControlName="username"]', user.username);
    await page.fill('[formControlName="email"]', user.email);
    await page.fill('[formControlName="password"]', user.password);
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL(/dashboard/);
  });

  test('login with valid credentials redirects to dashboard', async ({ page }) => {
    await page.goto('/login');

    await page.fill('[formControlName="username"]', user.username);
    await page.fill('[formControlName="password"]', user.password);
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL(/dashboard/);
  });

  test('login with invalid password shows error', async ({ page }) => {
    await page.goto('/login');

    await page.fill('[formControlName="username"]', user.username);
    await page.fill('[formControlName="password"]', 'wrong_password');
    await page.click('button[type="submit"]');

    await expect(page.locator('.error-msg, [class*="error"]')).toBeVisible();
    await expect(page).not.toHaveURL(/dashboard/);
  });

  test('logout clears session and redirects to login', async ({ page }) => {
    await page.goto('/login');
    await page.fill('[formControlName="username"]', user.username);
    await page.fill('[formControlName="password"]', user.password);
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/dashboard/);

    await page.click('text=Logout');
    await expect(page).toHaveURL(/login/);
  });

  test('accessing protected route without auth redirects to login', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/login/);
  });
});
