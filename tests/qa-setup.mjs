import { spawnSync } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const dir = fileURLToPath(new URL('.', import.meta.url));
const root = fileURLToPath(new URL('..', import.meta.url));
const manifest = JSON.parse(readFileSync(join(dir, 'fixtures', 'manifest.json'), 'utf8'));
const project = process.env.TOURNEYHUB_QA_PROJECT || 'tourneyhub-qa';
if (!/^tourneyhub-qa(?:-[a-z0-9]+)*$/.test(project)) throw Error('TOURNEYHUB_QA_PROJECT должен начинаться с tourneyhub-qa.');
const args = new Set(process.argv.slice(2));
if ([...args].some(arg => !['--local','--stop'].includes(arg))) throw Error('Допустимые параметры: --local, --stop.');
const offset = Number(process.env.TOURNEYHUB_QA_PORT_OFFSET || '0');
if (!Number.isInteger(offset) || offset < 0 || offset > 50000) throw Error('TOURNEYHUB_QA_PORT_OFFSET должен быть целым числом от 0 до 50000.');
const ports = { frontend:5173+offset, api:8080+offset, swagger:8081+offset, mailpit:8025+offset, db:5433+offset };
const local = `http://localhost:${ports.frontend}`;
const stateDir = join(dir, '.qa', project);
const statePath = join(stateDir, 'state.json');
const overlay = join(stateDir, 'compose.qa.yaml');
const agent = `${project}-ngrok`;
const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms));
const newPassword = () => `QA!${randomBytes(12).toString('hex')}z`;

function docker(arguments_, { input, capture = false } = {}) {
  const result = spawnSync('docker', arguments_, { cwd:root, env:environment, encoding:'utf8', input, stdio:capture?['pipe','pipe','pipe']:'inherit', maxBuffer:8*1024*1024 });
  if (result.error) throw Error(`Не удалось запустить Docker. ${result.error.message}`);
  if (result.status !== 0) {
    const operation = arguments_.slice(0, 2).join(' ');
    let detail = capture ? result.stderr.trim() : '';
    for (const secret of [environment.NGROK_AUTHTOKEN, state.databasePassword, ...Object.values(state.passwords)]) {
      if (secret) detail = detail.replaceAll(secret, '[скрыто]');
    }
    throw Error(`Docker ${operation} завершился с кодом ${result.status}.${detail ? ` ${detail}` : ''}`);
  }
  return capture ? result.stdout.trim() : '';
}

if (args.has('--stop') && !existsSync(statePath)) throw Error(`Стенд ${project} ещё не создавался.`);
const state = existsSync(statePath) ? JSON.parse(readFileSync(statePath,'utf8')) : {
  installId: randomBytes(16).toString('hex'), databasePassword: newPassword(), passwords: Object.fromEntries(manifest.accounts.map(a => [a.email,newPassword()])), publicUrl: local, offset,
};
if (state.offset !== offset) throw Error('Для существующего проекта используй тот же TOURNEYHUB_QA_PORT_OFFSET. Для другой конфигурации создай новый QA-проект.');
if (!/^[a-f0-9]{32}$/.test(state.installId) || !/^QA![a-f0-9]{24}z$/.test(state.databasePassword) || manifest.accounts.some(a => !/^QA![a-f0-9]{24}z$/.test(state.passwords[a.email]))) throw Error('Некорректная локальная конфигурация. Для чистого стенда задай новое имя QA-проекта.');
const environment = { ...process.env, DB_USER:'tourneyhub', DB_PASSWORD:state.databasePassword, API_PORT:String(ports.api), POSTGRES_PORT:String(ports.db), MAILPIT_PORT:String(ports.mailpit), SWAGGER_PORT:String(ports.swagger), POSTGRES_IMAGE:'postgres:16-alpine', POSTGRES_DATA_PATH:'/var/lib/postgresql/data', POSTGRES_VOLUME_NAME:`${project}_postgres_data_v16`, PUBLIC_WEB_URL:state.publicUrl };
const composeArgs = ['compose','-p',project,'-f',join(root,'compose.yaml'),'-f',overlay];
const compose = (...values) => docker([...composeArgs,...values]);
const save = () => writeFileSync(statePath, JSON.stringify(state,null,2)+'\n',{encoding:'utf8',mode:0o600});
const sql = (statement) => docker([...composeArgs,'exec','-T','postgres','psql','-X','-U','tourneyhub','-d','tourneyhub','-qAt','-v','ON_ERROR_STOP=1','--single-transaction'], {input:statement,capture:true});
const exists = () => docker(['ps','-a','--filter',`name=^/${agent}$`,'--format','{{.Names}}'],{capture:true}) === agent;

