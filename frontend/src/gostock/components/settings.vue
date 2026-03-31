<script setup>
import {h, onBeforeUnmount, onMounted, ref} from "vue";
import {
  AddPrompt,
  CheckSponsorCode,
  DelPrompt,
  ExportConfig,
  GetConfig,
  GetPromptTemplates,
  SendDingDingMessageByType,
  TestAIConfigConnection,
  UpdateConfig
} from "../../../wailsjs/go/main/App";
import {NTag, useMessage} from "naive-ui";
import {data} from "../../../wailsjs/go/models";
import {EventsEmit, EventsOff, EventsOn} from "../../../wailsjs/runtime";

const message = useMessage()

const AI_PROVIDER_PRESETS = {
  nvidia: {
    key: "nvidia",
    name: "NVIDIA NIM",
    baseUrl: "https://integrate.api.nvidia.com/v1",
    modelName: "meta/llama-3.1-70b-instruct",
    temperature: 0.2,
    maxTokens: 4096,
    timeOut: 180,
    description: "OpenAI-compatible endpoint for NVIDIA NIM.",
  },
  glm: {
    key: "glm",
    name: "GLM",
    baseUrl: "https://open.bigmodel.cn/api/paas/v4",
    modelName: "glm-4-flash",
    temperature: 0.2,
    maxTokens: 4096,
    timeOut: 180,
    description: "OpenAI-compatible endpoint for GLM models.",
  },
  custom: {
    key: "custom",
    name: "Custom",
    baseUrl: "",
    modelName: "",
    temperature: 0.2,
    maxTokens: 4096,
    timeOut: 180,
    description: "Manual OpenAI-compatible provider configuration.",
  }
}

const providerOptions = [
  {label: "NVIDIA NIM", value: "nvidia"},
  {label: "GLM", value: "glm"},
  {label: "Custom", value: "custom"},
]

const formRef = ref(null)
const testingAiConfigIndex = ref(-1)
const formValue = ref({
  ID: 1,
  tushareToken: '',
  dingPush: {
    enable: false,
    dingRobot: ''
  },
  localPush: {
    enable: true,
  },
  updateBasicInfoOnStart: false,
  refreshInterval: 1,
  openAI: {
    enable: false,
    aiConfigs: [],
    prompt: "",
    questionTemplate: "{{stockName}} analysis and summary",
    crawlTimeOut: 30,
    kDays: 30,
    httpProxy: "",
    httpProxyEnabled: false,
  },
  enableDanmu: false,
  browserPath: '',
  enableNews: false,
  darkTheme: true,
  enableFund: false,
  enablePushNews: true,
  enableOnlyPushRedNews: true,
  sponsorCode: "",
  httpProxy: "",
  httpProxyEnabled: false,
  enableAgent: false,
  qgqpBId: '',
  assetUnlockPassword: '',
})

function createAiConfig(providerKey = "custom", overrides = {}) {
  const preset = AI_PROVIDER_PRESETS[providerKey] || AI_PROVIDER_PRESETS.custom
  return new data.AIConfig({
    name: preset.name,
    baseUrl: preset.baseUrl,
    apiKey: "",
    modelName: preset.modelName,
    temperature: preset.temperature,
    maxTokens: preset.maxTokens,
    timeOut: preset.timeOut,
    httpProxy: "",
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
    createAiConfig("nvidia"),
    createAiConfig("glm"),
  ]
}

function inferProviderKey(aiConfig) {
  const baseUrl = (aiConfig?.baseUrl || "").toLowerCase()
  const name = (aiConfig?.name || "").toLowerCase()
  if (baseUrl.includes("integrate.api.nvidia.com") || name.includes("nvidia")) {
    return "nvidia"
  }
  if (baseUrl.includes("bigmodel.cn") || name.includes("glm") || name.includes("zhipu")) {
    return "glm"
  }
  return "custom"
}

function getProviderDescription(aiConfig) {
  return AI_PROVIDER_PRESETS[inferProviderKey(aiConfig)].description
}

function addAiConfig() {
  formValue.value.openAI.aiConfigs.push(createAiConfig("custom"))
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
    name: `${current.name || AI_PROVIDER_PRESETS[providerKey].name} Copy`,
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
      message.error(result.message || "Connection failed")
    }
  } catch (error) {
    message.error(error?.message || "Connection test failed")
  } finally {
    testingAiConfigIndex.value = -1
  }
}

