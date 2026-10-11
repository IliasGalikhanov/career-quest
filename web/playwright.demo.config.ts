import { defineConfig } from "@playwright/test";
export default defineConfig({
  testDir: "./demo",
  workers: 1,
  timeout: 150_000,
  reporter: "list",
  webServer: {
    command: "go run ./cmd/server",
    cwd: "..",
    url: "http://127.0.0.1:8082/api/health",
    env: {
      CQ_ADDR: "127.0.0.1:8082",
      CQ_DATABASE: "data/video-demo.db",
      CQ_WEB_DIR: "web/dist",
    },
    reuseExistingServer: false,
    timeout: 180_000,
  },
});