function prepareFiles() {
  mkdirSync(stateDir,{recursive:true,mode:0o700});
  save();
  writeFileSync(overlay,`services:\n  frontend:\n    ports: !override\n      - "127.0.0.1:${ports.frontend}:80"\n  api:\n    environment:\n      WORKER_INTERVAL: 1s\n      CORS_ALLOWED_ORIGINS: http://localhost:${ports.frontend},http://127.0.0.1:${ports.frontend},http://localhost:${ports.swagger},http://127.0.0.1:${ports.swagger}\n`,'utf8');
}

function restore() {
  compose('up','-d','--wait','--wait-timeout','90','postgres');
  compose('run','--rm','--no-deps','migrate');
  if (sql("SELECT to_regclass('qa_fixture.seed') IS NOT NULL;\n") === 't') {
    if (sql("SELECT version || ':' || install_id FROM qa_fixture.seed;\n") !== `${manifest.version}:${state.installId}`) throw Error('Том данных не соответствует локальной конфигурации. Для чистого запуска задай другое имя QA-проекта.');
    console.log('Данные уже загружены. Текущий прогресс сохранён.');
    return;
  }
  if (sql('SELECT (SELECT count(*) FROM users) + (SELECT count(*) FROM tournaments) + (SELECT count(*) FROM disciplines);\n') !== '0') throw Error('База уже содержит данные. Для чистого запуска задай другое имя QA-проекта.');
  const copy = readFileSync(join(dir,'fixtures','snapshot.sql'),'utf8').replace(/\r\n/g,'\n');
  const capturedAt = new Date(manifest.capturedAt).toISOString();
  const credentialRows = manifest.accounts.map(a => `('${a.id}'::uuid,'${state.passwords[a.email]}')`).join(',\n');
  const shift = (table, columns) => `UPDATE ${table} SET ${columns.split(' ').map(c=>`${c}=${c}+d.delta`).join(', ')} FROM qa_dates d;`;
  const times = [
    ['users','email_verified_at blocked_at personal_data_consent_at created_at updated_at'],
    ['teams','deleted_at created_at updated_at'], ['team_members','joined_at'], ['invitations','expires_at created_at'],
    ['tournaments','registration_opens_at registration_closes_at starts_at ends_at deleted_at created_at updated_at'],
    ['applications','created_at updated_at'], ['matches','starts_at created_at updated_at'],
    ['result_reports','confirmation_deadline confirmed_at created_at'], ['disputes','resolved_at created_at'],
    ['rating_changes','created_at'], ['notifications','read_at created_at'], ['audit_entries','created_at'],
  ];
  const assertions = Object.entries(manifest.counts).map(([table,n]) => `IF (SELECT count(*) FROM public.${table}) <> ${n} THEN RAISE EXCEPTION 'Wrong fixture count: ${table}'; END IF;`).join('\n');
  sql(`SET session_replication_role = replica;\n${copy}\nSET session_replication_role = origin;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
UPDATE users u SET password_hash=crypt(p.password,gen_salt('bf',10)) FROM (VALUES ${credentialRows}) p(id,password) WHERE p.id=u.id;
CREATE TEMP TABLE qa_dates AS SELECT now()-'${capturedAt}'::timestamptz AS delta;
${times.map(([t,c])=>shift(t,c)).join('\n')}
UPDATE invitations SET expires_at=GREATEST(expires_at,now()+interval '7 days') WHERE status='pending';
UPDATE result_reports SET confirmation_deadline=GREATEST(confirmation_deadline,now()+interval '24 hours') WHERE status='pending';
DO $$ BEGIN
${assertions}
IF EXISTS (SELECT FROM users WHERE password_hash='qa-password-reset-required') THEN RAISE EXCEPTION 'Missing fixture password'; END IF;
END $$;
CREATE SCHEMA qa_fixture;
CREATE TABLE qa_fixture.seed(version integer NOT NULL, install_id text NOT NULL);
INSERT INTO qa_fixture.seed VALUES (${manifest.version}, '${state.installId}');\n`);
  console.log(`Восстановлено ${manifest.counts.users} аккаунтов, ${manifest.counts.teams} команд и ${manifest.counts.tournaments} турниров.`);
}

async function getOk(url, check = () => true) {
  for (let attempt=0; attempt<20; attempt++) {
    try {
      const response=await fetch(url,{headers:{'ngrok-skip-browser-warning':'true'},signal:AbortSignal.timeout(4000)});
      if (response.ok && await check(response)) return;
    } catch { /* Service is still starting. */ }
    await delay(1000);
  }
  throw Error(`Сервис не ответил по адресу ${url}.`);
}

