import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import { PNG } from 'pngjs';

// Compare two PNG buffers pixel by pixel
function compareImages(img1Buffer: Buffer, img2Buffer: Buffer): { match: boolean; diffPixels: number; totalPixels: number } {
  const png1 = PNG.sync.read(img1Buffer);
  const png2 = PNG.sync.read(img2Buffer);

  if (png1.width !== png2.width || png1.height !== png2.height) {
    return { match: false, diffPixels: -1, totalPixels: png1.width * png1.height };
  }

  let diffPixels = 0;
  const totalPixels = png1.width * png1.height;

  for (let i = 0; i < png1.data.length; i += 4) {
    // Compare RGBA values
    if (png1.data[i] !== png2.data[i] ||
        png1.data[i + 1] !== png2.data[i + 1] ||
        png1.data[i + 2] !== png2.data[i + 2] ||
        png1.data[i + 3] !== png2.data[i + 3]) {
      diffPixels++;
    }
  }

  return { match: diffPixels === 0, diffPixels, totalPixels };
}

// Test cases with binary graphics
const testCases = [
  {
    name: 'USPS Priority Mail Domestic',
    zplPath: 'testdata/visual/usps_domestic/label.zpl',
    width: 812,
    height: 1218,
    dpi: 203,
  },
  {
    name: 'USPS APO Military',
    zplPath: 'testdata/visual/usps_apo/label.zpl',
    width: 812,
    height: 1218,
    dpi: 203,
  },
  {
    name: 'USPS Priority Mail International',
    zplPath: 'testdata/visual/usps_intl/label.zpl',
    width: 812,
    height: 1218,
    dpi: 203,
  },
];

test.describe('WASM vs CLI Rendering', () => {
  for (const tc of testCases) {
    test(`${tc.name} - WASM matches CLI output`, async ({ page }) => {
      // Skip if ZPL file doesn't exist
      const zplFullPath = path.join(process.cwd(), tc.zplPath);
      if (!fs.existsSync(zplFullPath)) {
        test.skip();
        return;
      }

      // 1. Render with CLI
      const cliOutputPath = `/tmp/cli-${tc.name.replace(/\s+/g, '-')}.png`;
      execSync(
        `go run ./cmd/zplrender -o "${cliOutputPath}" -width ${tc.width} -height ${tc.height} -dpi ${tc.dpi} "${zplFullPath}"`,
        { cwd: process.cwd() }
      );
      const cliImage = fs.readFileSync(cliOutputPath);

      // 2. Load the page and wait for Monaco editor and WASM to be ready
      await page.goto('/go-zpl/');
      await page.waitForSelector('.monaco-editor', { timeout: 10000 });
      // Wait for WASM to render (render time shows up)
      await page.waitForSelector('#render-time:not(:empty)', { timeout: 10000 });

      // 3. Load the ZPL content into the editor using Monaco's global API
      const zplContent = fs.readFileSync(zplFullPath);
      const zplBase64 = zplContent.toString('base64');

      // Set the ZPL in Monaco editor via the model
      await page.evaluate((b64) => {
        const decoded = atob(b64);
        // Get Monaco's editor instance via the model
        const models = (window as any).monaco.editor.getModels();
        if (models.length > 0) {
          models[0].setValue(decoded);
        }
      }, zplBase64);

      // Set dimensions
      await page.fill('#width', String(tc.width / tc.dpi));
      await page.fill('#height', String(tc.height / tc.dpi));
      await page.selectOption('#unit', 'in');
      await page.selectOption('#dpi', String(tc.dpi));

      // 4. Wait for render and get the image
      await page.waitForTimeout(1000); // Wait for debounced render

      // Get the first preview image as base64
      const wasmImageBase64 = await page.evaluate(() => {
        const img = document.querySelector('.preview-image') as HTMLImageElement;
        if (!img || !img.src.startsWith('data:image/png;base64,')) {
          return null;
        }
        return img.src.replace('data:image/png;base64,', '');
      });

      expect(wasmImageBase64).not.toBeNull();
      const wasmImage = Buffer.from(wasmImageBase64!, 'base64');

      // 5. Compare images
      const result = compareImages(cliImage, wasmImage);

      // Allow tiny differences (< 0.01%) for anti-aliasing, etc.
      const diffPercent = (result.diffPixels / result.totalPixels) * 100;
      expect(diffPercent).toBeLessThan(0.01);

      // Cleanup
      fs.unlinkSync(cliOutputPath);
    });
  }
});
