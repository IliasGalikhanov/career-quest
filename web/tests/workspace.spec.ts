import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'

test.beforeEach(async ({ page }) => { await page.addInitScript(() => { localStorage.setItem('cq-language', 'en') }) })

async function openDemo(page: Page) {
  await page.goto('/')
  await page.getByRole('button', { name: 'Data & settings', exact: true }).click()
  await page.getByRole('button', { name: 'Create another demo' }).click()
  await expect(page.getByTestId('level-architecture')).toHaveText('2')
}

test('complete a recommendation and keep progress after reload', async ({ page }) => {
  const errors: string[] = []; page.on('pageerror', error => errors.push(error.message))
  await openDemo(page)
  const before = await page.getByTestId('coverage').textContent()
  await page.screenshot({ path: 'test-results/trajectory-desktop.png', fullPage: true })
  await page.getByTestId('activity-architecture-lab').getByRole('button', { name: 'Start activity' }).click()
  await page.getByRole('button', { name: 'Mark complete' }).click()
  await expect(page.getByTestId('level-architecture')).toHaveText('3')
  await expect(page.getByTestId('coverage')).not.toHaveText(before!)
  await page.reload()
  await expect(page.getByTestId('level-architecture')).toHaveText('3')
  await expect(page.getByRole('cell', { name: 'Completed', exact: true })).toBeVisible()
  expect(errors).toEqual([])
})

test('update goal and explain unavailable activities', async ({ page }) => {
  await openDemo(page)
  await page.getByRole('button', { name: 'Update goal' }).click()
  await page.getByRole('combobox', { name: 'Target grade', exact: true }).selectOption('Lead')
  await page.getByRole('button', { name: 'Save changes' }).click()
  await expect(page.locator('.career-route')).toContainText('Lead')
  await page.getByRole('button', { name: 'Activity library', exact: true }).click()
  await page.getByRole('searchbox').fill('distributed')
  await expect(page.getByTestId('activity-architecture-deep-dive')).toContainText('Prerequisites are not met')
  await expect(page.getByTestId('activity-architecture-deep-dive').getByRole('button')).toBeDisabled()
})

test('preview, import and export a synthetic bundle', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: 'Data & settings', exact: true }).click()
  await page.getByLabel('Select input files').setInputFiles(path.resolve('../examples/demo.json'))
  await page.getByLabel('These data are fictional').check()
  await page.getByRole('button', { name: 'Validate & preview' }).click()
  await expect(page.getByText('Ready to import')).toBeVisible()
  await expect(page.locator('.preview-stats')).toContainText('12 profiles')
  await page.getByRole('button', { name: 'Import validated data' }).click()
  await expect(page.getByRole('status')).toHaveText('Dataset imported')
  const downloadPromise = page.waitForEvent('download')
  await page.getByRole('link', { name: 'Export backup', exact: true }).click()
  const download = await downloadPromise
  expect(download.suggestedFilename()).toBe('career-quest-backup.json')
})

test('invalid import has a readable error and preserves the current dataset', async ({ page }) => {
  await openDemo(page)
  await page.getByRole('button', { name: 'Data & settings', exact: true }).click()
  await page.getByLabel('Select input files').setInputFiles({ name: 'broken.json', mimeType: 'application/json', buffer: Buffer.from('{"meta":') })
  await page.getByLabel('These data are fictional').check()
  await page.getByRole('button', { name: 'Validate & preview' }).click()
  await expect(page.getByRole('alert')).toContainText('broken.json')
  await expect(page.getByRole('button', { name: 'Import validated data' })).toHaveCount(0)
  await page.getByRole('button', { name: 'My trajectory', exact: true }).click()
  await expect(page.getByTestId('level-architecture')).toHaveText('2')
})

test('filter HR insights by department', async ({ page }) => {
  await openDemo(page)
  await page.getByRole('button', { name: 'People insights', exact: true }).click()
  await page.getByRole('combobox', { name: 'Department', exact: true }).selectOption('Platform')
  await expect(page.locator('.stat-strip').first()).toContainText('4')
  await expect(page.getByRole('heading', { name: 'People below requirement' })).toBeVisible()
})

test('mobile layout, languages and dark theme remain usable', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await openDemo(page)
  await page.getByRole('button', { name: 'Switch theme' }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')
  await page.getByLabel('Language', { exact: true }).selectOption('kk')
  await expect(page.getByRole('heading', { name: 'Сізге ұсынылады' })).toBeVisible()
  await page.getByLabel('Тіл', { exact: true }).selectOption('ru')
  await expect(page.getByRole('heading', { name: 'Рекомендуем вам' })).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await page.screenshot({ path: 'test-results/trajectory-mobile-dark.png', fullPage: true })
})