const promptTemplates = ref([])
onMounted(() => {
  GetConfig().then(res => {
    formValue.value.ID = res.ID
    formValue.value.tushareToken = res.tushareToken
    formValue.value.dingPush = {
      enable: res.dingPushEnable,
      dingRobot: res.dingRobot
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
      questionTemplate: res.questionTemplate ? res.questionTemplate : '{{stockName}} analysis and summary',
      crawlTimeOut: res.crawlTimeOut,
      kDays: res.kDays,
      httpProxy: "",
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
  EventsOff("updateSettings")
})

EventsOn("updateSettings", (config) => {
  if (config && Object.prototype.hasOwnProperty.call(config, "assetUnlockPassword")) {
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
    assetUnlockPassword: formValue.value.assetUnlockPassword
  })

  if (config.sponsorCode) {
    CheckSponsorCode(config.sponsorCode).then(res => {
      if (res.code) {
        UpdateConfig(config).then(saveMessage => {
          message.success(saveMessage)
          EventsEmit("updateSettings", config);
        })
      } else {
        message.error(res.msg)
      }
    })
  } else {
    UpdateConfig(config).then(saveMessage => {
      message.success(saveMessage)
      EventsEmit("updateSettings", config);
    })
  }
}

function sendTestNotice() {
  const markdown = "### go-stock test\n" + new Date()
  const msg = '{' +
      '     "msgtype": "markdown",' +
      '     "markdown": {' +
      '         "title":"go-stock' + new Date() + '",' +
      '         "text": "' + markdown + '"' +
      '     },' +
      '      "at": {' +
      '          "isAtAll": true' +
      '      }' +
      ' }'

  SendDingDingMessageByType(msg, "test-" + new Date().getTime(), 1).then(res => {
    message.info(res)
  })
}

function exportConfig() {
  ExportConfig().then(res => {
    message.info(res)
  })
}

function importConfig() {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = '.json';
  input.onchange = (e) => {
    const file = e.target.files[0];
    const reader = new FileReader();
    reader.onload = (readerEvent) => {
      const config = JSON.parse(readerEvent.target.result);
      formValue.value.ID = config.ID
      formValue.value.tushareToken = config.tushareToken
      formValue.value.dingPush = {
        enable: config.dingPushEnable,
        dingRobot: config.dingRobot
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
        questionTemplate: config.questionTemplate || '{{stockName}} analysis and summary',
        crawlTimeOut: config.crawlTimeOut,
        kDays: config.kDays
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
    };
    reader.readAsText(file);
  };
  input.click();
}

window.onerror = function (event, source, lineno, colno, error) {
  EventsEmit("frontendError", {
    page: "settings.vue",
    message: event,
    source: source,
    lineno: lineno,
    colno: colno,
    error: error ? error.stack : null
  });
  return true;
};

const showManagePromptsModal = ref(false)
const promptTypeOptions = [
  {label: "婵犵數濮烽弫鍛婃叏閻戝鈧倿鎸婃竟鈺嬬秮瀹曘劑寮堕幋婵堚偓顓烆渻閵堝懐绠伴柣妤€妫濋幃鐐哄垂椤愮姳绨婚梺鐟版惈濡绂嶉崜褏纾奸柛鎾楀棙顎楅梺鍛婄懃閸熸潙鐣峰ú顏勭劦妞ゆ帊闄嶆禍婊堟煙閻戞ê鐏ユい蹇撶摠娣囧﹪顢曢敐鍛紝闂佸搫鏈惄顖氼嚕娴犲惟鐟滃秹鍩涘畝鍕拺闁革富鍙庨崝婊呯磼缂佹绠栧ǎ鍥э躬瀹曞ジ寮撮悙闈涒偓鐐烘偡濠婂啰绠荤€规洏鍨荤划娆撴儍婵夌〇pt", value: '婵犵數濮烽弫鍛婃叏閻戝鈧倿鎸婃竟鈺嬬秮瀹曘劑寮堕幋婵堚偓顓烆渻閵堝懐绠伴柣妤€妫濋幃鐐哄垂椤愮姳绨婚梺鐟版惈濡绂嶉崜褏纾奸柛鎾楀棙顎楅梺鍛婄懃閸熸潙鐣峰ú顏勭劦妞ゆ帊闄嶆禍婊堟煙閻戞ê鐏ユい蹇撶摠娣囧﹪顢曢敐鍛紝闂佸搫鏈惄顖氼嚕娴犲惟鐟滃秹鍩涘畝鍕拺闁革富鍙庨崝婊呯磼缂佹绠栧ǎ鍥э躬瀹曞ジ寮撮悙闈涒偓鐐烘偡濠婂啰绠荤€规洏鍨荤划娆撴儍婵夌〇pt'},
  {label: "婵犵數濮烽弫鍛婃叏閻戝鈧倿鎸婃竟鈺嬬秮瀹曘劑寮堕幋婵堚偓顓烆渻閵堝懐绠伴柣妤€妫濋幃鐐哄垂椤愮姳绨婚梺鐟版惈濡绂嶉崜褏纾奸柛鎾楀棙顎楅梺鍛婄懃閸熸潙鐣峰ú顏勭劦妞ゆ帊闄嶆禍婊堟煙閻戞ê鐏ラ柍褜鍓氶幃鍌氱暦閵忋倕绠绘い鏃傛櫕閸橀潧顪冮妶鍡橆梿鐎规洜鏁婚幆灞解枎韫囧﹥鏂€闂佹枼鏅涢崯顖炲磹閹扮増鐓涘ù锝囶焾閺嗭綁鏌涢埞鎯т壕婵＄偑鍊栫敮鎺斺偓姘煎弮瀹曟垿宕掗悙瀵稿幐闂佸憡娲﹂崑渚€鍩€椤掆偓缁愬pt", value: '婵犵數濮烽弫鍛婃叏閻戝鈧倿鎸婃竟鈺嬬秮瀹曘劑寮堕幋婵堚偓顓烆渻閵堝懐绠伴柣妤€妫濋幃鐐哄垂椤愮姳绨婚梺鐟版惈濡绂嶉崜褏纾奸柛鎾楀棙顎楅梺鍛婄懃閸熸潙鐣峰ú顏勭劦妞ゆ帊闄嶆禍婊堟煙閻戞ê鐏ラ柍褜鍓氶幃鍌氱暦閵忋倕绠绘い鏃傛櫕閸橀潧顪冮妶鍡橆梿鐎规洜鏁婚幆灞解枎韫囧﹥鏂€闂佹枼鏅涢崯顖炲磹閹扮増鐓涘ù锝囶焾閺嗭綁鏌涢埞鎯т壕婵＄偑鍊栫敮鎺斺偓姘煎弮瀹曟垿宕掗悙瀵稿幐闂佸憡娲﹂崑渚€鍩€椤掆偓缁愬pt'},
]
const formPromptRef = ref(null)
const formPrompt = ref({
  ID: 0,
  Name: '',
  Content: '',
  Type: '',
})

function savePrompt() {
  AddPrompt(formPrompt.value).then(res => {
    message.success(res)
    GetPromptTemplates("", "").then(promptRes => {
      promptTemplates.value = promptRes
    })
    showManagePromptsModal.value = false
  })
}

function editPrompt(prompt) {
  formPrompt.value.ID = prompt.ID
  formPrompt.value.Name = prompt.name
  formPrompt.value.Content = prompt.content
  formPrompt.value.Type = prompt.type
  showManagePromptsModal.value = true
}

function deletePrompt(ID) {
  DelPrompt(ID).then(res => {
    message.success(res)
    GetPromptTemplates("", "").then(promptRes => {
      promptTemplates.value = promptRes
    })
  })
}
</script>

<template>
  <n-flex justify="left" style="text-align: left; --wails-draggable:no-drag">
    <n-form ref="formRef" :label-placement="'left'" :label-align="'left'">
      <n-space vertical size="large">
        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => 'Basic settings')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
            <n-form-item-gi :span="10" label="Tushare Token" path="tushareToken">
              <n-input type="text" placeholder="Tushare api token" v-model:value="formValue.tushareToken" clearable/>
            </n-form-item-gi>
            <n-form-item-gi :span="4" label="Refresh basic data on startup" path="updateBasicInfoOnStart">
              <n-switch v-model:value="formValue.updateBasicInfoOnStart"/>
            </n-form-item-gi>
            <n-form-item-gi :span="4" label="Refresh interval" path="refreshInterval">
              <n-input-number v-model:value="formValue.refreshInterval" placeholder="Enter refresh interval in seconds">
                <template #suffix>s</template>
              </n-input-number>
            </n-form-item-gi>
            <n-form-item-gi :span="6" label="Dark theme" path="darkTheme">
              <n-switch v-model:value="formValue.darkTheme"/>
            </n-form-item-gi>
            <n-form-item-gi :span="10" label="Browser path" path="browserPath">
              <n-input type="text" placeholder="Browser install path" v-model:value="formValue.browserPath" clearable/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="Funds" path="enableFund">
              <n-switch v-model:value="formValue.enableFund"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="AI agent" path="enableAgent">
              <n-switch v-model:value="formValue.enableAgent"/>
            </n-form-item-gi>
            <n-form-item-gi :span="11" label="QGQP BId" path="qgqpBId">
              <n-input type="text" placeholder="Eastmoney qgqp bid" v-model:value="formValue.qgqpBId" clearable/>
            </n-form-item-gi>
            <n-form-item-gi :span="11" label="Sponsor code" path="sponsorCode">
              <n-input-group>
                <n-input :show-count="true" placeholder="Sponsor code" v-model:value="formValue.sponsorCode"/>
                <n-button type="success" secondary strong
                          @click="CheckSponsorCode(formValue.sponsorCode).then((res) => {message.warning(res.msg)})">濠电姷鏁告慨鎾儉婢舵劕绾ч幖瀛樻尭娴滅偓淇婇妶鍕妽闁告瑥绻橀弻鐔虹磼閵忕姵鐏嶉梺绋垮椤ㄥ懘濡撮幒鎴僵闁挎繂鎳嶆竟鏇㈡⒒?
                </n-button>
	              </n-input-group>
	            </n-form-item-gi>
	          </n-grid>
	        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => 'Notification settings')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
            <n-form-item-gi :span="3" label="DingTalk push" path="dingPush.enable">
              <n-switch v-model:value="formValue.dingPush.enable"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="Local push" path="localPush.enable">
              <n-switch v-model:value="formValue.localPush.enable"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="Danmu" path="enableDanmu">
              <n-switch v-model:value="formValue.enableDanmu"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="News ticker" path="enableNews">
              <n-switch v-model:value="formValue.enableNews"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="Push news" path="enablePushNews">
              <n-switch v-model:value="formValue.enablePushNews"/>
            </n-form-item-gi>
            <n-form-item-gi v-if="formValue.enablePushNews" :span="4" label="Only push red or watched news" path="enableOnlyPushRedNews">
              <n-switch v-model:value="formValue.enableOnlyPushRedNews"/>
            </n-form-item-gi>
            <n-form-item-gi :span="22" v-if="formValue.dingPush.enable" label="DingTalk robot webhook" path="dingPush.dingRobot">
              <n-input placeholder="Enter DingTalk robot webhook" v-model:value="formValue.dingPush.dingRobot"/>
              <n-button type="primary" @click="sendTestNotice">Send test notification</n-button>
            </n-form-item-gi>
          </n-grid>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => 'AI settings')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left;">
            <n-form-item-gi :span="24" label="AI analysis" path="openAI.enable">
              <n-switch v-model:value="formValue.openAI.enable"/>
            </n-form-item-gi>

            <n-form-item-gi :span="6" v-if="formValue.openAI.enable" label="Crawler Timeout (s)" title="News crawl timeout in seconds" path="openAI.crawlTimeOut">
              <n-input-number min="30" step="1" v-model:value="formValue.openAI.crawlTimeOut"/>
            </n-form-item-gi>
            <n-form-item-gi :span="4" v-if="formValue.openAI.enable" title="More days will use more tokens" label="K-line days" path="openAI.kDays">
              <n-input-number min="30" step="1" max="60" v-model:value="formValue.openAI.kDays"/>
            </n-form-item-gi>
            <n-form-item-gi :span="2" label="HTTP proxy" path="httpProxyEnabled">
              <n-switch v-model:value="formValue.httpProxyEnabled"/>
            </n-form-item-gi>
            <n-form-item-gi :span="10" v-if="formValue.httpProxyEnabled" title="http濠电姷鏁告慨鐑藉极閹间礁纾绘繛鎴欏焺閺佸銇勯幘璺烘瀻闁搞劍妫冮幃妤呮偨濞堣法鍔哥紓浣界堪閸婃繈寮诲☉婊庢Ъ濡炪們鍔岄幊鎰垝婵犳艾鍐€鐟滃寮ㄦ禒瀣厽婵☆垰鎼痪褔鏌熼崗鐓庡闁哄瞼鍠栭幃鐑藉箥椤旇偐鍘介梻浣告惈婢跺洭宕滃┑鍡╁殫闁告洦鍘搁崑鎾绘晲鎼存繃鍠氶梺鎼炲€曢鍥╂? label="http濠电姷鏁告慨鐑藉极閹间礁纾绘繛鎴欏焺閺佸銇勯幘璺烘瀻闁搞劍妫冮幃妤呮偨濞堣法鍔哥紓浣界堪閸婃繈寮诲☉婊庢Ъ濡炪們鍔岄幊鎰垝婵犳艾鍐€鐟滃寮ㄦ禒瀣厽婵☆垰鎼痪褔鏌熼崗鐓庡闁哄瞼鍠栭幃鐑藉箥椤旇偐鍘介梻浣告惈婢跺洭宕滃┑鍡╁殫闁告洦鍘搁崑鎾绘晲鎼存繃鍠氶梺鎼炲€曢鍥╂? path="httpProxy">
              <n-input type="text" placeholder="HTTP proxy address" v-model:value="formValue.httpProxy" clearable/>
            </n-form-item-gi>

            <n-gi :span="24" v-if="formValue.openAI.enable">
              <n-divider title-placement="left">Default prompts</n-divider>
            </n-gi>
            <n-form-item-gi :span="12" v-if="formValue.openAI.enable" label="Default system prompt" path="openAI.prompt">
              <n-input v-model:value="formValue.openAI.prompt" type="textarea" :show-count="true" placeholder="Enter a default system prompt" :autosize="{ minRows: 4, maxRows: 8 }"/>
            </n-form-item-gi>
            <n-form-item-gi :span="12" v-if="formValue.openAI.enable" label="Default stock analysis prompt" path="openAI.questionTemplate">
              <n-input v-model:value="formValue.openAI.questionTemplate" type="textarea" :show-count="true" placeholder="For example: analyze and summarize {{stockName}} [{{stockCode}}]" :autosize="{ minRows: 4, maxRows: 8 }"/>
            </n-form-item-gi>

            <n-gi :span="24" v-if="formValue.openAI.enable">
              <n-divider title-placement="left">AI source configuration</n-divider>
            </n-gi>
            <n-gi :span="24" v-if="formValue.openAI.enable">
              <n-alert type="info" :show-icon="false" class="ai-alert">
                Configure all AI sources here. The first source acts as the default for the rest of the app. NVIDIA uses https://integrate.api.nvidia.com/v1 and GLM uses https://open.bigmodel.cn/api/paas/v4.
              </n-alert>
            </n-gi>
            <n-gi :span="24" v-if="formValue.openAI.enable">
              <n-space wrap>
                <n-button type="primary" dashed @click="addPresetAiConfig('nvidia')">+ Add NVIDIA source</n-button>
                <n-button type="success" dashed @click="addPresetAiConfig('glm')">+ Add GLM source</n-button>
                <n-button dashed @click="addAiConfig">+ Add custom source</n-button>
              </n-space>
            </n-gi>

            <n-gi :span="24" v-if="formValue.openAI.enable">
              <n-space vertical size="large" style="width: 100%">
                <n-card v-for="(aiConfig, index) in formValue.openAI.aiConfigs" :key="aiConfig.ID || `${index}-${aiConfig.name}`" size="small" class="ai-config-card">
                  <template #header>
                    <n-flex justify="space-between" align="center">
                      <n-space align="center" wrap>
                        <n-text strong>AI Source #{{ index + 1 }}</n-text>
                        <n-tag :bordered="false" type="success" v-if="index === 0">Primary</n-tag>
                        <n-tag :bordered="false" type="info">{{ providerOptions.find(item => item.value === inferProviderKey(aiConfig))?.label }}</n-tag>
                      </n-space>
                      <n-space wrap>
                        <n-button tertiary size="tiny" @click="setPrimaryAiConfig(index)" :disabled="index === 0">Set primary</n-button>
                        <n-button tertiary size="tiny" @click="moveAiConfig(index, -1)" :disabled="index === 0">Move up</n-button>
                        <n-button tertiary size="tiny" @click="moveAiConfig(index, 1)" :disabled="index === formValue.openAI.aiConfigs.length - 1">Move down</n-button>
                        <n-button tertiary size="tiny" @click="cloneAiConfig(index)">Clone</n-button>
                        <n-button tertiary size="tiny" type="primary" :loading="testingAiConfigIndex === index" @click="testAiConfig(index)">Test</n-button>
                        <n-button tertiary size="tiny" type="error" @click="removeAiConfig(index)">Delete</n-button>
                      </n-space>
                    </n-flex>
                  </template>

                  <n-space vertical size="medium" style="width: 100%">
                    <n-text depth="3">{{ getProviderDescription(aiConfig) }}</n-text>
                    <n-grid :cols="24" :x-gap="24">
                      <n-form-item-gi :span="24" hidden :path="`openAI.aiConfigs[${index}].ID`">
                        <n-input v-model:value="aiConfig.ID"/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="6" label="Provider">
                        <n-select :value="inferProviderKey(aiConfig)" :options="providerOptions" @update:value="(value) => applyProviderPreset(index, value)"/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="8" label="Source name" :path="`openAI.aiConfigs[${index}].name`">
                        <n-input type="text" placeholder="e.g. NVIDIA primary / GLM backup" v-model:value="aiConfig.name" clearable/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="10" label="Base URL" :path="`openAI.aiConfigs[${index}].baseUrl`">
                        <n-input type="text" placeholder="AI provider base URL" v-model:value="aiConfig.baseUrl" clearable/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="12" label="API Key" :path="`openAI.aiConfigs[${index}].apiKey`">
                        <n-input type="password" placeholder="闂傚倸鍊搁崐椋庣矆娓氣偓楠炴牠顢曚綅閸ヮ剦鏁嶉柣鎰綑娴滆鲸绻濋悽闈浶㈡繛灞傚€楃划缁樺鐎涙鍘甸梻鍌氬€搁顓⑺囬敃鍌涚厽妞ゆ挾鍣ュ▓婊堟煛鐏炲墽娲撮柛鈺佸瀹曟﹢顢旈崨顓熺彫闂備浇顕х€涒晝鍠婂鍛殕闁归棿绀侀拑鐔兼煛閸モ晛鏋庣紒鍓佸仦缁绘盯骞嬮悜鍥︾返濠电偛妯婇崣鍐潖閾忓湱纾兼俊顖滃帶閳峰本绻涚€涙鐭婇柣鏍с偢閻涱噣宕橀妸搴㈡瀹曘劑顢欓梻瀵哥处闂傚倷绀佹竟濠囨偂閸儱纾婚柛鏇ㄥ亐閺嬫梻鈧厜鍋撻柍褜鍓涘Σ鎰板箻鐎涙ê顎撻梻鍌氱墛缁嬫垿鈥栨径鎰拺缂備焦蓱閹牏绱掔紒妯肩疄闁?API Key" v-model:value="aiConfig.apiKey" clearable show-password-on="click"/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="6" label="Model name" :path="`openAI.aiConfigs[${index}].modelName`">
                        <n-input type="text" placeholder="濠电姷鏁告慨鐑藉极閹间礁纾婚柣鎰惈缁犳澘鈹戦悩宕囶暡闁搞倕鑻灃闁挎繂鎳庨弸銈夋煛娴ｇ顏柡灞剧☉閳藉鈻庨幋鐐搭唭闂?glm-4-flash" v-model:value="aiConfig.modelName" clearable/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="3" label="Temperature" :path="`openAI.aiConfigs[${index}].temperature`">
                        <n-input-number placeholder="temperature" v-model:value="aiConfig.temperature" :step="0.1"/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="3" label="MaxTokens" :path="`openAI.aiConfigs[${index}].maxTokens`">
                        <n-input-number placeholder="maxTokens" v-model:value="aiConfig.maxTokens"/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="4" label="Timeout (s)" :path="`openAI.aiConfigs[${index}].timeOut`">
                        <n-input-number min="30" step="1" placeholder="Timeout (s)" v-model:value="aiConfig.timeOut"/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="4" label="Proxy" :path="`openAI.aiConfigs[${index}].httpProxyEnabled`">
                        <n-switch v-model:value="aiConfig.httpProxyEnabled"/>
                      </n-form-item-gi>
                      <n-form-item-gi :span="20" v-if="aiConfig.httpProxyEnabled" title="HTTP proxy address" :path="`openAI.aiConfigs[${index}].httpProxy`">
                        <n-input type="text" placeholder="HTTP proxy address" v-model:value="aiConfig.httpProxy" clearable/>
                      </n-form-item-gi>
                    </n-grid>
                  </n-space>
                </n-card>
              </n-space>
            </n-gi>

            <n-gi :span="24">
              <n-divider/>
            </n-gi>

            <n-gi :span="24">
              <n-space vertical>
                <n-space justify="center">
                  <n-button type="primary" strong @click="saveConfig">Save settings</n-button>
                  <n-button type="info" @click="exportConfig">Export settings</n-button>
                  <n-button type="error" @click="importConfig">Import settings</n-button>
                </n-space>

                <n-flex justify="start" style="margin-top: 10px" v-if="promptTemplates.length > 0">
                  <n-tag :bordered="false" type="warning">Prompt templates</n-tag>
                  <n-tag size="medium" secondary v-for="prompt in promptTemplates" closable
                         @close="deletePrompt(prompt.ID)" @click="editPrompt(prompt)" :title="prompt.content"
                         :type="prompt.type === '婵犵數濮烽弫鍛婃叏閻戝鈧倿鎸婃竟鈺嬬秮瀹曘劑寮堕幋婵堚偓顓烆渻閵堝懐绠伴柣妤€妫濋幃鐐哄垂椤愮姳绨婚梺鐟版惈濡绂嶉崜褏纾奸柛鎾楀棙顎楅梺鍛婄懃閸熸潙鐣峰ú顏勭劦妞ゆ帊闄嶆禍婊堟煙閻戞ê鐏ユい蹇撶摠娣囧﹪顢曢敐鍛紝闂佸搫鏈惄顖氼嚕娴犲惟鐟滃秹鍩涘畝鍕拺闁革富鍙庨崝婊呯磼缂佹绠栧ǎ鍥э躬瀹曞ジ寮撮悙闈涒偓鐐烘偡濠婂啰绠荤€规洏鍨荤划娆撴儍婵夌〇pt' ? 'success' : 'info'" :bordered="false">
                    {{ prompt.name }}
                  </n-tag>
                </n-flex>
              </n-space>
            </n-gi>
          </n-grid>
        </n-card>
      </n-space>
    </n-form>
  </n-flex>

  <n-modal v-model:show="showManagePromptsModal" closable :mask-closable="false">
    <n-card style="width: 800px; height: 600px; text-align: left" :bordered="false"
            :title="(formPrompt.ID > 0 ? 'Edit' : 'Add') + ' prompt template'" size="huge" role="dialog" aria-modal="true">
      <n-form ref="formPromptRef" :label-placement="'left'" :label-align="'left'">
        <n-form-item label="Name">
          <n-input v-model:value="formPrompt.Name" placeholder="Enter a prompt template name"/>
        </n-form-item>
        <n-form-item label="Type">
          <n-select v-model:value="formPrompt.Type" :options="promptTypeOptions" placeholder="Select a prompt type"/>
        </n-form-item>
        <n-form-item label="Content">
          <n-input v-model:value="formPrompt.Content" type="textarea" :show-count="true" placeholder="Enter prompt content"
                   :autosize="{ minRows: 12, maxRows: 12 }"/>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-flex justify="end">
          <n-button type="primary" @click="savePrompt">Save</n-button>
          <n-button type="warning" @click="showManagePromptsModal = false">Cancel</n-button>
        </n-flex>
      </template>
    </n-card>
  </n-modal>
</template>

<style scoped>
.ai-alert {
  width: 100%;
}

.ai-config-card {
  width: 100%;
}
</style>
