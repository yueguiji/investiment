import fs from 'node:fs/promises'
import path from 'node:path'
import os from 'node:os'
import http from 'node:http'
import { fileURLToPath } from 'node:url'
import { chromium } from '../.codex/skills/wails-ui-automation/scripts/node_modules/playwright/index.mjs'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const repoRoot = path.resolve(__dirname, '..')
const distRoot = path.join(repoRoot, 'frontend', 'dist')
const tempRoot = path.join(os.tmpdir(), 'investment-quant-linkage-regression')
const port = 41731

const templates = [
  {
    id: 101,
    name: '宽基ETF多篮子轮动网格',
    description: '跟踪沪深300 上证50 中证500 中证1000 创业板 科创50 红利 低波和行业 ETF 的轮动网格。',
    searchKeywords: '宽基 ETF 网格 轮动 沪深300 上证50 中证500 中证1000 红利 行业',
    strategyType: 'grid',
    scriptCategory: 'etf-rotation',
    brokerPlatform: 'QMT',
    language: 'Python',
    status: 'active',
    tags: ['宽基', 'ETF', '网格', '轮动', 'Python', '行业']
  },
  {
    id: 102,
    name: '红利低波防守轮动',
    description: '适合风险偏好回落阶段的红利低波和防守切换脚本。',
    searchKeywords: '红利 低波 防守 ETF',
    strategyType: 'rotation',
    scriptCategory: 'defensive',
    brokerPlatform: 'QMT',
    language: 'Python',
    status: 'active',
    tags: ['红利', '低波', '防守', 'Python']
  },
  {
    id: 103,
    name: '行业动量轮动',
    description: '聚焦热门行业 ETF 和强主题轮动。',
    searchKeywords: '行业 ETF 动量 轮动 科技 芯片 AI',
    strategyType: 'rotation',
    scriptCategory: 'sector',
    brokerPlatform: 'QMT',
    language: 'Python',
    status: 'draft',
    tags: ['行业', 'ETF', '动量', 'Python', '科技']
  }
]

const aiConfigs = [
  { ID: 1, name: '默认配置', modelName: 'gpt-4.1-mini' },
  { ID: 2, name: '备用AI', modelName: 'deepseek-chat' }
]

const taxonomy = [
  { key: '宽基', label: '宽基' },
  { key: 'ETF', label: 'ETF' },
  { key: '网格', label: '网格' },
  { key: '轮动', label: '轮动' },
  { key: '行业', label: '行业' },
  { key: '动量', label: '动量' },
  { key: '红利', label: '红利' },
  { key: '低波', label: '低波' },
  { key: '防守', label: '防守' },
  { key: 'Python', label: 'Python' },
  { key: '科技', label: '科技' }
]

const hotTopics = [
  { name: 'AI算力' },
  { name: '芯片' },
  { name: '机器人' },
  { name: '央企红利' }
]

const hotStocks = [
  { name: '中际旭创' },
  { name: '工业富联' },
  { name: '寒武纪' },
  { name: '中国石油' }
]

const moneyRank = Array.from({ length: 10 }, (_, index) => ({
  name: `活跃资金${index + 1}`,
  netamount: 1000000 - index * 10000
}))

const industryRank = [
  { 板块名称: '半导体', 涨跌幅: '2.86' },
  { 板块名称: 'AI算力', 涨跌幅: '2.43' },
  { 板块名称: '机器人', 涨跌幅: '1.98' }
]

const globalIndexes = [
  { name: '沪深300', code: '000300', zdf: '1.35' },
  { name: '上证50', code: '000016', zdf: '1.12' },
  { name: '中证A50', code: '930050', zdf: '1.08' },
  { name: '中证500', code: '000905', zdf: '1.24' },
  { name: '中证1000', code: '000852', zdf: '1.46' },
  { name: '创业板指', code: '399006', zdf: '1.18' },
  { name: '科创50', code: '000688', zdf: '1.09' },
  { name: '红利指数', code: '000015', zdf: '0.88' }
]

