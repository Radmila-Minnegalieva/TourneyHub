import { test, expect } from '@playwright/test'
import process from 'node:process'
test('live PostgreSQL: register, verify, profile, teams, publish, apply, draw, result and rating', async ({ page, browser, request }) => {
  test.skip(process.env.TOURNEYHUB_LIVE !== '1', 'Requires a running backend, Mailpit and a local administrator')
  test.setTimeout(180_000)
  const email = process.env.TEST_ADMIN_EMAIL!, password = process.env.TEST_ADMIN_PASSWORD!
  expect(email).toBeTruthy(); expect(password).toBeTruthy()
  const login = await request.post('/api/v1/auth/login', { data: { email, password } }); expect(login.ok()).toBeTruthy()
  const adminTokens = await login.json(); const headers = { Authorization: `Bearer ${adminTokens.access_token}` }
  const suffix=Date.now().toString(); const playerEmail=`player${suffix}@example.test`; const playerPassword='Integration_Test_2026!'
  const disciplineResponse=await request.post('/api/v1/admin/disciplines',{headers,data:{name:`Шахматы ${suffix}`,min_roster_size:1,max_roster_size:1,allows_draw:true,score_unit:'очки'}}); expect(disciplineResponse.status()).toBe(201); const discipline=await disciplineResponse.json()
  await page.goto('/auth/register'); await page.locator('[name=email]').fill(playerEmail); await page.locator('[name=nickname]').fill(`p${suffix}`); await page.locator('[name=name]').fill('Проверка интеграции'); await page.locator('[name=password]').fill(playerPassword); await page.locator('[name=repeat]').fill(playerPassword); await page.getByRole('checkbox').check(); await page.getByRole('button',{name:'Зарегистрироваться'}).click(); await expect(page.getByRole('status').filter({hasText:'Проверьте почту'})).toBeVisible()
  let verifyToken=''
  await expect.poll(async () => {
    const response=await request.get('http://localhost:8025/api/v1/messages'); const mailbox=await response.json(); const letter=mailbox.messages?.find((m:{To:{Address:string}[]})=>m.To.some(to=>to.Address===playerEmail)); if(!letter) return false
    const detail=await (await request.get(`http://localhost:8025/api/v1/message/${letter.ID}`)).json(); verifyToken=detail.Text.match(/token=([^\s]+)/)?.[1] || ''; return Boolean(verifyToken)
  },{timeout:80_000,intervals:[1000,2000,3000]}).toBe(true)
  await page.goto(`/auth/verify-email?token=${verifyToken}`); await page.getByRole('button',{name:'Подтвердить email'}).click(); await expect(page.getByRole('status').filter({hasText:'Email подтверждён'})).toBeVisible()
  await page.goto('/auth/login'); await page.locator('[name=email]').fill(playerEmail); await page.locator('[name=password]').fill(playerPassword); await page.getByRole('main').getByRole('button',{name:'Войти',exact:true}).click(); await expect(page).toHaveURL(/tournaments$/)
  await page.goto('/profile'); await page.locator('[name=name]').fill('Имя сохранено в PostgreSQL'); await page.getByRole('button',{name:'Сохранить',exact:true}).click(); await expect(page.getByRole('status')).toHaveText('Изменения сохранены'); await page.reload(); await expect(page.locator('[name=name]')).toHaveValue('Имя сохранено в PostgreSQL')
  await page.goto('/teams'); await page.getByRole('button',{name:'Создать команду',exact:true}).click(); await page.locator('[name=name]').fill(`Команда ${suffix}`); await page.getByRole('button',{name:'Создать команду',exact:true}).click(); await page.getByRole('link',{name:'Открыть команду'}).click(); await expect(page.getByRole('heading',{name:`Команда ${suffix}`,exact:true})).toBeVisible(); await page.reload(); await expect(page.getByRole('main').getByText(`@p${suffix}`,{exact:false})).toBeVisible()
  const adminContext=await browser.newContext(); const admin=await adminContext.newPage()
  await admin.addInitScript(tokens=>localStorage.setItem('tourneyhub.tokens',JSON.stringify(tokens)),adminTokens)
  await admin.goto('/tournaments/new'); await admin.locator('[name=name]').fill(`Кубок ${suffix}`); await admin.locator('[name=discipline]').selectOption(discipline.id)
  await admin.getByRole('button',{name:'Далее',exact:true}).click()
  await admin.getByRole('button',{name:'Далее',exact:true}).click()
  const localTime=(date:Date) => { const d=new Date(date.getTime()-date.getTimezoneOffset()*60000); return d.toISOString().slice(0,16) }
  const now=Date.now()
  for(const [key,delta] of [['registration_opens_at',-3600000],['registration_closes_at',3600000],['starts_at',7200000],['ends_at',86400000]] as const) await admin.locator(`[name=${key}]`).fill(localTime(new Date(now+delta)))
  await admin.getByRole('button',{name:'Далее',exact:true}).click()
  await admin.getByRole('button',{name:'Сохранить черновик'}).click(); await expect(admin.getByRole('button',{name:'Опубликовать',exact:true})).toBeVisible(); const tournamentId=new URL(admin.url()).pathname.split('/')[2]
  await admin.getByRole('button',{name:'Опубликовать',exact:true}).click(); await expect(admin.getByRole('status').filter({hasText:'Турнир опубликован'})).toBeVisible()
  await page.goto(`/tournaments/${tournamentId}`); await page.getByRole('button',{name:'Подать заявку',exact:true}).click(); await expect(page.getByRole('status').filter({hasText:'Заявка отправлена'})).toBeVisible()
  await admin.reload(); await admin.getByRole('button',{name:'Подать заявку',exact:true}).click(); await expect(admin.getByRole('status').filter({hasText:'Заявка отправлена'})).toBeVisible()
  await admin.getByRole('link',{name:'Участники и заявки'}).click(); await expect(admin.getByRole('button',{name:'Принять',exact:true})).toHaveCount(2)
  await admin.getByRole('button',{name:'Принять',exact:true}).first().click(); await expect(admin.getByRole('button',{name:'Принять',exact:true})).toHaveCount(1); await admin.getByRole('button',{name:'Принять',exact:true}).click(); await expect(admin.getByRole('button',{name:'Принять',exact:true})).toHaveCount(0)
  // Advance the test fixture's registration window without touching the clock or database directly.
  const t=await (await request.get(`/api/v1/tournaments/${tournamentId}`,{headers})).json(); const adjusted=await request.patch(`/api/v1/tournaments/${tournamentId}`,{headers,data:{version:t.version,registration_opens_at:new Date(Date.now()-3600000).toISOString(),registration_closes_at:new Date(Date.now()-180000).toISOString(),starts_at:new Date(Date.now()-120000).toISOString()}}); expect(adjusted.ok()).toBeTruthy()
  await admin.reload(); await admin.getByRole('button',{name:'Предпросмотр жеребьёвки по Эло'}).click(); await admin.getByRole('button',{name:'Подтвердить жеребьёвку'}).click(); await expect(admin.getByRole('status').filter({hasText:'Жеребьёвка подтверждена'})).toBeVisible()
  await admin.getByRole('link',{name:'Расписание',exact:true}).click(); await admin.getByRole('link',{name:/Раунд 1:/}).click()
  await admin.locator('[name=starts_at]').fill(localTime(new Date(Date.now()-60000))); await admin.locator('[name=venue]').fill('Тестовая площадка'); await admin.getByRole('button',{name:'Назначить матч'}).click(); await expect(admin.getByRole('status').filter({hasText:'Расписание сохранено'})).toBeVisible()
  const spectator = await browser.newPage({ baseURL: 'http://127.0.0.1:5173' })
  await spectator.goto(admin.url())
  await expect(spectator.getByText('Запланирован · Раунд 1')).toBeVisible()
  await admin.locator('[name=home1]').fill('1'); await admin.locator('[name=away1]').fill('0'); await admin.getByRole('button',{name:'Зафиксировать результат судьёй'}).click(); await expect(admin.getByRole('status').filter({hasText:'Результат отправлен'})).toBeVisible(); await expect(admin.getByText('Завершён · Раунд 1')).toBeVisible()
  await expect(spectator.getByText('Завершён · Раунд 1')).toBeVisible({ timeout: 5000 })
  await spectator.close()
  await admin.goto(`/tournaments/${tournamentId}`); await expect(admin.getByRole('heading',{name:`Кубок ${suffix}`})).toBeVisible(); if(await admin.getByRole('button',{name:'Завершить турнир'}).isVisible()) { await admin.getByRole('button',{name:'Завершить турнир'}).click() }; await expect.poll(async()=> (await (await request.get(`/api/v1/tournaments/${tournamentId}`,{headers})).json()).status).toBe('completed')
  await admin.getByRole('link',{name:'Итоги',exact:true}).click(); await expect(admin.getByRole('link',{name:'Скачать итоги CSV'})).toBeVisible()
  await page.goto('/rating'); await page.getByRole('combobox',{name:'Дисциплина'}).selectOption(discipline.id); await expect(page.locator('tbody tr')).toHaveCount(2)
  await page.screenshot({path:'test-results/live-rating.png',fullPage:true})
  await adminContext.close()
})
