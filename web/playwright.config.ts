import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  workers: 1,
  timeout: 45_000,
  expect: { timeout: 10_000 },
  reporter: [["list"], ["html", { open: "never" }]],
  use: {
    baseURL: "http://127.0.0.1:8081",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1440, height: 1080 },
      },
    },
  ],
  webServer: {
    command: "go run ./cmd/server",
    cwd: "..",
    url: "http://127.0.0.1:8081/api/health",
    env: {
      CQ_ADDR: "127.0.0.1:8081",
      CQ_DATABASE: "data/browser-tests.db",
      CQ_WEB_DIR: "web/dist",
    },
    reuseExistingServer: false,
    timeout: 180_000,
  },
});
