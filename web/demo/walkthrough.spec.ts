import { test, expect } from "@playwright/test";
import path from "node:path";

// Real UI recording with reading pauses. Not part of the fast CI suite.
test("record the portfolio walkthrough", async ({ browser }) => {
  const context = await browser.newContext({
    viewport: { width: 1440, height: 1080 },
    recordVideo: {
      dir: "test-results/demo-video",
      size: { width: 1440, height: 1080 },
    },
  });
  const page = await context.newPage();
  const video = page.video()!;
  await page.addInitScript(() => localStorage.setItem("cq-language", "en"));
  const caption = async (text: string, pause = 6500) => {
    await page.evaluate((text) => {
      let element = document.getElementById("demo-caption");
      if (!element) {
        element = document.createElement("div");
        element.id = "demo-caption";
        element.style.cssText =
          "position:fixed;bottom:20px;left:260px;right:30px;z-index:10000;padding:18px 24px;border-radius:12px;background:#123c30;color:white;font:600 20px system-ui;box-shadow:0 4px 30px #0003;pointer-events:none";
        document.body.append(element);
      }
      element.textContent = text;
    }, text);
    await page.waitForTimeout(pause);
  };
  await page.goto("http://127.0.0.1:8082");
  await page
    .getByRole("button", { name: "Data & settings", exact: true })
    .click();
  await page.getByRole("button", { name: "Create another demo" }).click();
  await expect(page.getByTestId("level-architecture")).toHaveText("2");
  await caption(
    "Career Quest · Go + React + SQLite · Independent synthetic demo",
  );
  await caption(
    "Choose a career goal. Assessed skills stay separate from estimated progress.",
  );
  const card = page.getByTestId("activity-architecture-lab");
  await card.getByText("Exercise & review", { exact: true }).click();
  await card.scrollIntoViewIfNeeded();
  await caption(
    "Each activity has an original exercise, a deliverable and self-check criteria.",
  );
  await card.getByRole("button", { name: "Start activity" }).click();
  const note = page.getByRole("textbox", { name: "Work note (optional)" });
  await note.scrollIntoViewIfNeeded();
  await note.pressSequentially(
    "Designed an order service. Compared synchronous payment with a queue; documented retries and duplicate handling.",
    { delay: 28 },
  );
  await caption(
    "Record what you produced and learned. Completion is self-reported.",
    5000,
  );
  await page.getByRole("button", { name: "Mark complete" }).click();
  await expect(page.getByTestId("level-architecture")).toHaveText("3");
  await page.getByTestId("level-architecture").scrollIntoViewIfNeeded();
  await caption(
    "Architecture: estimated 2 → 3. The original assessment remains unchanged.",
  );
  await page.reload();
  await expect(page.getByTestId("level-architecture")).toHaveText("3");
  await caption(
    "Reload the page: your goal, participation and work note remain saved.",
    5000,
  );
  await page
    .getByRole("button", { name: "Activity library", exact: true })
    .click();
  await page.getByRole("searchbox").fill("distributed");
  await caption(
    "Unavailable activities explain the reason, including missing prerequisites.",
  );
  await page
    .getByRole("button", { name: "People insights", exact: true })
    .click();
  await page
    .getByRole("combobox", { name: "Department", exact: true })
    .selectOption("Platform");
  await caption(
    "HR view shows skill gaps and participation, without ranking employees.",
  );
  await page
    .getByRole("button", { name: "Data & settings", exact: true })
    .click();
  await caption(
    "Preview imports, export backups, rename workspaces and tune constraints.",
  );
  await page
    .getByRole("button", { name: "My trajectory", exact: true })
    .click();
  await page.getByRole("button", { name: "Switch theme" }).click();
  await caption(
    "Three languages · two themes · local data · reproducible tests and Docker.",
    6000,
  );
  await context.close();
  await video.saveAs(path.resolve("../docs/career-quest-demo.webm"));
});
