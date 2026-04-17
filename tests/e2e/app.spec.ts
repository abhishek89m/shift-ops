import { expect, test } from '@playwright/test';

async function resetSeed(request: import('@playwright/test').APIRequestContext) {
  const response = await request.post('http://127.0.0.1:8080/v1/dev/reset');
  expect(response.ok()).toBeTruthy();
}

test.beforeEach(async ({ request }) => {
  await resetSeed(request);
});

test('recommended task card is visible and active work also stays visible in the in-progress list', async ({ page }) => {
  await page.goto('/');

  await expect(page.locator('.recommended-card')).toContainText('Finish quality check already started');
  await expect(page.locator('.task-row', { hasText: 'Finish quality check already started' })).toHaveCount(1);
});

test('terminal task disables invalid actions', async ({ page, request }) => {
  const response = await request.patch('http://127.0.0.1:8080/v1/tasks/task_quality_01', {
    data: {
      status: 'completed',
      completed_by: 'playwright',
      resolution_code: 'checked_ok'
    }
  });
  expect(response.ok()).toBeTruthy();

  await page.goto('/');
  await page.getByRole('button', { name: /Finish quality check already started/i }).click();

  await expect(page.getByRole('button', { name: 'Start' })).toBeDisabled();
  await expect(page.getByRole('button', { name: 'Complete' })).toBeDisabled();
  await expect(page.getByRole('button', { name: 'Skip' })).toBeDisabled();
});

test('failed task switch keeps current task active', async ({ page, request }) => {
  let nextTaskPatchCount = 0;

  await page.route('**/v1/tasks/task_parking_01', async (route) => {
    if (route.request().method() === 'PATCH') {
      nextTaskPatchCount += 1;
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'forced switch failure' })
      });
      return;
    }

    await route.continue();
  });

  await page.goto('/');
  await page.getByRole('button', { name: /Repark scooter blocking cycle lane/i }).click();
  await page.getByRole('button', { name: 'Switch task' }).click();

  await expect(page.getByText('forced switch failure')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Finish quality check already started' })).toBeVisible();

  const tasksResponse = await request.get('http://127.0.0.1:8080/v1/tasks');
  expect(tasksResponse.ok()).toBeTruthy();
  const payload = (await tasksResponse.json()) as { tasks: Array<{ id: string; status: string }> };

  expect(payload.tasks.find((task) => task.id === 'task_quality_01')?.status).toBe('in_progress');
  expect(payload.tasks.find((task) => task.id === 'task_parking_01')?.status).toBe('pending');
  await expect.poll(() => nextTaskPatchCount).toBe(1);
});

test('mobile checklist updates do not jump the page back to the top', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Open task' }).click();

  const checkbox = page.getByLabel('Inspect stem and deck');
  await checkbox.scrollIntoViewIfNeeded();
  await page.evaluate(() => window.scrollBy(0, 220));

  const beforeScroll = await page.evaluate(() => window.scrollY);
  await checkbox.check();

  await expect.poll(() => page.evaluate(() => window.scrollY)).toBeGreaterThan(beforeScroll - 120);
  await expect(checkbox).toBeChecked();
});

test('complete only activates after all checklist steps are checked, then returns mobile back to the list', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Open task' }).click();

  const completeButton = page.getByRole('button', { name: 'Complete' });
  const firstStep = page.getByLabel('Check brakes and bell');
  const secondStep = page.getByLabel('Inspect stem and deck');
  const thirdStep = page.getByLabel('Confirm scooter is ride-ready');

  await expect(completeButton).toBeDisabled();
  await expect(firstStep).toBeEnabled();
  await expect(secondStep).toBeEnabled();
  await expect(thirdStep).toBeEnabled();

  await firstStep.check();
  await expect(completeButton).toBeDisabled();
  await expect(secondStep).toBeEnabled();
  await expect(thirdStep).toBeEnabled();

  await secondStep.check();
  await expect(completeButton).toBeDisabled();
  await expect(thirdStep).toBeEnabled();

  await thirdStep.check();
  await expect(completeButton).toBeEnabled();

  await completeButton.click();

  await expect(page.getByRole('button', { name: 'Open task' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Finish quality check already started' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: /Finish quality check already started/i })).toBeVisible();

  await page.getByRole('button', { name: /Finish quality check already started/i }).click();
  await expect(page.getByLabel('Inspect stem and deck')).toBeDisabled();
});

test('remaining tasks are view-only until they are started', async ({ page, request }) => {
  const response = await request.patch('http://127.0.0.1:8080/v1/tasks/task_quality_01', {
    data: {
      status: 'completed',
      completed_by: 'playwright',
      resolution_code: 'checked_ok'
    }
  });
  expect(response.ok()).toBeTruthy();

  await page.goto('/');
  await expect(page.locator('.recommended-card')).toContainText('Repark scooter blocking cycle lane');
  await expect(page.locator('.task-row', { hasText: 'Repark scooter blocking cycle lane' })).toHaveCount(1);
  await page.getByRole('button', { name: 'Open task' }).click();

  await expect(page.getByRole('button', { name: 'Start' })).toBeEnabled();
  await expect(page.getByRole('button', { name: 'Complete' })).toBeDisabled();
  await expect(page.getByRole('button', { name: 'Skip' })).toBeDisabled();
  await expect(page.getByLabel('Stand scooter inside the parking area')).toBeDisabled();
  await expect(page.getByText('Confirm switch')).toHaveCount(0);
});
