/**
 * E2E tests for bank account statement — requires Playwright to run.
 * Install: npm install --save-dev @playwright/test
 * Run: npx playwright test e2e/tests/bank-account-statement.spec.ts
 */
import { test, expect } from '@playwright/test';

const BASE = process.env['APP_URL'] ?? 'http://localhost:4200';
const API = process.env['API_URL'] ?? 'http://localhost:8080';

async function registerAndLogin(request: any): Promise<string> {
  const username = `e2euser_${Date.now()}`;
  const reg = await request.post(`${API}/api/auth/register`, {
    data: { name: 'E2E User', username, email: `${username}@test.com`, password: 'pass123' },
  });
  const body = await reg.json();
  return body.token as string;
}

async function createAccount(request: any, token: string, name: string, balance: number): Promise<string> {
  const res = await request.post(`${API}/api/bank-accounts`, {
    headers: { Authorization: `Bearer ${token}` },
    data: { name, initial_balance: balance },
  });
  const body = await res.json();
  return body.id as string;
}

async function createTransfer(
  request: any,
  token: string,
  name: string,
  value: number,
  sourceId: string,
  targetId: string,
): Promise<void> {
  await request.post(`${API}/api/transfers`, {
    headers: { Authorization: `Bearer ${token}` },
    data: { name, value, source_account_id: sourceId, target_account_id: targetId },
  });
}

test.describe('Bank account statement includes transfers', () => {
  test('transfer appears in source account statement with subtract operation', async ({ page, request }) => {
    const token = await registerAndLogin(request);
    const srcId = await createAccount(request, token, 'Source', 1000);
    const tgtId = await createAccount(request, token, 'Target', 0);
    await createTransfer(request, token, 'Rent payment', 300, srcId, tgtId);

    await page.goto(`${BASE}/login`);
    await page.evaluate((t) => localStorage.setItem('token', t), token);
    await page.goto(`${BASE}/bank-accounts/${srcId}`);

    await expect(page.locator('.badge-transfer').first()).toBeVisible();
    await expect(page.locator('.badge-transfer').first()).toHaveText('Transfer');
  });

  test('transfer amount shows as subtract on source account statement', async ({ page, request }) => {
    const token = await registerAndLogin(request);
    const srcId = await createAccount(request, token, 'Source2', 500);
    const tgtId = await createAccount(request, token, 'Target2', 0);
    await createTransfer(request, token, 'Savings', 150, srcId, tgtId);

    await page.evaluate((t) => localStorage.setItem('token', t), token);
    await page.goto(`${BASE}/bank-accounts/${srcId}`);

    const row = page.locator('tbody tr').first();
    await expect(row.locator('.amount-subtract')).toBeVisible();
  });

  test('transfer appears in target account statement with add operation', async ({ page, request }) => {
    const token = await registerAndLogin(request);
    const srcId = await createAccount(request, token, 'Source3', 600);
    const tgtId = await createAccount(request, token, 'Target3', 0);
    await createTransfer(request, token, 'Income transfer', 200, srcId, tgtId);

    await page.evaluate((t) => localStorage.setItem('token', t), token);
    await page.goto(`${BASE}/bank-accounts/${tgtId}`);

    const row = page.locator('tbody tr').first();
    await expect(row.locator('.amount-add')).toBeVisible();
    await expect(page.locator('.badge-transfer').first()).toBeVisible();
  });
});