async function startTunnel() {
  if (exists()) {
    const running = docker(['inspect','--format','{{.State.Running}}',agent],{capture:true});
    if (running === 'true' && /^https:\/\/[a-z0-9.-]+$/i.test(state.publicUrl)) return state.publicUrl;
    if (running !== 'true') docker(['start',agent]);
  } else {
    if (!environment.NGROK_AUTHTOKEN) throw Error('Задай свой NGROK_AUTHTOKEN и повтори запуск. Для локального запуска используй --local.');
    docker(['run','-d','--restart','unless-stopped','--name',agent,'--network',`${project}_default`,'-e','NGROK_AUTHTOKEN','ngrok/ngrok:latest','http','http://frontend:80','--log=stdout','--log-format=json']);
  }
  delete environment.NGROK_AUTHTOKEN;
  for (let i=0; i<30; i++) {
    const logs = spawnSync('docker',['logs','--tail','50',agent],{cwd:root,env:environment,encoding:'utf8'});
    const entries = `${logs.stdout??''}\n${logs.stderr??''}`.split(/\r?\n/).flatMap(line => {
      try { return [JSON.parse(line)]; } catch { return []; }
    });
    const started=entries.findLast(e => e.msg === 'started tunnel' && /^https:\/\/[a-z0-9.-]+\/?$/i.test(e.url));
    if (started) return started.url.replace(/\/$/,'');
    await delay(1000);
  }
  throw Error(`ngrok не выдал URL. Проверь docker logs ${agent}. Локальная версия уже запущена.`);
}

function writeGuide(url) {
  const paths = {
    site:url, login:`${url}/auth/login`, mailpit:`http://localhost:${ports.mailpit}`, swagger:`http://localhost:${ports.swagger}`,
    ...Object.fromEntries(Object.entries(manifest.scenarios.tournament).map(([name,id])=>[name,`${url}/tournaments/${id}`])),
    personal:`${url}/tournaments/${manifest.scenarios.personal}`,
    ...Object.fromEntries(['pending','disputed','reportMatch','refereeMatch'].map(name=>[name,`${url}/matches/${manifest.scenarios[name].id}`])),
    ...Object.fromEntries(manifest.scenarios.teams.map((team,index)=>[`team${index+1}`,`${url}/teams/${team.id}`])),
  };
  const cases=readFileSync(join(dir,'test-cases.md'),'utf8').replace(/\{\{([^}]+)}}/g,(_,key)=>{
    if (key.startsWith('password:') && state.passwords[key.slice(9)]) return state.passwords[key.slice(9)];
    if (paths[key]) return paths[key];
    throw Error(`Неизвестная подстановка в тест-кейсах: ${key}`);
  });
  const accounts = manifest.accounts.map(a => `| ${a.nickname} | \`${a.email}\` | \`${state.passwords[a.email]}\` | ${a.roles.join(', ')} | ${a.blocked?'Заблокирован':a.verified?'Вход разрешён':'Почта не подтверждена'} |`);
  writeFileSync(join(stateDir,'test-cases.md'),cases,{encoding:'utf8',mode:0o600});
  writeFileSync(join(stateDir,'accounts.md'),[
    '# Аккаунты тестового стенда', '', `Вход: ${paths.login}`, `Mailpit: ${paths.mailpit}`, `Swagger: ${paths.swagger}`, '',
    '| Никнейм | Логин | Пароль | Роли | Статус |','|---|---|---|---|---|',...accounts,'',
    'Пароли относятся только к этому локальному стенду. Файл не предназначен для коммита.','',
  ].join('\n'),{encoding:'utf8',mode:0o600});
}

try {
  docker(['version','--format','{{.Server.Version}}'],{capture:true});
  prepareFiles();
  compose('config','--quiet');
  if (args.has('--stop')) {
    if (exists()) docker(['stop',agent]);
    compose('stop');
    console.log('Стенд остановлен. Том и пароли сохранены.');
  } else {
    if (args.has('--local') && state.publicUrl !== local) {
      if (exists()) docker(['stop',agent]);
      state.publicUrl = local;
      environment.PUBLIC_WEB_URL = local;
      save();
    }
    compose('build','api','frontend');
    restore();
    compose('up','-d','--wait','--wait-timeout','120','--no-build');
    await getOk(`${local}/api/v1/tournaments?limit=1`,async r=>(await r.json()).total>0);
    await getOk(`http://localhost:${ports.mailpit}/api/v1/info`);
    writeGuide(local);
    if (!args.has('--local')) {
      const url = await startTunnel();
      if (state.publicUrl !== url) {
        state.publicUrl = url;
        environment.PUBLIC_WEB_URL = url;
        compose('up','-d','--no-deps','--wait','--wait-timeout','60','api');
        compose('restart','frontend');
        save();
      }
      await getOk(`${url}/tournaments`,async r=>(await r.text()).includes('id="root"'));
      await getOk(`${url}/api/v1/tournaments?limit=1`,async r=>(await r.json()).total>0);
      writeGuide(url);
      console.log(`Публичная ссылка: ${url}`);
    }
    console.log(`Фронтенд: ${local}\nMailpit: http://localhost:${ports.mailpit}\nSwagger: http://localhost:${ports.swagger}\nПошаговые кейсы: ${join(stateDir,'test-cases.md')}\nПароли: ${join(stateDir,'accounts.md')}`);
  }
} catch (error) {
  console.error(error.message);
  process.exitCode=1;
}
