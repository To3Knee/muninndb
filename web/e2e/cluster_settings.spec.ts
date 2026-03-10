import { test, expect } from './fixtures/auth.js'

test.describe('Cluster Settings: Persistence (#175 regression guard)', () => {
  test('all four cluster settings fields persist after save via API', async ({ page }) => {
    // Verify the settings API returns the expected defaults first.
    const defaults = await page.evaluate(async () => {
      const res = await fetch('/api/admin/cluster/settings')
      return res.json()
    })
    expect(typeof defaults.heartbeat_ms).toBe('number')
    expect(typeof defaults.sdown_beats).toBe('number')
    expect(typeof defaults.ccs_interval_seconds).toBe('number')
    expect(typeof defaults.reconcile_on_heal).toBe('boolean')

    // Save new values for all four fields via the API directly.
    const newSettings = {
      heartbeat_ms: 750,
      sdown_beats: 5,
      ccs_interval_seconds: 60,
      reconcile_on_heal: false,
    }
    const saveResp = await page.evaluate(async (settings) => {
      const res = await fetch('/api/admin/cluster/settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify(settings),
      })
      return { status: res.status, body: await res.json() }
    }, newSettings)
    expect(saveResp.status).toBe(200)
    expect(saveResp.body.saved).toBe(true)

    // Read back from the GET endpoint — this is the core #175 regression guard.
    // Before the fix, only heartbeat_ms was persisted; the other three were silently dropped.
    const saved = await page.evaluate(async () => {
      const res = await fetch('/api/admin/cluster/settings')
      return res.json()
    })
    expect(saved.heartbeat_ms).toBe(750)
    expect(saved.sdown_beats).toBe(5)          // failed before #175 fix
    expect(saved.ccs_interval_seconds).toBe(60) // failed before #175 fix
    expect(saved.reconcile_on_heal).toBe(false) // failed before #175 fix
  })

  test('cluster settings form fields are present and interact', async ({ page }) => {
    // Navigate to Settings → Cluster tab (if visible).
    // The cluster tab is only shown when cluster mode is enabled, so this test
    // just checks that the form elements exist when the cluster section is shown.
    // We navigate directly to verify the testids are wired correctly.
    await page.goto('/')
    await page.locator('.sidebar-item').filter({ hasText: 'Settings' }).click()

    // If cluster is disabled the form is hidden — skip UI interaction in that case.
    const clusterSection = page.locator('[data-testid="btn-save-cluster-settings"]')
    const isVisible = await clusterSection.isVisible()
    if (!isVisible) {
      test.skip()
      return
    }

    // When visible, verify all four inputs have the expected testids.
    await expect(page.getByTestId('input-cluster-heartbeat-ms')).toBeVisible()
    await expect(page.getByTestId('input-cluster-sdown-beats')).toBeVisible()
    await expect(page.getByTestId('input-cluster-ccs-interval')).toBeVisible()
    await expect(page.getByTestId('toggle-cluster-reconcile-heal')).toBeVisible()
  })
})
