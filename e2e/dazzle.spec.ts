import { test, expect, type Page } from '@playwright/test';

const DAZZLE_URL = 'http://localhost:29100/**';

// The real Dazzle server answers with permissive CORS; the print POST is a CORS request.
const dazzleHeaders = {
  'Access-Control-Allow-Origin': '*',
};

async function waitForEditor(page: Page) {
  await page.waitForSelector('.monaco-editor', { timeout: 10000 });
}

async function setEditorValue(page: Page, text: string) {
  await page.evaluate((value) => {
    const models = (window as any).monaco.editor.getModels();
    if (models.length > 0) {
      models[0].setValue(value);
    }
  }, text);
}

async function optIn(page: Page) {
  await page.locator('#dazzle-connect').click();
  await expect(page.locator('#dazzle-print-btn')).toBeVisible();
}

// Each test gets a fresh browser context, so localStorage starts empty (not opted in).
test.describe('Dazzle print server', () => {
  test('No localhost request before opt-in', async ({ page }) => {
    let dazzleRequests = 0;
    await page.route(DAZZLE_URL, async (route) => {
      dazzleRequests++;
      await route.abort();
    });

    await page.goto('/go-zpl/');
    await page.waitForTimeout(1500);

    expect(dazzleRequests).toBe(0);
    await expect(page.locator('#dazzle-notice')).toBeVisible();
    await expect(page.locator('#dazzle-connect')).toBeVisible();
    await expect(page.locator('#dazzle-print-btn')).toBeHidden();
  });

  test('Opt-in is remembered and status controls visibility', async ({ page }) => {
    let statusOk = true;
    // The first probe is held so the connecting state is observable.
    let releaseStatus = () => {};
    let statusHeld: Promise<void> | null = new Promise<void>((resolve) => {
      releaseStatus = resolve;
    });
    await page.route(DAZZLE_URL, async (route) => {
      if (statusHeld) {
        await statusHeld;
        statusHeld = null;
      }
      if (!statusOk) {
        await route.abort();
        return;
      }
      await route.fulfill({
        status: 200,
        headers: dazzleHeaders,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'ok', version: 'test' }),
      });
    });

    await page.goto('/go-zpl/');
    await expect(page.locator('#dazzle-notice')).toBeVisible();

    let statusRequests = 0;
    page.on('request', (req) => {
      if (req.url().startsWith('http://localhost:29100/')) {
        statusRequests++;
      }
    });

    await page.locator('#dazzle-connect').click();
    await expect(page.locator('#dazzle-connecting')).toBeVisible();
    await expect(page.locator('#dazzle-download')).toBeHidden();
    await expect(page.locator('#dazzle-print-btn')).toBeHidden();
    releaseStatus();
    await expect(page.locator('#dazzle-print-btn')).toBeVisible();
    await expect(page.locator('#dazzle-connecting')).toBeHidden();
    expect(statusRequests).toBeGreaterThanOrEqual(1);
    await expect(page.locator('#dazzle-notice')).toBeHidden();

    await page.reload();
    await expect(page.locator('#dazzle-print-btn')).toBeVisible();
    await expect(page.locator('#dazzle-notice')).toBeHidden();

    statusOk = false;
    await expect(page.locator('#dazzle-print-btn')).toBeHidden({ timeout: 5000 });
    await expect(page.locator('#dazzle-notice')).toBeVisible();
    await expect(page.locator('#dazzle-connect')).toBeHidden();
  });

  test('Print sends exact bytes — Latin-1 binary', async ({ page }) => {
    const bytes = Buffer.from([0x5e, 0x58, 0x41, 0x80, 0xff, 0x00, 0x41, 0x5e, 0x58, 0x5a]);
    const zpl = bytes.toString('latin1');
    let printUrl = '';
    let printBody: string | null = null;

    await page.route(DAZZLE_URL, async (route) => {
      const req = route.request();
      if (req.method() === 'POST' && req.url().includes('/print')) {
        printUrl = req.url();
        printBody = req.postData();
        await route.fulfill({
          status: 200,
          headers: dazzleHeaders,
          contentType: 'application/json',
          body: JSON.stringify({ job_id: 'test-job' }),
        });
        return;
      }
      await route.fulfill({
        status: 200,
        headers: dazzleHeaders,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'ok', version: 'test' }),
      });
    });

    await page.goto('/go-zpl/');
    await waitForEditor(page);
    await optIn(page);
    await setEditorValue(page, zpl);
    await page.locator('#dazzle-print-btn').click();

    const toast = page.locator('#toasts .toast');
    await expect(toast).toHaveText('Sent to printer (job test-job)');
    await expect(toast).not.toHaveClass(/toast-error/);

    expect(printUrl).toContain('encoding=base64');
    expect(printBody).not.toBeNull();
    expect(Buffer.from(printBody!, 'base64').equals(bytes)).toBe(true);
  });

  test('Print sends exact bytes — dropped file with C1 bytes', async ({ page }) => {
    // 0x80–0x9F are exactly what a windows-1252 decode would remap.
    const bytes = Buffer.from([0x5e, 0x58, 0x41, 0x41, 0x80, 0x81, 0x9f, 0xff, 0x5e, 0x58, 0x5a]);
    let printBody: string | null = null;

    await page.route(DAZZLE_URL, async (route) => {
      const req = route.request();
      if (req.method() === 'POST' && req.url().includes('/print')) {
        printBody = req.postData();
        await route.fulfill({
          status: 200,
          headers: dazzleHeaders,
          contentType: 'application/json',
          body: JSON.stringify({ job_id: 'test-job' }),
        });
        return;
      }
      await route.fulfill({
        status: 200,
        headers: dazzleHeaders,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'ok', version: 'test' }),
      });
    });

    await page.goto('/go-zpl/');
    await waitForEditor(page);
    await optIn(page);

    const dataTransfer = await page.evaluateHandle((b64) => {
      const raw = atob(b64);
      const data = Uint8Array.from(raw, (c) => c.charCodeAt(0));
      const dt = new DataTransfer();
      dt.items.add(new File([data], 'label.zpl', { type: 'application/octet-stream' }));
      return dt;
    }, bytes.toString('base64'));
    await page.locator('#drop-zone').dispatchEvent('drop', { dataTransfer });
    await expect
      .poll(() => page.evaluate(() => (window as any).monaco.editor.getModels()[0].getValue().length))
      .toBe(bytes.length);

    await page.locator('#dazzle-print-btn').click();
    await expect(page.locator('#toasts .toast')).toHaveText('Sent to printer (job test-job)');
    expect(printBody).not.toBeNull();
    expect(Buffer.from(printBody!, 'base64').equals(bytes)).toBe(true);
  });

  test('Print sends exact bytes — Unicode', async ({ page }) => {
    const zpl = '^XA^CI28^FDcafé ✓^FS^XZ';
    let printBody: string | null = null;

    await page.route(DAZZLE_URL, async (route) => {
      const req = route.request();
      if (req.method() === 'POST' && req.url().includes('/print')) {
        printBody = req.postData();
        await route.fulfill({
          status: 200,
          headers: dazzleHeaders,
          contentType: 'application/json',
          body: JSON.stringify({ job_id: 'test-job' }),
        });
        return;
      }
      await route.fulfill({
        status: 200,
        headers: dazzleHeaders,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'ok', version: 'test' }),
      });
    });

    await page.goto('/go-zpl/');
    await waitForEditor(page);
    await optIn(page);
    await setEditorValue(page, zpl);
    await page.locator('#dazzle-print-btn').click();

    await expect(page.locator('#toasts .toast')).toHaveText('Sent to printer (job test-job)');
    expect(printBody).not.toBeNull();
    expect(Buffer.from(printBody!, 'base64').equals(Buffer.from(zpl, 'utf8'))).toBe(true);
  });

  test('Connect before Monaco loads keeps the saved label', async ({ page }) => {
    const savedZpl = '^XA^FDKEEP-ME^FS^XZ';
    await page.addInitScript((zpl) => {
      localStorage.setItem('zpl-renderer', JSON.stringify({ zplBase64: btoa(zpl) }));
    }, savedZpl);
    // Hold Monaco back so Connect is clicked while `editor` is still undefined.
    let releaseMonaco = () => {};
    const monacoHeld = new Promise<void>((resolve) => {
      releaseMonaco = resolve;
    });
    await page.route('**/vs/editor/editor.main.js', async (route) => {
      await monacoHeld;
      await route.continue();
    });
    await page.route(DAZZLE_URL, async (route) => {
      await route.fulfill({
        status: 200,
        headers: dazzleHeaders,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'ok', version: 'test' }),
      });
    });

    // `load` would wait for the held editor bundle, so don't wait for it.
    await page.goto('/go-zpl/', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('#dazzle-connect')).toBeVisible();
    await expect(page.locator('.monaco-editor')).toHaveCount(0);
    await page.locator('#dazzle-connect').click();
    await expect(page.locator('#dazzle-print-btn')).toBeVisible();

    const stored = await page.evaluate(() => JSON.parse(localStorage.getItem('zpl-renderer') || '{}'));
    expect(atob(stored.zplBase64)).toBe(savedZpl);
    expect(await page.evaluate(() => localStorage.getItem('zpl-renderer-dazzle'))).toBe('1');

    releaseMonaco();
    await waitForEditor(page);
  });

  test('Print failure toast, with the button held while sending', async ({ page }) => {
    let releasePrint = () => {};
    await page.route(DAZZLE_URL, async (route) => {
      const req = route.request();
      if (req.method() === 'POST' && req.url().includes('/print')) {
        await new Promise<void>((resolve) => {
          releasePrint = resolve;
        });
        await route.fulfill({
          status: 500,
          headers: dazzleHeaders,
          body: 'printer offline',
        });
        return;
      }
      await route.fulfill({
        status: 200,
        headers: dazzleHeaders,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'ok', version: 'test' }),
      });
    });

    await page.goto('/go-zpl/');
    await waitForEditor(page);
    await optIn(page);

    await page.locator('#dazzle-print-btn').click();
    await expect(page.locator('#dazzle-print-btn')).toBeDisabled();
    await expect(page.locator('#toasts .toast')).toHaveCount(0);

    releasePrint();
    const toast = page.getByRole('status').locator('.toast');
    await expect(toast).toHaveClass(/toast-error/);
    await expect(toast).toContainText('printer offline');
    await expect(page.locator('#dazzle-print-btn')).toBeEnabled();
  });
});
