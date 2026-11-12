import { test, expect } from "@playwright/test";
const profile = {
  id: "u1",
  email: "player@example.test",
  nickname: "player",
  name: "Игрок",
  timezone: "Europe/Moscow",
  email_notifications: true,
  roles: ["participant", "organizer"],
  email_verified: true,
  blocked: false,
  version: 3,
};
const tokens = {
  access_token: "test-access",
  refresh_token: "test-refresh",
  token_type: "Bearer",
  expires_in: 900,
};
const discipline = {
  id: "d1",
  name: "Шахматы",
  min_roster_size: 1,
  max_roster_size: 1,
  allows_draw: true,
  score_unit: "очки",
  active: true,
  version: 1,
};
const tournament = {
  id: "t1",
  organizer: { id: "u1", nickname: "player" },
  name: "Кубок интеграции",
  discipline_id: "d1",
  format: "single_elimination",
  max_participants: 8,
  timezone: "Europe/Moscow",
  rules: { best_of: 1 },
  status: "registration",
  registration_opens_at: new Date(Date.now() - 3_600_000).toISOString(),
  registration_closes_at: new Date(Date.now() + 3_600_000).toISOString(),
  hidden: false,
  accepted_count: 0,
  version: 1,
  created_at: "2026-11-11T10:00:00Z",
};
const pageOf = (items: unknown[]) => ({
  items,
  total: items.length,
  limit: 100,
  offset: 0,
});
test.beforeEach(async ({ page }) => {
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname.replace("/api/v1", "");
    let body: unknown = null;
    if (path === "/disciplines") body = pageOf([discipline]);
    else if (path === "/tournaments") body = pageOf([tournament]);
    else if (path === "/users/me/tournaments") body = pageOf([]);
    else if (path === "/users/me") body = profile;
    else if (path === "/users/me/teams" || path === "/users/me/invitations")
      body = pageOf([]);
    else if (path === "/auth/login") body = tokens;
    else if (path === "/tournaments/t1") body = tournament;
    else if (path.endsWith("/bracket"))
      body = { stages: [], groups: [], matches: [] };
    else if (path.endsWith("/standings")) body = [];
    else if (path.endsWith("/applications") || path.endsWith("/participants"))
      body = pageOf([]);
    else if (path === "/ratings") body = pageOf([]);
    else {
      await route.fulfill({
        status: 404,
        json: { message: `Unexpected route ${path}` },
      });
      return;
    }
    await route.fulfill({ json: body });
  });
});
test("catalog uses API data, opens the selected tournament and shows empty states", async ({
  page,
}) => {
  await page.goto("/tournaments");
  await expect(
    page.getByRole("heading", { name: "Кубок интеграции" }),
  ).toBeVisible();
  await expect(
    page.getByText("Осенний кубок по мини-футболу 2026"),
  ).toHaveCount(0);
  await page.getByRole("button", { name: "Подать заявку" }).click();
  await expect(page).toHaveURL(/tournaments\/t1\/groups/);
  await expect(
    page.getByText("Таблицы появятся после жеребьёвки"),
  ).toBeVisible();
  await page.getByRole("link", { name: "Расписание", exact: true }).click();
  await expect(
    page.getByText("Матчи появятся после жеребьёвки."),
  ).toBeVisible();
});
test("login errors do not authenticate and a successful session survives reload", async ({
  page,
}) => {
  await page.route("**/api/v1/auth/login", (route) =>
    route.fulfill({ status: 401, json: { message: "Неверный пароль" } }),
  );
  await page.goto("/teams");
  await expect(page).toHaveURL(/auth\/login/);
  await page.locator("[name=email]").fill("player@example.test");
  await page.locator("[name=password]").fill("WrongPassword12");
  await page.getByRole("main").getByRole("button", { name: "Войти", exact: true }).click();
  await expect(page.getByRole("alert")).toContainText("Неверный пароль");
  await expect(page).toHaveURL(/auth\/login/);
  await page.unroute("**/api/v1/auth/login");
  await page.getByRole("main").getByRole("button", { name: "Войти", exact: true }).click();
  await expect(page).toHaveURL(/\/teams$/);
  await page.reload();
  await expect(
    page.getByRole("heading", { name: "Мои команды" }),
  ).toBeVisible();
});
test("registration sends the password and consent, then asks for email verification", async ({
  page,
}) => {
  await page.route("**/api/v1/auth/register", async (route) => {
    expect(route.request().postDataJSON()).toMatchObject({
      password: "Password_2026",
      personal_data_consent: true,
      nickname: "new_player",
    });
    await route.fulfill({ status: 202 });
  });
  await page.goto("/auth/register");
  await page.locator("[name=email]").fill("new@example.test");
  await page.locator("[name=nickname]").fill("new_player");
  await page.locator("[name=name]").fill("Новый игрок");
  await page.locator("[name=password]").fill("Password_2026");
  await page.locator("[name=repeat]").fill("Password_2026");
  await page.getByRole("checkbox").check();
  await page.getByRole("button", { name: "Зарегистрироваться" }).click();
  await expect(page.getByRole("status")).toContainText("Проверьте почту");
  await expect(
    page.getByRole("button", { name: "Войти", exact: true }),
  ).toBeVisible();
});
test("profile sends the optimistic version and only reports success after API success", async ({
  page,
}) => {
  await page.addInitScript(
    (tokens) =>
      localStorage.setItem("tourneyhub.tokens", JSON.stringify(tokens)),
    tokens,
  );
  let saved = false;
  await page.route("**/api/v1/users/me", async (route) => {
    if (route.request().method() === "PATCH") {
      expect(route.request().postDataJSON()).toMatchObject({
        version: 3,
        name: "Новое имя",
        email_notifications: false,
      });
      saved = true;
      await route.fulfill({
        json: {
          ...profile,
          name: "Новое имя",
          version: 4,
          email_notifications: false,
        },
      });
    } else await route.fulfill({ json: profile });
  });
  await page.goto("/profile");
  await page.locator("[name=name]").fill("Новое имя");
  await page.locator("[name=email]").uncheck();
  await page.getByRole("button", { name: "Сохранить", exact: true }).click();
  await expect(page.getByRole("status")).toHaveText("Изменения сохранены");
  expect(saved).toBe(true);
});
test("missing tournament shows 404 instead of mock data", async ({ page }) => {
  await page.goto("/tournaments/missing");
  await expect(page.getByRole("alert")).toBeVisible();
  await expect(
    page.getByText("Осенний кубок по мини-футболу 2026"),
  ).toHaveCount(0);
});
test("verification and reset links submit tokens to the correct endpoints", async ({
  page,
}) => {
  await page.route("**/api/v1/auth/verify-email", async (route) => {
    expect(route.request().postDataJSON()).toEqual({ token: "verify-token" });
    await route.fulfill({ status: 204 });
  });
  await page.goto("/auth/verify-email?token=verify-token");
  await page.getByRole("button", { name: "Подтвердить email" }).click();
  await expect(page.getByRole("status")).toContainText("Email подтверждён");
  await page.route("**/api/v1/auth/password-reset/confirm", async (route) => {
    expect(route.request().postDataJSON()).toEqual({
      token: "reset-token",
      password: "NewPassword_2026",
    });
    await route.fulfill({ status: 204 });
  });
  await page.goto("/auth/reset-password?token=reset-token");
  await page.locator("[name=password]").fill("NewPassword_2026");
  await page.locator("[name=repeat]").fill("NewPassword_2026");
  await page.getByRole("button", { name: "Сохранить пароль" }).click();
  await expect(page.getByRole("status")).toContainText("Пароль изменён");
});
test("new teams persist server identifiers and errors remain visible", async ({
  page,
}) => {
  await page.addInitScript(
    (tokens) =>
      localStorage.setItem("tourneyhub.tokens", JSON.stringify(tokens)),
    tokens,
  );
  let teams: unknown[] = [];
  await page.route("**/api/v1/users/me/teams?**", (route) =>
    route.fulfill({ json: pageOf(teams) }),
  );
  await page.route("**/api/v1/teams", async (route) => {
    expect(route.request().postDataJSON()).toEqual({ name: "Команда API" });
    teams = [
      {
        id: "server-team",
        name: "Команда API",
        captain_id: "u1",
        personal: false,
        version: 1,
      },
    ];
    await route.fulfill({ status: 201, json: teams[0] });
  });
  await page.goto("/teams");
  await page
    .getByRole("button", { name: "Создать команду", exact: true })
    .click();
  await page.locator("[name=name]").fill("Команда API");
  await page
    .getByRole("button", { name: "Создать команду", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Команда API" }),
  ).toBeVisible();
  await page.reload();
  await expect(
    page.getByRole("link", { name: "Открыть команду" }),
  ).toHaveAttribute("href", "/teams/server-team");
});

