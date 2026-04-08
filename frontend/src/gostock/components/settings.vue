<script setup>
import { h, onBeforeUnmount, onMounted, ref } from 'vue'
import { NTag, useMessage } from 'naive-ui'
import {
  CheckSponsorCode,
  ExportConfig,
  GetConfig,
  SendDingDingMessageByType,
  TestAIConfigConnection,
  UpdateConfig,
} from '../../../wailsjs/go/main/App'
import { data } from '../../../wailsjs/go/models'
import { EventsEmit, EventsOff, EventsOn } from '../../../wailsjs/runtime'

const message = useMessage()

const SYSTEM_PROMPT_TYPE = '模型系统Prompt'
const USER_PROMPT_TYPE = '模型用户Prompt'
const LEGACY_SYSTEM_PROMPT_TYPE = '濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴濐潟閳ь剙鍊块幐濠冪珶閳哄绉€规洏鍔戝鍫曞箣濠靛牃鍋撻鐑嗘富闁靛牆鎳愮粻浼存煟濡も偓濡繈骞冮悙鍝勫瀭妞ゆ劗濮崇花濠氭⒑閻熺増鎯堟俊顐ｎ殕缁傚秹宕滆绾惧ジ鏌涢幘妤€妫欓妤呮⒑閸涘﹦鎳冮柛鐔告綑閻ｅ嘲煤椤忓嫮鍔﹀銈嗗笂闂勫秵绂嶅鍫熺厵闁绘垶锚閻忋儲銇勮箛鎾舵憼濞ｅ洤锕、鏇㈡晲閸涱垯绱濋梻浣告惈閺堫剟鎯勯姘煎殨濞寸姴顑呮儫閻熸粌绉归崺娑樼暆閸曨兘鎷洪梺闈╁瘜閸欏酣宕濆鍛＜缂備焦顭囩粻鏍庨崶褝韬€规洖銈稿鎾倷闂堟稈鍋撻悙鐑樺仭婵犲﹤鍟扮粻鑽も偓瑙勬磸閸ㄨ崵鍒掑▎鎾村剭濠靛銆噋t'
const LEGACY_USER_PROMPT_TYPE = '濠电姷鏁告慨鐑藉极閸涘﹥鍙忛柣鎴濐潟閳ь剙鍊块幐濠冪珶閳哄绉€规洏鍔戝鍫曞箣濠靛牃鍋撻鐑嗘富闁靛牆鎳愮粻浼存煟濡も偓濡繈骞冮悙鍝勫瀭妞ゆ劗濮崇花濠氭⒑閻熺増鎯堟俊顐ｎ殕缁傚秹宕滆绾惧ジ鏌涢幘妤€妫欓妤呮⒑閸涘﹦鎳冮柛鐔告綑閻ｅ嘲煤椤忓嫮鍔﹀銈嗗笂闂勫秵绂嶅鍫熺厵闁绘垶锚閻忋儵鏌嶈閸撴岸骞冮崒姘辨殾闁靛繈鍊曠粻缁樸亜閺冨倹娅曢柛姗€娼ч—鍐Χ閸℃﹩姊块悗瑙勬礈閺佸骞嗙仦瑙ｆ瀻闊洤锕ラ弬鈧梻浣规灱閺呮盯宕鐐茬９闁规壆澧楅悡娑樏归敐鍥剁劸闁哄棴缍侀弻娑㈠煘閹傚濠碉紕鍋戦崐鏍暜閹烘柡鍋撳鐓庡籍鐎规洘鍨垮畷鎺楁倷鐎电骞愰梻浣告啞濞诧箓宕戞笟鈧崺鈧い鎺嗗亾缂佹劕顩皃t'