const mockScript = `
window.__mockNotifications = [];
window.runtime = {
  EventsOn() { return () => {}; },
  EventsOnMultiple() { return () => {}; },
  EventsOff() {},
  EventsOffAll() {},
  EventsEmit() {},
  WindowReload() {},
  WindowReloadApp() {},
  WindowSetAlwaysOnTop() {},
  WindowSetSystemDefaultTheme() {},
  WindowSetLightTheme() {},
  WindowSetDarkTheme() {},
  WindowCenter() {},
  WindowSetTitle() {},
  WindowFullscreen() {},
  WindowUnfullscreen() {},
  WindowIsFullscreen() { return Promise.resolve(false); },
  WindowGetSize() { return Promise.resolve({ width: 1600, height: 980 }); },
  WindowSetSize() {},
  WindowSetMaxSize() {},
  WindowSetMinSize() {},
  WindowSetPosition() {},
  WindowGetPosition() { return Promise.resolve({ x: 0, y: 0 }); },
  WindowHide() {},
  WindowShow() {},
  WindowMaximise() {},
  WindowToggleMaximise() {},
  WindowUnmaximise() {},
  WindowIsMaximised() { return Promise.resolve(false); },
  WindowMinimise() {},
  WindowUnminimise() {},
  WindowSetBackgroundColour() {},
  ScreenGetAll() { return Promise.resolve([]); },
  WindowIsMinimised() { return Promise.resolve(false); },
  WindowIsNormal() { return Promise.resolve(true); },
  BrowserOpenURL() {},
  Environment() { return Promise.resolve({}); },
  Quit() {},
  Hide() {},
  Show() {},
  ClipboardGetText() { return Promise.resolve(''); },
  ClipboardSetText() { return Promise.resolve(true); },
  LogPrint() {},
  LogTrace() {},
  LogDebug() {},
  LogInfo() {},
  LogWarning() {},
  LogError() {},
  LogFatal() {}
};

window.go = {
  main: {
    App: {
      GetQuantTemplates() {
        return Promise.resolve([${JSON.stringify(templates)}, ${templates.length}]);
      },
      GetQuantTagTaxonomy() {
        return Promise.resolve(${JSON.stringify(taxonomy)});
      },
      GetAiConfigs() {
        return Promise.resolve(${JSON.stringify(aiConfigs)});
      },
      HotTopic() {
        return Promise.resolve(${JSON.stringify(hotTopics)});
      },
      HotStock() {
        return Promise.resolve(${JSON.stringify(hotStocks)});
      },
      GetMoneyRankSina() {
        return Promise.resolve(${JSON.stringify(moneyRank)});
      },
      GetIndustryRank() {
        return Promise.resolve(${JSON.stringify(industryRank)});
      },
      GlobalStockIndexes() {
        return Promise.resolve(${JSON.stringify(globalIndexes)});
      },
      AnalyzeQuantLinkageWithAI(payload, configId) {
        return Promise.resolve({
          aiEnabled: true,
          model: configId === 2 ? '备用AI / deepseek-chat' : '默认配置 / gpt-4.1-mini',
          message: 'AI 推荐已更新。',
          parsed: {
            scriptName: '宽基ETF多篮子轮动网格',
            action: '优先切换',
            reason: '宽基指数强度、成长主题和行业热度同时共振，ETF 网格轮动脚本更贴合当前节奏。',
            riskHint: '若量能快速回落，建议降低网格频率并控制仓位。',
            confidence: 8
          }
        });
      },
      SendLocalNotification(title, body, key) {
        window.__mockNotifications.push({ title, body, key });
        return Promise.resolve(true);
      },
      OpenURL() {
        return Promise.resolve(true);
      }
    }
  }
};
`

async function ensureDir(dir) {
  await fs.mkdir(dir, { recursive: true })
}

async function copyDir(source, target) {
  await fs.rm(target, { recursive: true, force: true })
  await ensureDir(target)
  const entries = await fs.readdir(source, { withFileTypes: true })
  for (const entry of entries) {
    const srcPath = path.join(source, entry.name)
    const destPath = path.join(target, entry.name)
    if (entry.isDirectory()) {
      await copyDir(srcPath, destPath)
    } else {
      await fs.copyFile(srcPath, destPath)
    }
  }
}

async function prepareMockSite() {
  await copyDir(distRoot, tempRoot)
  const indexPath = path.join(tempRoot, 'index.html')
  const mockPath = path.join(tempRoot, 'mock-wails.js')
  let html = await fs.readFile(indexPath, 'utf8')
  html = html.replace('<script type="module" crossorigin src="/assets/', '<script src="/mock-wails.js"></script>\n    <script type="module" crossorigin src="/assets/')
  await fs.writeFile(indexPath, html, 'utf8')
  await fs.writeFile(mockPath, mockScript, 'utf8')
}

