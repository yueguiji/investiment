<template>
  <div class="agent-chat">
    <div class="toolbar">
      <n-select
        v-model:value="selectedModelId"
        :options="modelOptions"
        label-field="name"
        value-field="ID"
        placeholder="选择模型"
        size="small"
        style="width: 220px"
      />
    </div>

    <div ref="messageContainerRef" class="message-list">
      <div
        v-for="(message, index) in messages"
        :key="`${message.role}-${index}`"
        class="message-item"
        :class="message.role"
      >
        <div class="message-header">
          <span class="name">{{ message.name }}</span>
          <span class="time">{{ message.datetime }}</span>
        </div>
        <div v-if="message.reasoning" class="reasoning">{{ message.reasoning }}</div>
        <div class="content">{{ message.content }}</div>
      </div>
    </div>

    <div class="sender">
      <n-input
        v-model:value="inputValue"
        type="textarea"
        :autosize="{ minRows: 3, maxRows: 5 }"
        placeholder="请输入消息..."
        @keydown.enter.exact.prevent="sendMessage"
      />
      <div class="actions">
        <n-button @click="clearMessages">清空</n-button>
        <n-button type="primary" :loading="loading" @click="sendMessage">发送</n-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { nextTick, onBeforeMount, onBeforeUnmount, ref } from 'vue'
import { NButton, NInput, NSelect } from 'naive-ui'
import { ChatWithAgent, GetAiConfigs } from '../../../wailsjs/go/main/App'
import { EventsOff, EventsOn } from '../../../wailsjs/runtime'

const inputValue = ref('')
const loading = ref(false)
const modelOptions = ref([])
const selectedModelId = ref(null)
const messageContainerRef = ref(null)

const messages = ref([
  {
    role: 'assistant',
    name: 'Go-Stock AI',
    datetime: '',
    reasoning: '',
    content: '我是您的AI赋能股票分析助手，您可以问我任何关于股票投资的问题。',
  },
])

function formatNow() {
  return new Date().toLocaleString('zh-CN', { hour12: false })
}

function scrollToBottom() {
  nextTick(() => {
    const container = messageContainerRef.value
    if (container) {
      container.scrollTop = container.scrollHeight
    }
  })
}

function ensureAssistantPlaceholder() {
  const lastMessage = messages.value[messages.value.length - 1]
  if (!lastMessage || lastMessage.role !== 'assistant') {
    messages.value.push({
      role: 'assistant',
      name: 'Go-Stock AI',
      datetime: formatNow(),
      reasoning: '',
      content: '',
    })
  }
  return messages.value[messages.value.length - 1]
}

function handleAgentMessage(data) {
  if (data?.role !== 'assistant') {
    return
  }

  const currentMessage = ensureAssistantPlaceholder()

  if (data.reasoning_content) {
    currentMessage.reasoning += data.reasoning_content
  }
  if (data.content) {
    currentMessage.content += data.content
  }
  if (Array.isArray(data.tool_calls)) {
    for (const tool of data.tool_calls) {
      const toolName = tool?.function?.name || 'tool'
      const args = tool?.function?.arguments || '无参数'
      currentMessage.reasoning += `\n[${toolName}] ${args}\n`
    }
  }
  if (data.response_meta?.finish_reason === 'stop') {
    loading.value = false
  }

  scrollToBottom()
}

async function sendMessage() {
  const question = inputValue.value.trim()
  if (!question || loading.value) {
    return
  }

  messages.value.push({
    role: 'user',
    name: '我',
    datetime: formatNow(),
    reasoning: '',
    content: question,
  })

  messages.value.push({
    role: 'assistant',
    name: 'Go-Stock AI',
    datetime: formatNow(),
    reasoning: '',
    content: '',
  })

  loading.value = true
  inputValue.value = ''
  scrollToBottom()

  try {
    await ChatWithAgent(question, selectedModelId.value || 0, 0)
  } catch (error) {
    loading.value = false
    messages.value[messages.value.length - 1].content = `请求失败：${error}`
    scrollToBottom()
  }
}

function clearMessages() {
  messages.value = [
    {
      role: 'assistant',
      name: 'Go-Stock AI',
      datetime: '',
      reasoning: '',
      content: '我是您的AI赋能股票分析助手，您可以问我任何关于股票投资的问题。',
    },
  ]
}

onBeforeMount(async () => {
  const configs = (await GetAiConfigs()) || []
  modelOptions.value = configs
  if (configs.length > 0) {
    selectedModelId.value = configs[0].ID
  }
  EventsOn('agent-message', handleAgentMessage)
})

onBeforeUnmount(() => {
  EventsOff('agent-message')
})
</script>

<style scoped>
.agent-chat {
  display: flex;
  flex: 1;
  min-height: 620px;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.toolbar {
  display: flex;
  justify-content: flex-start;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: rgba(10, 14, 26, 0.75);
}

.message-item {
  max-width: 78%;
  margin-bottom: 16px;
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  text-align: left;
}

.message-item.assistant {
  background: rgba(17, 24, 39, 0.88);
}

.message-item.user {
  margin-left: auto;
  background: rgba(22, 58, 49, 0.9);
  border-color: rgba(70, 180, 150, 0.35);
}

.message-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.reasoning {
  margin-bottom: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-secondary);
  white-space: pre-wrap;
  font-size: 12px;
}

.content {
  white-space: pre-wrap;
  line-height: 1.7;
}

.sender {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