const AI_PROVIDER_PRESETS = {
  nvidia: {
    key: 'nvidia',
    name: 'NVIDIA NIM',
    baseUrl: 'https://integrate.api.nvidia.com/v1',
    modelName: 'meta/llama-3.1-70b-instruct',
    temperature: 0.2,
    maxTokens: 4096,
    timeOut: 180,
    description: '适用于 NVIDIA NIM 的 OpenAI 兼容接口。',
  },
  glm: {
    key: 'glm',
    name: 'GLM',
    baseUrl: 'https://open.bigmodel.cn/api/paas/v4',
    modelName: 'glm-4-flash',
    temperature: 0.2,
    maxTokens: 4096,
    timeOut: 180,
    description: '适用于 GLM 模型的 OpenAI 兼容接口。',
  },
  custom: {
    key: 'custom',
    name: '自定义',
    baseUrl: '',
    modelName: '',
    temperature: 0.2,
    maxTokens: 4096,
    timeOut: 180,
    description: '手动配置 OpenAI 兼容服务商。',
  },
}

const providerOptions = [
  { label: 'NVIDIA NIM', value: 'nvidia' },
  { label: 'GLM', value: 'glm' },
  { label: '自定义', value: 'custom' },
]

const formRef = ref(null)
const testingAiConfigIndex = ref(-1)

const formValue = ref({
  ID: 1,
  tushareToken: '',
  dingPush: {
    enable: false,
    dingRobot: '',
  },
  localPush: {
    enable: true,
  },
  updateBasicInfoOnStart: false,
  refreshInterval: 1,
  openAI: {
    enable: false,
    aiConfigs: [],
    prompt: '',
    questionTemplate: '{{stockName}}分析和总结',
    crawlTimeOut: 30,
    kDays: 30,
    httpProxy: '',
    httpProxyEnabled: false,
  },
  enableDanmu: false,
  browserPath: '',
  enableNews: false,
  darkTheme: true,
  enableFund: false,
  enablePushNews: true,
  enableOnlyPushRedNews: true,
  sponsorCode: '',
  httpProxy: '',
  httpProxyEnabled: false,
  enableAgent: false,
  qgqpBId: '',
  assetUnlockPassword: '',
})

function createAiConfig(providerKey = 'custom', overrides = {}) {
  const preset = AI_PROVIDER_PRESETS[providerKey] || AI_PROVIDER_PRESETS.custom
  return new data.AIConfig({
    name: preset.name,
    baseUrl: preset.baseUrl,
    apiKey: '',
    modelName: preset.modelName,
    temperature: preset.temperature,
    maxTokens: preset.maxTokens,
    timeOut: preset.timeOut,
    httpProxy: '',
    httpProxyEnabled: false,
    ...overrides,
  })
}

function normalizeAiConfigs(configs = []) {
  return configs.map((config) => createAiConfig(inferProviderKey(config), config))
}

function ensureStarterAiConfigs() {
  if (formValue.value.openAI.aiConfigs.length > 0) {
    return
  }
  formValue.value.openAI.aiConfigs = [
    createAiConfig('nvidia'),
    createAiConfig('glm'),
  ]
}

function inferProviderKey(aiConfig) {
  const baseUrl = (aiConfig?.baseUrl || '').toLowerCase()
  const name = (aiConfig?.name || '').toLowerCase()
  if (baseUrl.includes('integrate.api.nvidia.com') || name.includes('nvidia')) {
    return 'nvidia'
  }
  if (baseUrl.includes('bigmodel.cn') || name.includes('glm') || name.includes('zhipu')) {
    return 'glm'
  }
  return 'custom'
}

function getProviderDescription(aiConfig) {
  return AI_PROVIDER_PRESETS[inferProviderKey(aiConfig)].description
}

function addAiConfig() {
  formValue.value.openAI.aiConfigs.push(createAiConfig('custom'))
}

function addPresetAiConfig(providerKey) {
  formValue.value.openAI.aiConfigs.push(createAiConfig(providerKey))
}

function removeAiConfig(index) {
  formValue.value.openAI.aiConfigs = formValue.value.openAI.aiConfigs.filter((_, i) => i !== index)
}

function cloneAiConfig(index) {
  const current = formValue.value.openAI.aiConfigs[index]
  const providerKey = inferProviderKey(current)
  formValue.value.openAI.aiConfigs.splice(index + 1, 0, createAiConfig(providerKey, {
    ...current,
    ID: 0,
    name: `${current.name || AI_PROVIDER_PRESETS[providerKey].name} 副本`,
  }))
}