function createStaticServer(rootDir) {
  return http.createServer(async (req, res) => {
    const requestUrl = new URL(req.url, `http://127.0.0.1:${port}`)
    let filePath = path.join(rootDir, requestUrl.pathname === '/' ? 'index.html' : requestUrl.pathname.slice(1))
    try {
      let stat = await fs.stat(filePath)
      if (stat.isDirectory()) {
        filePath = path.join(filePath, 'index.html')
        stat = await fs.stat(filePath)
      }
      const ext = path.extname(filePath).toLowerCase()
      const contentType =
        ext === '.html' ? 'text/html; charset=utf-8' :
        ext === '.js' ? 'application/javascript; charset=utf-8' :
        ext === '.css' ? 'text/css; charset=utf-8' :
        ext === '.json' ? 'application/json; charset=utf-8' :
        'application/octet-stream'
      res.writeHead(200, { 'Content-Type': contentType })
      const data = await fs.readFile(filePath)
      res.end(data)
    } catch {
      res.writeHead(404)
      res.end('Not found')
    }
  })
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

async function run() {
  await prepareMockSite()
  const server = createStaticServer(tempRoot)
  await new Promise((resolve) => server.listen(port, '127.0.0.1', resolve))

  const browser = await chromium.launch({ headless: true })
  const page = await browser.newPage({ viewport: { width: 1600, height: 1200 } })
  const consoleErrors = []

  page.on('pageerror', (error) => {
    consoleErrors.push(`pageerror: ${error.message}`)
  })
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      consoleErrors.push(`console: ${msg.text()}`)
    }
  })

  try {
    await page.goto(`http://127.0.0.1:${port}/#/quant/linkage`, { waitUntil: 'networkidle' })
    await page.locator('.n-page-header__title', { hasText: '联动推荐' }).waitFor({ timeout: 15000 })
    await page.getByText('双轨推荐', { exact: true }).waitFor({ timeout: 15000 })

    await page.waitForFunction(() => document.body.innerText.includes('宽基ETF多篮子轮动网格'))
    const initialText = await page.locator('body').innerText()
    assert(!initialText.includes('当前没有可供 AI 评审的 Python 脚本'), 'AI 推荐区仍显示没有可评审脚本')
    assert(!initialText.includes('当前暂无规则推荐结果'), '规则推荐仍为空')
    assert(initialText.includes('宽基ETF多篮子轮动网格'), '未出现预期的宽基 ETF 规则推荐')
    assert(initialText.includes('AI 推荐已更新') || initialText.includes('备用AI / deepseek-chat') || initialText.includes('默认配置 / gpt-4.1-mini'), 'AI 周期推荐未生成')

    await page.getByText('默认 AI 源').click()
    await page.getByText('备用AI / deepseek-chat').click()
    await page.waitForTimeout(400)
    const storedAfterSelect = await page.evaluate(() => localStorage.getItem('investment.quant.linkage.ai-config-id'))
    assert(storedAfterSelect === '2', `AI 源未写入 localStorage，实际为 ${storedAfterSelect}`)

    await page.goto(`http://127.0.0.1:${port}/#/quant/templates`, { waitUntil: 'networkidle' })
    await page.goto(`http://127.0.0.1:${port}/#/quant/linkage`, { waitUntil: 'networkidle' })
    await page.waitForTimeout(800)

    const linkageText = await page.locator('body').innerText()
    assert(linkageText.includes('备用AI / deepseek-chat'), '切页返回后 AI 源没有保持为上次选择')
    assert(linkageText.includes('宽基ETF多篮子轮动网格'), '切页返回后推荐脚本没有出现')

    const notifications = await page.evaluate(() => window.__mockNotifications || [])
    assert(notifications.some((item) => item.body.includes('Python 规则推荐')), '没有触发 Python 规则通知')
    assert(notifications.some((item) => item.body.includes('AI 周期推荐')), '没有触发 AI 周期通知')

    if (consoleErrors.length) {
      throw new Error(`页面存在运行错误:\n${consoleErrors.join('\n')}`)
    }

    console.log(JSON.stringify({
      ok: true,
      checkedAt: new Date().toISOString(),
      notifications,
      storedAiConfigId: storedAfterSelect
    }, null, 2))
  } finally {
    await browser.close()
    await new Promise((resolve) => server.close(resolve))
  }
}

run().catch((error) => {
  console.error(error)
  process.exit(1)
})
