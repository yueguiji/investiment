<script setup>
import 'md-editor-v3/lib/preview.css'
import { h, onBeforeUnmount, onMounted, ref } from 'vue'
import { CheckUpdate, GetVersionInfo, OpenURL } from "../../../wailsjs/go/main/App"
import { Environment, EventsOff, EventsOn } from "../../../wailsjs/runtime"
import { NAvatar, NButton, useNotification } from "naive-ui"

const updateLog = ref('')
const versionInfo = ref('dev')
const icon = ref('')
const notify = useNotification()

const upstreamRepo = 'https://github.com/ArvinLovegood/go-stock'
const wailsRepo = 'https://github.com/wailsapp/wails'
const vueRepo = 'https://github.com/vuejs/core'
const naiveUiRepo = 'https://github.com/tusen-ai/naive-ui'

async function openUrl(url) {
  const env = await Environment()
  if (env.platform === 'windows') {
    window.open(url)
    return
  }
  OpenURL(url)
}

onMounted(() => {
  document.title = '关于'
  GetVersionInfo().then((res) => {
    updateLog.value = res.content || ''
    versionInfo.value = res.version || 'dev'
    icon.value = res.icon || ''
  })
})

onBeforeUnmount(() => {
  notify.destroyAll()
  EventsOff("updateVersion")
})

EventsOn("updateVersion", async (msg) => {
  const utcDate = new Date(msg.published_at)
  const date = new Date(utcDate.getTime())
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  const formattedDate = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`

  notify.info({
    avatar: () =>
      h(NAvatar, {
        size: 'small',
        round: false,
        src: icon.value
      }),
    title: '发现新版本: ' + msg.tag_name,
    content: () => {
      return h('div', {
        style: {
          'text-align': 'left',
          'font-size': '14px'
        }
      }, { default: () => msg.commit?.message })
    },
    duration: 5000,
    meta: "发布时间:" + formattedDate,
    action: () => {
      return h(NButton, {
        type: 'primary',
        size: 'small',
        onClick: () => openUrl(msg.html_url)
      }, { default: () => '查看' })
    }
  })
})
</script>

<template>
  <n-space vertical size="large" style="--wails-draggable:no-drag">
    <n-card size="large">
      <n-divider title-placement="center">关于软件</n-divider>
      <n-space vertical>
        <n-image v-if="icon" width="100" :src="icon" preview-disabled />
        <h1>
          <n-badge :value="versionInfo" :offset="[120,10]" type="success">
            <n-gradient-text type="info" :size="42">Rubin Investment</n-gradient-text>
          </n-badge>
        </h1>
        <n-button size="tiny" type="info" tertiary @click="CheckUpdate(1)">检查更新</n-button>
        <div class="content-block">
          <p>一个面向本地运行的投资研究桌面应用，整合投资跟踪、资产管理、量化模板与 AI 分析能力。</p>
          <p>运行时数据默认写入程序目录下的 <code>data/</code> 与 <code>logs/</code>，开源仓库本身不包含私人数据库与私人凭据。</p>
          <p v-if="updateLog">版本说明：{{ updateLog }}</p>
          <p class="risk-note">本软件仅供学习研究和流程演示，不构成任何投资建议。</p>
        </div>
      </n-space>

      <n-divider title-placement="center">开源说明</n-divider>
      <div class="content-block">
        <p>当前桌面应用在 <code>go-stock</code> 开源项目基础上进行了二次整合与扩展，包含家庭资产、组合管理与量化工作流等模块。</p>
        <p>仓库发布时采用“公开代码仓库 + 本地私有配置注入”的方式，敏感配置请放在本机，不要提交到版本库。</p>
        <p>
          上游项目致谢：
          <a href="#" @click.prevent="openUrl(upstreamRepo)">go-stock</a>
        </p>
      </div>

      <n-divider title-placement="center">反馈与协作</n-divider>
      <div class="content-block">
        <p>公开版本建议通过代码托管平台的 Issue 和 Pull Request 反馈问题、提交修复或讨论改进方案。</p>
        <p>涉及私人凭据、打包差异或本地数据库的内容，请不要直接贴到公开 Issue。</p>
      </div>

      <n-divider title-placement="center">技术栈</n-divider>
      <div class="content-block">
        <p>
          <a href="#" @click.prevent="openUrl(wailsRepo)">Wails</a>
          <n-divider vertical />
          <a href="#" @click.prevent="openUrl(vueRepo)">Vue</a>
          <n-divider vertical />
          <a href="#" @click.prevent="openUrl(naiveUiRepo)">Naive UI</a>
          <n-divider vertical />
          <a href="#" @click.prevent="openUrl(upstreamRepo)">go-stock</a>
        </p>
      </div>
    </n-card>
  </n-space>
</template>

<style scoped>
h1 {
  margin: 0;
  padding: 6px 0;
}

p {
  margin: 2px 0;
}

a {
  color: #18a058;
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}

code {
  font-family: "Cascadia Code", "Consolas", monospace;
}

.content-block {
  justify-self: center;
  text-align: left;
}

.risk-note {
  color: crimson;
}
</style>