function moveAiConfig(index, direction) {
  const target = index + direction
  if (target < 0 || target >= formValue.value.openAI.aiConfigs.length) {
    return
  }
  const next = [...formValue.value.openAI.aiConfigs]
  const [current] = next.splice(index, 1)
  next.splice(target, 0, current)
  formValue.value.openAI.aiConfigs = next
}

function setPrimaryAiConfig(index) {
  if (index <= 0) {
    return
  }
  moveAiConfig(index, -index)
}

function applyProviderPreset(index, providerKey) {
  const current = formValue.value.openAI.aiConfigs[index]
  const preset = createAiConfig(providerKey, {
    ID: current.ID,
    apiKey: current.apiKey,
    httpProxy: current.httpProxy,
    httpProxyEnabled: current.httpProxyEnabled,
  })
  formValue.value.openAI.aiConfigs.splice(index, 1, preset)
}

async function testAiConfig(index) {
  const aiConfig = formValue.value.openAI.aiConfigs[index]
  testingAiConfigIndex.value = index
  try {
    const result = await TestAIConfigConnection(aiConfig)
    if (result.success) {
      message.success(result.message)
    } else {
      message.error(result.message || '连接失败')
    }
  } catch (error) {
    message.error(error?.message || '连接测试失败')
  } finally {
    testingAiConfigIndex.value = -1
  }
}

onMounted(() => {
  GetConfig().then((res) => {
    formValue.value.ID = res.ID
    formValue.value.tushareToken = res.tushareToken
    formValue.value.dingPush = {
      enable: res.dingPushEnable,
      dingRobot: res.dingRobot,
    }
    formValue.value.localPush = {
      enable: res.localPushEnable,
    }
    formValue.value.updateBasicInfoOnStart = res.updateBasicInfoOnStart
    formValue.value.refreshInterval = res.refreshInterval
    formValue.value.openAI = {
      enable: res.openAiEnable,
      aiConfigs: normalizeAiConfigs(res.aiConfigs || []),
      prompt: res.prompt,
      questionTemplate: res.questionTemplate ? res.questionTemplate : '{{stockName}}分析和总结',
      crawlTimeOut: res.crawlTimeOut,
      kDays: res.kDays,
      httpProxy: '',
      httpProxyEnabled: false,
    }
    ensureStarterAiConfigs()

    formValue.value.enableDanmu = res.enableDanmu
    formValue.value.browserPath = res.browserPath
    formValue.value.enableNews = res.enableNews
    formValue.value.darkTheme = res.darkTheme
    formValue.value.enableFund = res.enableFund
    formValue.value.enablePushNews = res.enablePushNews
    formValue.value.enableOnlyPushRedNews = res.enableOnlyPushRedNews
    formValue.value.sponsorCode = res.sponsorCode
    formValue.value.httpProxy = res.httpProxy
    formValue.value.httpProxyEnabled = res.httpProxyEnabled
    formValue.value.enableAgent = res.enableAgent
    formValue.value.qgqpBId = res.qgqpBId
    formValue.value.assetUnlockPassword = res.assetUnlockPassword || ''
  })

})

onBeforeUnmount(() => {
  message.destroyAll()
  EventsOff('updateSettings')
})

EventsOn('updateSettings', (config) => {
  if (config && Object.prototype.hasOwnProperty.call(config, 'assetUnlockPassword')) {
    formValue.value.assetUnlockPassword = config.assetUnlockPassword || ''
  }
})

