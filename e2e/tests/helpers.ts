import { Page } from '@playwright/test';

export const TEST_USER = {
  name: 'E2E Test User',
  username: `e2e_user_${Date.now()}`,
  email: `e2e_${Date.now()}@test.com`,
  password: 'e2e_password_123',
};

export async function register(page: Page, user = TEST_USER): Promise<void> {
  await page.goto('/register');
  await page.fill('[formControlName="name"]', user.name);
  await page.fill('[formControlName="username"]', user.username);
  await page.fill('[formControlName="email"]', user.email);
  await page.fill('[formControlName="password"]', user.password);
  await page.click('button[type="submit"]');
  await page.waitForURL('**/dashboard');
}

export async function login(page: Page, user = TEST_USER): Promise<void> {
  await page.goto('/login');
  await page.fill('[formControlName="username"]', user.username);
  await page.fill('[formControlName="password"]', user.password);
  await page.click('button[type="submit"]');
  await page.waitForURL('**/dashboard');
}