function saveConfig() {
  const config = new data.SettingConfig({
    ID: formValue.value.ID,
    dingPushEnable: formValue.value.dingPush.enable,
    dingRobot: formValue.value.dingPush.dingRobot,
    localPushEnable: formValue.value.localPush.enable,
    updateBasicInfoOnStart: formValue.value.updateBasicInfoOnStart,
    refreshInterval: formValue.value.refreshInterval,
    openAiEnable: formValue.value.openAI.enable,
    aiConfigs: formValue.value.openAI.aiConfigs,
    tushareToken: formValue.value.tushareToken,
    prompt: formValue.value.openAI.prompt,
    questionTemplate: formValue.value.openAI.questionTemplate,
    crawlTimeOut: formValue.value.openAI.crawlTimeOut,
    kDays: formValue.value.openAI.kDays,
    enableDanmu: formValue.value.enableDanmu,
    browserPath: formValue.value.browserPath,
    enableNews: formValue.value.enableNews,
    darkTheme: formValue.value.darkTheme,
    enableFund: formValue.value.enableFund,
    enablePushNews: formValue.value.enablePushNews,
    enableOnlyPushRedNews: formValue.value.enableOnlyPushRedNews,
    sponsorCode: formValue.value.sponsorCode,
    httpProxy: formValue.value.httpProxy,
    httpProxyEnabled: formValue.value.httpProxyEnabled,
    enableAgent: formValue.value.enableAgent,
    qgqpBId: formValue.value.qgqpBId,
    assetUnlockPassword: formValue.value.assetUnlockPassword,
  })

  if (config.sponsorCode) {
    CheckSponsorCode(config.sponsorCode).then((res) => {
      if (res.code) {
        UpdateConfig(config).then((saveMessage) => {
          message.success(saveMessage)
          EventsEmit('updateSettings', config)
        })
      } else {
        message.error(res.msg)
      }
    })
  } else {
    UpdateConfig(config).then((saveMessage) => {
      message.success(saveMessage)
      EventsEmit('updateSettings', config)
    })
  }
}

function sendTestNotice() {
  const markdown = `### go-stock test\n${new Date()}`
  const msg = '{' +
    '     "msgtype": "markdown",' +
    '     "markdown": {' +
    `         "title":"go-stock ${new Date()}",` +
    `         "text": "${markdown}"` +
    '     },' +
    '      "at": {' +
    '          "isAtAll": true' +
    '      }' +
    ' }'

  SendDingDingMessageByType(msg, `test-${new Date().getTime()}`, 1).then((res) => {
    message.info(res)
  })
}

function exportConfig() {
  ExportConfig().then((res) => {
    message.info(res)
  })
}

function importConfig() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = (e) => {
    const file = e.target.files[0]
    const reader = new FileReader()
    reader.onload = (readerEvent) => {
      const config = JSON.parse(readerEvent.target.result)
      formValue.value.ID = config.ID
      formValue.value.tushareToken = config.tushareToken
      formValue.value.dingPush = {
        enable: config.dingPushEnable,
        dingRobot: config.dingRobot,
      }
      formValue.value.localPush = {
        enable: config.localPushEnable,
      }
      formValue.value.updateBasicInfoOnStart = config.updateBasicInfoOnStart
      formValue.value.refreshInterval = config.refreshInterval
      formValue.value.openAI = {
        enable: config.openAiEnable,
        aiConfigs: normalizeAiConfigs(config.aiConfigs || []),
        prompt: config.prompt,
        questionTemplate: config.questionTemplate || '{{stockName}}分析和总结',
        crawlTimeOut: config.crawlTimeOut,
        kDays: config.kDays,
      }
      ensureStarterAiConfigs()
      formValue.value.enableDanmu = config.enableDanmu
      formValue.value.browserPath = config.browserPath
      formValue.value.enableNews = config.enableNews
      formValue.value.darkTheme = config.darkTheme
      formValue.value.enableFund = config.enableFund
      formValue.value.enablePushNews = config.enablePushNews
      formValue.value.enableOnlyPushRedNews = config.enableOnlyPushRedNews
      formValue.value.sponsorCode = config.sponsorCode
      formValue.value.httpProxy = config.httpProxy
      formValue.value.httpProxyEnabled = config.httpProxyEnabled
      formValue.value.enableAgent = config.enableAgent
      formValue.value.qgqpBId = config.qgqpBId
      formValue.value.assetUnlockPassword = config.assetUnlockPassword || ''
    }
    reader.readAsText(file)
  }
  input.click()
}

window.onerror = function (event, source, lineno, colno, error) {
  EventsEmit('frontendError', {
    page: 'settings.vue',
    message: event,
    source,
    lineno,
    colno,
    error: error ? error.stack : null,
  })
  return true
}

</script>

<template>
  <n-flex justify="left" style="text-align: left; --wails-draggable:no-drag">
    <n-form ref="formRef" label-placement="left" label-align="left">
      <n-space vertical size="large">
        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => '基础设置')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
            <n-form-item-gi :span="10" label="Tushare Token" path="tushareToken">
              <n-input v-model:value="formValue.tushareToken" type="text" placeholder="请输入 Tushare API Token" clearable />
            </n-form-item-gi>
            <n-form-item-gi :span="4" label="启动时刷新基础数据" path="updateBasicInfoOnStart">
              <n-switch v-model:value="formValue.updateBasicInfoOnStart" />
            </n-form-item-gi>
            <n-form-item-gi :span="4" label="刷新间隔" path="refreshInterval">
              <n-input-number v-model:value="formValue.refreshInterval" placeholder="请输入秒数">
                <template #suffix>秒</template>
              </n-input-number>
            </n-form-item-gi>
            <n-form-item-gi :span="6" label="深色主题" path="darkTheme">
              <n-switch v-model:value="formValue.darkTheme" />
            </n-form-item-gi>
            <n-form-item-gi :span="10" label="浏览器路径" path="browserPath">
              <n-input v-model:value="formValue.browserPath" type="text" placeholder="请输入浏览器安装路径" clearable />
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="基金功能" path="enableFund">
              <n-switch v-model:value="formValue.enableFund" />
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="AI 智能体" path="enableAgent">
              <n-switch v-model:value="formValue.enableAgent" />
            </n-form-item-gi>
            <n-form-item-gi :span="11" label="东方财富 BId" path="qgqpBId">
              <n-input v-model:value="formValue.qgqpBId" type="text" placeholder="请输入东方财富 qgqp BId" clearable />
            </n-form-item-gi>
            <n-form-item-gi :span="11" label="赞助码" path="sponsorCode">
              <n-input-group>
                <n-input v-model:value="formValue.sponsorCode" :show-count="true" placeholder="请输入赞助码" />
                <n-button
                  type="success"
                  secondary
                  strong
                  @click="CheckSponsorCode(formValue.sponsorCode).then((res) => { message.warning(res.msg) })"
                >
                  验证
                </n-button>
              </n-input-group>
            </n-form-item-gi>
          </n-grid>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => '通知设置')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
            <n-form-item-gi :span="3" label="钉钉推送" path="dingPush.enable">
              <n-switch v-model:value="formValue.dingPush.enable" />
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="本地推送" path="localPush.enable">
              <n-switch v-model:value="formValue.localPush.enable" />
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="弹幕功能" path="enableDanmu">
              <n-switch v-model:value="formValue.enableDanmu" />
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="滚动快讯" path="enableNews">
              <n-switch v-model:value="formValue.enableNews" />
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="资讯提醒" path="enablePushNews">
              <n-switch v-model:value="formValue.enablePushNews" />
            </n-form-item-gi>
            <n-form-item-gi v-if="formValue.enablePushNews" :span="4" label="仅提醒红字或关注项" path="enableOnlyPushRedNews">
              <n-switch v-model:value="formValue.enableOnlyPushRedNews" />
            </n-form-item-gi>
            <n-form-item-gi v-if="formValue.dingPush.enable" :span="22" label="钉钉机器人 Webhook" path="dingPush.dingRobot">
              <n-input v-model:value="formValue.dingPush.dingRobot" placeholder="请输入钉钉机器人 Webhook 地址" />
              <n-button type="primary" @click="sendTestNotice">发送测试通知</n-button>
            </n-form-item-gi>
          </n-grid>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => 'AI 设置')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
            <n-form-item-gi :span="24" label="启用 AI 分析" path="openAI.enable">
              <n-switch v-model:value="formValue.openAI.enable" />
            </n-form-item-gi>

            <n-form-item-gi v-if="formValue.openAI.enable" :span="6" label="爬虫超时（秒）" title="资讯抓取超时时间（秒）" path="openAI.crawlTimeOut">
              <n-input-number v-model:value="formValue.openAI.crawlTimeOut" min="30" step="1" />
            </n-form-item-gi>
            <n-form-item-gi v-if="formValue.openAI.enable" :span="4" label="K 线天数" title="天数越多，消耗 token 越多" path="openAI.kDays">
              <n-input-number v-model:value="formValue.openAI.kDays" min="30" max="60" step="1" />
            </n-form-item-gi>
            <n-form-item-gi :span="2" label="HTTP 代理" path="httpProxyEnabled">
              <n-switch v-model:value="formValue.httpProxyEnabled" />
            </n-form-item-gi>
            <n-form-item-gi v-if="formValue.httpProxyEnabled" :span="10" label="代理地址" title="HTTP 代理地址" path="httpProxy">
              <n-input v-model:value="formValue.httpProxy" type="text" placeholder="请输入 HTTP 代理地址" clearable />
            </n-form-item-gi>

            <n-gi v-if="formValue.openAI.enable" :span="24">
              <n-divider title-placement="left">默认提示词设置</n-divider>
            </n-gi>
            <n-form-item-gi v-if="formValue.openAI.enable" :span="12" label="默认系统提示词" path="openAI.prompt">
              <n-input
                v-model:value="formValue.openAI.prompt"
                type="textarea"
                :show-count="true"
                placeholder="请输入默认系统提示词"
                :autosize="{ minRows: 4, maxRows: 8 }"
              />
            </n-form-item-gi>
            <n-form-item-gi v-if="formValue.openAI.enable" :span="12" label="默认个股分析提示词" path="openAI.questionTemplate">
              <n-input
                v-model:value="formValue.openAI.questionTemplate"
                type="textarea"
                :show-count="true"
                placeholder="例如：请分析并总结 {{stockName}} [{{stockCode}}]"
                :autosize="{ minRows: 4, maxRows: 8 }"
              />
            </n-form-item-gi>

            <n-gi v-if="formValue.openAI.enable" :span="24">
              <n-divider title-placement="left">AI 源配置</n-divider>
            </n-gi>
            <n-gi v-if="formValue.openAI.enable" :span="24">
              <n-alert type="info" :show-icon="false" class="ai-alert">
                在这里统一配置所有 AI 服务源。第一个配置会作为应用默认源。NVIDIA 使用 `https://integrate.api.nvidia.com/v1`，GLM 使用 `https://open.bigmodel.cn/api/paas/v4`。
              </n-alert>
            </n-gi>
            <n-gi v-if="formValue.openAI.enable" :span="24">
              <n-space wrap>
                <n-button type="primary" dashed @click="addPresetAiConfig('nvidia')">+ 添加 NVIDIA 源</n-button>
                <n-button type="success" dashed @click="addPresetAiConfig('glm')">+ 添加 GLM 源</n-button>
                <n-button dashed @click="addAiConfig">+ 添加自定义源</n-button>
              </n-space>
            </n-gi>

            <n-gi v-if="formValue.openAI.enable" :span="24">
              <n-space vertical size="large" style="width: 100%">
                <n-card
                  v-for="(aiConfig, index) in formValue.openAI.aiConfigs"
                  :key="aiConfig.ID || `${index}-${aiConfig.name}`"
                  size="small"
                  class="ai-config-card"
                >
                  <template #header>
                    <n-flex justify="space-between" align="center">
                      <n-space align="center" wrap>
                        <n-text strong>AI 源 #{{ index + 1 }}</n-text>
                        <n-tag v-if="index === 0" :bordered="false" type="success">默认</n-tag>
                        <n-tag :bordered="false" type="info">
                          {{ providerOptions.find((item) => item.value === inferProviderKey(aiConfig))?.label }}
                        </n-tag>
                      </n-space>
                      <n-space wrap>
                        <n-button tertiary size="tiny" :disabled="index === 0" @click="setPrimaryAiConfig(index)">设为默认</n-button>
                        <n-button tertiary size="tiny" :disabled="index === 0" @click="moveAiConfig(index, -1)">上移</n-button>
                        <n-button
                          tertiary
                          size="tiny"
                          :disabled="index === formValue.openAI.aiConfigs.length - 1"
                          @click="moveAiConfig(index, 1)"
                        >
                          下移
                        </n-button>
                        <n-button tertiary size="tiny" @click="cloneAiConfig(index)">复制</n-button>
                        <n-button tertiary size="tiny" type="primary" :loading="testingAiConfigIndex === index" @click="testAiConfig(index)">
                          测试
                        </n-button>
                        <n-button tertiary size="tiny" type="error" @click="removeAiConfig(index)">删除</n-button>
                      </n-space>
                    </n-flex>
                  </template>

                  <n-space vertical size="medium" style="width: 100%">
                    <n-text depth="3">{{ getProviderDescription(aiConfig) }}</n-text>
                    <n-grid :cols="24" :x-gap="24">
                      <n-form-item-gi :span="24" hidden :path="`openAI.aiConfigs[${index}].ID`">
                        <n-input v-model:value="aiConfig.ID" />
                      </n-form-item-gi>
                      <n-form-item-gi :span="6" label="服务类型">
                        <n-select
                          :value="inferProviderKey(aiConfig)"
                          :options="providerOptions"
                          @update:value="(value) => applyProviderPreset(index, value)"
                        />
                      </n-form-item-gi>
                      <n-form-item-gi :span="8" :path="`openAI.aiConfigs[${index}].name`" label="名称">
                        <n-input v-model:value="aiConfig.name" type="text" placeholder="例如：NVIDIA 主源 / GLM 备用" clearable />
                      </n-form-item-gi>
                      <n-form-item-gi :span="10" :path="`openAI.aiConfigs[${index}].baseUrl`" label="Base URL">
                        <n-input v-model:value="aiConfig.baseUrl" type="text" placeholder="请输入 AI 服务 Base URL" clearable />
                      </n-form-item-gi>
                      <n-form-item-gi :span="12" :path="`openAI.aiConfigs[${index}].apiKey`" label="API Key">
                        <n-input
                          v-model:value="aiConfig.apiKey"
                          type="password"
                          placeholder="请输入 API Key"
                          clearable
                          show-password-on="click"
                        />
                      </n-form-item-gi>
                      <n-form-item-gi :span="6" :path="`openAI.aiConfigs[${index}].modelName`" label="模型名">
                        <n-input v-model:value="aiConfig.modelName" type="text" placeholder="例如：glm-4-flash" clearable />
                      </n-form-item-gi>
                      <n-form-item-gi :span="3" :path="`openAI.aiConfigs[${index}].temperature`" label="Temperature">
                        <n-input-number v-model:value="aiConfig.temperature" placeholder="temperature" :step="0.1" />
                      </n-form-item-gi>
                      <n-form-item-gi :span="3" :path="`openAI.aiConfigs[${index}].maxTokens`" label="MaxTokens">
                        <n-input-number v-model:value="aiConfig.maxTokens" placeholder="maxTokens" />
                      </n-form-item-gi>
                      <n-form-item-gi :span="4" :path="`openAI.aiConfigs[${index}].timeOut`" label="超时（秒）">
                        <n-input-number v-model:value="aiConfig.timeOut" min="30" step="1" placeholder="超时（秒）" />
                      </n-form-item-gi>
                      <n-form-item-gi :span="4" :path="`openAI.aiConfigs[${index}].httpProxyEnabled`" label="代理">
                        <n-switch v-model:value="aiConfig.httpProxyEnabled" />
                      </n-form-item-gi>
                      <n-form-item-gi v-if="aiConfig.httpProxyEnabled" :span="20" :path="`openAI.aiConfigs[${index}].httpProxy`" label="代理地址">
                        <n-input v-model:value="aiConfig.httpProxy" type="text" placeholder="请输入 HTTP 代理地址" clearable />
                      </n-form-item-gi>
                    </n-grid>
                  </n-space>
                </n-card>
              </n-space>
            </n-gi>

            <n-gi :span="24">
              <n-divider />
            </n-gi>

            <n-gi :span="24">
              <n-space vertical>
                <n-space justify="center">
                  <n-button type="primary" strong @click="saveConfig">保存设置</n-button>
                  <n-button type="info" @click="exportConfig">导出配置</n-button>
                  <n-button type="error" @click="importConfig">导入配置</n-button>
                </n-space>

              </n-space>
            </n-gi>
          </n-grid>
        </n-card>
      </n-space>
    </n-form>
  </n-flex>

</template>

<style scoped>
.ai-alert {
  line-height: 1.7;
}

.ai-config-card {
  width: 100%;
}
</style>
