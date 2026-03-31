<template>
  <div class="fade-in">
    <n-page-header
      :title="generatorTitle"
      :subtitle="generatorSubtitle"
    >
      <template #extra>
        <n-space>
          <n-button type="primary" secondary @click="startFreshScript">开始新脚本</n-button>
          <n-button @click="$router.push('/prompts')">提示词模板</n-button>
          <n-button @click="$router.push('/quant/search')">脚本搜索</n-button>
          <n-button @click="$router.push('/quant/templates')">返回程序库</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-grid :cols="2" :x-gap="20" style="margin-top: 20px;">
      <n-gi>
        <div class="platform-card">
          <n-alert v-if="isEditMode" type="info" :show-icon="false" style="margin-bottom: 16px;">
            当前是基于已有脚本做 AI 迭代。现有代码会作为上下文发送给模型，你只需要补充这次想修改的要求。
          </n-alert>

          <n-form label-placement="top">
            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="AI 源">
                  <n-select
                    v-model:value="selectedAiConfigId"
                    :options="aiSourceOptions"
                    placeholder="请选择已保存的 AI 源"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="当前模型">
                  <n-input :value="selectedAiModelName" readonly placeholder="将使用所选 AI 源中的模型" />
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="提示词模板">
                  <n-select
                    v-model:value="selectedPromptTemplateId"
                    :options="promptTemplateOptions"
                    placeholder="请选择系统 Prompt 模板"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="当前模板">
                  <n-input :value="selectedPromptTemplateName" readonly placeholder="未选择时使用默认模板" />
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-collapse arrow-placement="right" style="margin-bottom: 16px;">
              <n-collapse-item title="查看当前模板内容" name="prompt-template-preview">
                <div class="template-preview">
                  {{ selectedPromptTemplateContent || '当前未找到模板内容，将回退到系统内置标准模板。' }}
                </div>
              </n-collapse-item>
            </n-collapse>

            <n-form-item :label="isEditMode ? '本次修改要求' : '策略描述'">
              <n-input
                v-model:value="request.strategyDescription"
                type="textarea"
                :rows="6"
                :placeholder="strategyPlaceholder"
              />
            </n-form-item>

            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="平台适配">
                  <n-input v-model:value="request.brokerPlatform" placeholder="QMT / 掘金 / 聚宽 / 通用" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="脚本分类">
                  <n-select v-model:value="request.scriptCategory" :options="scriptCategoryOptions" />
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="策略类型">
                  <n-select v-model:value="request.strategyType" :options="styleOptions" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="风险等级">
                  <n-select v-model:value="request.riskLevel" :options="riskOptions" />
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="资金规模">
                  <n-input-number v-model:value="request.capital" :min="0" style="width: 100%;" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="关联标的">
                  <n-input v-model:value="request.stockCodes" placeholder="多个代码用逗号分隔，可留空" />
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-form-item label="特征因子">
              <n-select
                v-model:value="factorTags"
                multiple
                filterable
                tag
                :options="factorOptions"
                placeholder="选择因子维度"
              />
            </n-form-item>

            <n-form-item label="适配场景">
              <n-select
                v-model:value="sceneTags"
                multiple
                filterable
                tag
                :options="scenarioOptions"
                placeholder="选择行情场景"
              />
            </n-form-item>

            <n-collapse arrow-placement="right" style="margin-bottom: 16px;">
              <n-collapse-item :title="isEditMode ? '查看当前脚本上下文' : '查看已带入的草稿代码'" name="base-code-preview">
                <div v-if="request.baseCode" class="code-output compact-output">
                  <pre><code>{{ request.baseCode }}</code></pre>
                </div>
                <div v-else class="template-preview">当前还没有带入已有代码，模型会按你的描述新生成脚本。</div>
              </n-collapse-item>
            </n-collapse>
          </n-form>

          <div class="hint-box">
            <div class="hint-title">当前页面说明</div>
            <ul class="hint-list">
              <li>切换菜单后，本页输入和生成结果会自动保留。</li>
              <li>生成过程中离开页面，请求也会继续执行，回来后仍可看到结果。</li>
              <li>已有脚本迭代时，模型会同时参考原脚本和你这次的新增要求。</li>
            </ul>
          </div>

          <n-space vertical style="width: 100%">
            <n-button type="primary" block size="large" :loading="generating" @click="handleGenerate">
              {{ isEditMode ? '基于当前脚本继续生成' : '生成脚本草案' }}
            </n-button>
            <n-button v-if="isEditMode" block @click="startFreshScript">清空上下文，开始新脚本</n-button>
            <n-button block @click="$router.push('/settings')">去设置 AI 源</n-button>
          </n-space>
        </div>
      </n-gi>

      <n-gi>
        <div class="platform-card output-card">
          <div class="output-header">
            <div>
              <div class="output-title">生成结果</div>
              <div class="output-subtitle">生成中的任务会持续执行，离开后返回仍可查看。</div>
            </div>
            <n-space v-if="generatedCode">
              <n-tag type="info" :bordered="false">{{ selectedAiModelName || '未命名模型' }}</n-tag>
              <n-tag type="success" :bordered="false">{{ selectedPromptTemplateName || '默认模板' }}</n-tag>
              <n-button size="small" @click="handleCopy">复制代码</n-button>
              <n-button size="small" type="primary" @click="handleSaveAsNew">另存为新脚本</n-button>
              <n-button v-if="canUpdateExisting" size="small" type="warning" @click="handleUpdateExisting">覆盖更新原脚本</n-button>
            </n-space>
          </div>

          <div v-if="!hasAnyOutput && !generating" class="placeholder-box">
            <div class="placeholder-title">{{ isEditMode ? '先写清楚这次想怎么改，再让 AI 继续迭代' : '先填写策略需求，再生成脚本' }}</div>
            <div class="placeholder-subtitle">生成完成后，这里会展示说明、风险提示、模型输出和可折叠结果区。</div>
          </div>

          <div v-if="generating" class="placeholder-box">
            <n-spin size="large" />
            <div class="placeholder-subtitle" style="margin-top: 16px;">正在调用模型生成量化脚本，切换页面也不会中断。</div>
          </div>

          <n-alert v-if="generationMessage" type="info" :show-icon="false" class="generation-message">
            {{ generationMessage }}
          </n-alert>

          <div v-if="finalModelOutput" class="result-section">
            <div class="section-title">模型最终输出</div>
            <div class="text-output">{{ finalModelOutput }}</div>
          </div>

          <n-space vertical size="large" v-if="generationMeta.explanation || generationMeta.riskWarning || generationMeta.suggestedName">
            <div v-if="generationMeta.suggestedName" class="meta-box">
              <div class="meta-title">建议名称</div>
              <div class="meta-content">{{ generationMeta.suggestedName }}</div>
            </div>
            <div v-if="generationMeta.explanation" class="meta-box">
              <div class="meta-title">策略说明</div>
              <div class="meta-content">{{ generationMeta.explanation }}</div>
            </div>
            <div v-if="generationMeta.riskWarning" class="meta-box">
              <div class="meta-title">风险提示</div>
              <div class="meta-content">{{ generationMeta.riskWarning }}</div>
            </div>
          </n-space>

          <n-collapse v-if="generatedCode || thinkingContent || usedPrompt" class="detail-collapse" arrow-placement="right">
            <n-collapse-item v-if="generatedCode" title="代码结果" name="code-result">
              <div class="code-output">
                <pre><code>{{ generatedCode }}</code></pre>
              </div>
            </n-collapse-item>

            <n-collapse-item v-if="thinkingContent" title="思考内容" name="thinking-content">
              <div class="text-output reasoning-output">{{ thinkingContent }}</div>
            </n-collapse-item>

            <n-collapse-item v-if="usedPrompt" title="本次实际 Prompt" name="used-prompt">
              <div class="text-output">{{ usedPrompt }}</div>
            </n-collapse-item>
          </n-collapse>
        </div>
      </n-gi>
    </n-grid>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { useRoute, useRouter } from 'vue-router'
import { joinTagArray, splitTagString } from '../../utils/quant'

const DEFAULT_PROMPT_TEMPLATE_NAME = '量化脚本生成-标准模板'
const SYSTEM_PROMPT_TYPE = '模型系统Prompt'
const STORAGE_KEY = 'investment.quant.ai-generator.state.v3'
const RUNTIME_KEY = '__investmentQuantAiGeneratorRuntime'
const AI_DRAFT_KEY = 'investment.quant.ai-generator.draft'

const route = useRoute()
const router = useRouter()
const message = useMessage()

function createDefaultMeta() {
  return {
    explanation: '',
    riskWarning: '',
    suggestedName: '',
  }
}

function createDefaultRequest() {
  return {
    strategyDescription: '',
    brokerPlatform: 'QMT',
    strategyType: 'multi-factor',
    scriptCategory: 'strategy-main',
    stockCodes: '',
    riskLevel: 'medium',
    capital: 100000,
    factorTags: '',
    sceneTags: '',
    aiModel: '',
    promptTemplateId: 0,
    baseCode: '',
    existingScriptName: '',
    existingDescription: '',
  }
}

function getRuntimeStore() {
  if (!window[RUNTIME_KEY]) {
    window[RUNTIME_KEY] = {
      lastSnapshot: null,
      lastUpdatedAt: 0,
      activeRunId: null,
    }
  }
  return window[RUNTIME_KEY]
}

const generating = ref(false)
const generatedCode = ref('')
const generationMessage = ref('')
const usedPrompt = ref('')
const rawModelOutput = ref('')
const finalModelOutput = ref('')
const thinkingContent = ref('')
const taxonomy = ref([])
const factorTags = ref([])
const sceneTags = ref([])
const aiConfigs = ref([])
const promptTemplates = ref([])
const selectedAiConfigId = ref(null)
const selectedPromptTemplateId = ref(null)
const generationMeta = ref(createDefaultMeta())
const request = ref(createDefaultRequest())
const sourceTemplateId = ref(0)
const mode = ref('create')

let syncTimer = null
let suppressPersist = false

const isEditMode = computed(() => mode.value === 'edit' || Number(route.query.sourceTemplateId || 0) > 0)
const generatorTitle = computed(() => (isEditMode.value ? 'AI 迭代量化脚本' : 'AI 量化脚本生成'))
const generatorSubtitle = computed(() => (isEditMode.value ? '基于已有脚本继续修改、增强和重构' : '支持模板增强、结果持久化和离页继续生成'))
const strategyPlaceholder = computed(() =>
  isEditMode.value
    ? '例如：保留现有多因子框架，把止损改成 ATR 动态止损，并增加仓位上限、日志输出和回测摘要。'
    : '例如：生成一个适合快速轮动市场的宽基 ETF 网格策略，结合资金流、情绪强度、主题共振和波动率动态调整网格密度与仓位。'
)

const hasAnyOutput = computed(() =>
  Boolean(
    generatedCode.value ||
      finalModelOutput.value ||
      thinkingContent.value ||
      usedPrompt.value ||
      generationMeta.value.explanation ||
      generationMeta.value.riskWarning ||
      generationMeta.value.suggestedName
  )
)

const canUpdateExisting = computed(() => isEditMode.value && sourceTemplateId.value > 0 && Boolean(generatedCode.value))
const styleOptions = computed(() => getOptions('style'))
const factorOptions = computed(() => getOptions('factor'))
const scenarioOptions = computed(() => getOptions('scenario'))
const scriptCategoryOptions = computed(() => getOptions('script-category'))

const aiSourceOptions = computed(() =>
  aiConfigs.value.map((item) => ({
    label: `${item.name || '未命名源'} / ${item.modelName || '未配置模型'}`,
    value: item.ID,
  }))
)

const promptTemplateOptions = computed(() =>
  promptTemplates.value.map((item) => ({
    label: `${item.name} / ${item.type}`,
    value: item.ID,
  }))
)

const selectedAiConfig = computed(() =>
  aiConfigs.value.find((item) => item.ID === selectedAiConfigId.value) || null
)

const selectedAiModelName = computed(() => selectedAiConfig.value?.modelName || '')

const selectedPromptTemplate = computed(() =>
  promptTemplates.value.find((item) => item.ID === selectedPromptTemplateId.value) || null
)

const selectedPromptTemplateName = computed(() =>
  selectedPromptTemplate.value?.name || DEFAULT_PROMPT_TEMPLATE_NAME
)

const selectedPromptTemplateContent = computed(() =>
  selectedPromptTemplate.value?.content || ''
)

const riskOptions = [
  { label: '低风险', value: 'low' },
  { label: '中风险', value: 'medium' },
  { label: '高风险', value: 'high' },
]

function getOptions(key) {
  const group = taxonomy.value.find((item) => item.key === key)
  return (group?.options || []).map((item) => ({ label: item.label, value: item.value }))
}

function buildFallbackCode() {
  return `import logging
from dataclasses import dataclass
from typing import List

import numpy as np
import pandas as pd

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

@dataclass
class Config:
    capital: float = ${request.value.capital}
    risk_level: str = "${request.value.riskLevel}"
    symbols: List[str] = "${request.value.stockCodes}".split(",") if "${request.value.stockCodes}" else ["510300", "510500"]


def create_sample_data(rows: int = 120) -> pd.DataFrame:
    rng = np.random.default_rng(42)
    close = np.cumsum(rng.normal(0.1, 1.0, rows)) + 100
    volume = rng.integers(1000, 5000, rows)
    return pd.DataFrame({
        "close": close,
        "open": close - rng.normal(0, 0.5, rows),
        "high": close + rng.normal(0.5, 0.3, rows),
        "low": close - rng.normal(0.5, 0.3, rows),
        "volume": volume,
    })


def build_features(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()
    df["ma5"] = df["close"].rolling(5).mean()
    df["ma20"] = df["close"].rolling(20).mean()
    df["ret5"] = df["close"].pct_change(5)
    return df.dropna()


def generate_signal(df: pd.DataFrame) -> str:
    latest = df.iloc[-1]
    if latest["ma5"] > latest["ma20"] and latest["ret5"] > 0:
        return "buy"
    if latest["ma5"] < latest["ma20"]:
        return "sell"
    return "hold"


def run_strategy() -> None:
    data = build_features(create_sample_data())
    signal = generate_signal(data)
    logging.info("strategy_description=%s", "${request.value.strategyDescription}")
    logging.info("signal=%s", signal)
    print({"signal": signal, "rows": len(data)})


if __name__ == "__main__":
    run_strategy()
`
}

function splitModelOutput(content) {
  const raw = (content || '').trim()
  if (!raw) {
    return { thinking: '', final: '' }
  }

  const thinkMatch = raw.match(/<think>([\s\S]*?)<\/think>/i)
  const thinking = thinkMatch ? thinkMatch[1].trim() : ''
  const final = thinkMatch ? raw.replace(thinkMatch[0], '').trim() : raw
  return { thinking, final }
}

function applySnapshot(snapshot) {
  if (!snapshot) {
    return
  }

  suppressPersist = true
  request.value = { ...createDefaultRequest(), ...(snapshot.request || {}) }
  factorTags.value = [...(snapshot.factorTags || [])]
  sceneTags.value = [...(snapshot.sceneTags || [])]
  selectedAiConfigId.value = snapshot.selectedAiConfigId ?? null
  selectedPromptTemplateId.value = snapshot.selectedPromptTemplateId ?? null
  generating.value = Boolean(snapshot.generating)
  generatedCode.value = snapshot.generatedCode || ''
  generationMessage.value = snapshot.generationMessage || ''
  usedPrompt.value = snapshot.usedPrompt || ''
  rawModelOutput.value = snapshot.rawModelOutput || ''
  finalModelOutput.value = snapshot.finalModelOutput || ''
  thinkingContent.value = snapshot.thinkingContent || ''
  generationMeta.value = { ...createDefaultMeta(), ...(snapshot.generationMeta || {}) }
  sourceTemplateId.value = Number(snapshot.sourceTemplateId || 0)
  mode.value = snapshot.mode || 'create'
  suppressPersist = false
}

function buildSnapshot() {
  return {
    request: { ...request.value },
    factorTags: [...factorTags.value],
    sceneTags: [...sceneTags.value],
    selectedAiConfigId: selectedAiConfigId.value,
    selectedPromptTemplateId: selectedPromptTemplateId.value,
    generating: generating.value,
    generatedCode: generatedCode.value,
    generationMessage: generationMessage.value,
    usedPrompt: usedPrompt.value,
    rawModelOutput: rawModelOutput.value,
    finalModelOutput: finalModelOutput.value,
    thinkingContent: thinkingContent.value,
    generationMeta: { ...generationMeta.value },
    sourceTemplateId: sourceTemplateId.value,
    mode: mode.value,
  }
}

function persistState() {
  if (suppressPersist) {
    return
  }
  const runtime = getRuntimeStore()
  const snapshot = buildSnapshot()
  runtime.lastSnapshot = snapshot
  runtime.lastUpdatedAt = Date.now()
  localStorage.setItem(STORAGE_KEY, JSON.stringify(snapshot))
}

function restoreState() {
  const runtime = getRuntimeStore()
  if (runtime.lastSnapshot) {
    applySnapshot(runtime.lastSnapshot)
    return
  }

  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return
  }

  try {
    applySnapshot(JSON.parse(raw))
  } catch (error) {
    console.warn('Failed to restore AI generator state', error)
  }
}

function applyDraft(draft) {
  if (!draft) {
    return
  }

  suppressPersist = true
  mode.value = draft.mode || 'create'
  sourceTemplateId.value = Number(draft.sourceTemplateId || 0)
  request.value = {
    ...request.value,
    strategyDescription: draft.strategyDescription || request.value.strategyDescription,
    brokerPlatform: draft.brokerPlatform || request.value.brokerPlatform,
    strategyType: draft.strategyType || request.value.strategyType,
    scriptCategory: draft.scriptCategory || request.value.scriptCategory,
    stockCodes: draft.stockCodes || request.value.stockCodes,
    baseCode: draft.baseCode || request.value.baseCode,
    existingScriptName: draft.sourceTemplateName || request.value.existingScriptName,
    existingDescription: draft.sourceDescription || request.value.existingDescription,
  }
  factorTags.value = splitTagString(draft.factorTags)
  sceneTags.value = splitTagString(draft.sceneTags)
  suppressPersist = false
}

function loadDraftFromSession() {
  const raw = sessionStorage.getItem(AI_DRAFT_KEY)
  if (!raw) {
    mode.value = route.query.mode === 'edit' ? 'edit' : 'create'
    sourceTemplateId.value = Number(route.query.sourceTemplateId || 0)
    return
  }

  try {
    applyDraft(JSON.parse(raw))
  } catch (error) {
    console.warn('Failed to parse AI draft', error)
  }
}

function startFreshScript() {
  const keepAiConfigId = selectedAiConfigId.value
  const keepPromptTemplateId = selectedPromptTemplateId.value

  suppressPersist = true
  mode.value = 'create'
  sourceTemplateId.value = 0
  request.value = createDefaultRequest()
  factorTags.value = []
  sceneTags.value = []
  generatedCode.value = ''
  generationMessage.value = ''
  usedPrompt.value = ''
  rawModelOutput.value = ''
  finalModelOutput.value = ''
  thinkingContent.value = ''
  generationMeta.value = createDefaultMeta()
  selectedAiConfigId.value = keepAiConfigId
  selectedPromptTemplateId.value = keepPromptTemplateId
  suppressPersist = false

  sessionStorage.removeItem(AI_DRAFT_KEY)
  const runtime = getRuntimeStore()
  runtime.activeRunId = null
  persistState()
  router.replace('/quant/ai-generate?mode=create')
  message.success('已切换为新脚本模式')
}

function startRuntimeSync() {
  const runtime = getRuntimeStore()
  let lastSeen = runtime.lastUpdatedAt
  syncTimer = window.setInterval(() => {
    if (runtime.lastUpdatedAt > lastSeen && runtime.lastSnapshot) {
      lastSeen = runtime.lastUpdatedAt
      applySnapshot(runtime.lastSnapshot)
    }
  }, 800)
}

function stopRuntimeSync() {
  if (syncTimer) {
    window.clearInterval(syncTimer)
    syncTimer = null
  }
}

async function loadAiConfigs() {
  if (!window.go?.main?.App?.GetAiConfigs) {
    return
  }
  aiConfigs.value = (await window.go.main.App.GetAiConfigs()) || []
  if (!selectedAiConfigId.value && aiConfigs.value.length > 0) {
    selectedAiConfigId.value = aiConfigs.value[0].ID
  }
}

async function loadPromptTemplates() {
  if (!window.go?.main?.App?.GetPromptTemplates) {
    return
  }
  promptTemplates.value = (await window.go.main.App.GetPromptTemplates('', SYSTEM_PROMPT_TYPE)) || []
  const selectedExists = promptTemplates.value.some((item) => item.ID === selectedPromptTemplateId.value)
  if (selectedExists) {
    return
  }
  const defaultTemplate = promptTemplates.value.find((item) => item.name === DEFAULT_PROMPT_TEMPLATE_NAME)
  if (defaultTemplate) {
    selectedPromptTemplateId.value = defaultTemplate.ID
    return
  }
  if (!selectedPromptTemplateId.value && promptTemplates.value.length > 0) {
    selectedPromptTemplateId.value = promptTemplates.value[0].ID
  }
}

function resetGenerationState() {
  generatedCode.value = ''
  generationMessage.value = ''
  usedPrompt.value = ''
  rawModelOutput.value = ''
  finalModelOutput.value = ''
  thinkingContent.value = ''
  generationMeta.value = createDefaultMeta()
}

function applyGenerationSuccess(result) {
  const outputParts = splitModelOutput(result.rawContent || '')
  generatedCode.value = result.code || buildFallbackCode()
  rawModelOutput.value = result.rawContent || ''
  finalModelOutput.value = outputParts.final
  thinkingContent.value = outputParts.thinking
  usedPrompt.value = result.prompt || ''
  generationMessage.value = `已使用 ${result.model || selectedAiModelName.value} + ${selectedPromptTemplateName.value} 生成脚本`
  generationMeta.value = {
    explanation: result.explanation || '',
    riskWarning: result.riskWarning || '',
    suggestedName: sanitizeScriptName(result.suggestedName || '', request.value.strategyDescription, sourceTemplateId.value),
  }
}

function applyGenerationFailure(result, fallbackMessage) {
  generatedCode.value = buildFallbackCode()
  usedPrompt.value = result?.prompt || ''
  generationMessage.value = result?.message || fallbackMessage
}

function sanitizeScriptName(rawName, fallbackText = '', sourceId = 0) {
  const text = String(rawName || '')
    .replace(/\r/g, ' ')
    .replace(/\n/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
  const lower = text.toLowerCase()
  const broken =
    !text ||
    text.length > 48 ||
    lower.includes('```') ||
    lower.includes('<think>') ||
    lower.includes('import ') ||
    lower.includes('def ') ||
    lower.includes('class ') ||
    lower.includes('#!/usr/bin') ||
    lower.includes('请根据') ||
    lower.includes('用户需求') ||
    lower.includes('策略说明') ||
    lower.includes('风险提示') ||
    lower.includes('建议名称')

  if (!broken) {
    return text.slice(0, 32)
  }

  const fallback = String(fallbackText || '')
    .replace(/\r/g, ' ')
    .replace(/\n/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
  if (fallback) {
    return fallback.slice(0, 24)
  }
  if (sourceId) {
    return `量化脚本_${sourceId}`
  }
  return `AI脚本_${new Date().toLocaleDateString('zh-CN')}`
}

async function handleGenerate() {
  if (!request.value.strategyDescription.trim()) {
    message.warning(isEditMode.value ? '请先填写这次想修改的要求' : '请先填写策略描述')
    return
  }
  if (!selectedAiConfigId.value) {
    message.warning('请先在设置里配置并选择一个 AI 源')
    return
  }

  const runtime = getRuntimeStore()
  const runId = Date.now()
  runtime.activeRunId = runId

  generating.value = true
  resetGenerationState()
  request.value.factorTags = joinTagArray(factorTags.value)
  request.value.sceneTags = joinTagArray(sceneTags.value)
  request.value.aiModel = selectedAiModelName.value
  request.value.promptTemplateId = selectedPromptTemplateId.value || 0
  persistState()

  try {
    if (window.go?.main?.App?.GenerateQuantCode) {
      const payload = { ...request.value }
      const result = await window.go.main.App.GenerateQuantCode(payload, selectedAiConfigId.value)
      if (runtime.activeRunId !== runId) {
        return
      }

      if (result?.success) {
        applyGenerationSuccess(result)
      } else {
        applyGenerationFailure(result, '模型调用失败，已回退到本地草稿')
        message.error(generationMessage.value)
      }
      return
    }

    if (window.go?.main?.App?.BuildQuantPrompt) {
      const prompt = await window.go.main.App.BuildQuantPrompt({ ...request.value })
      if (runtime.activeRunId !== runId) {
        return
      }
      generatedCode.value = buildFallbackCode()
      usedPrompt.value = prompt || ''
      generationMessage.value = '当前环境未接入模型调用，已展示实际 Prompt 和本地草稿'
      return
    }

    generatedCode.value = buildFallbackCode()
  } catch (error) {
    if (runtime.activeRunId !== runId) {
      return
    }
    console.error(error)
    applyGenerationFailure(null, error?.message || '模型调用失败，已回退到本地草稿')
    message.error(generationMessage.value)
  } finally {
    if (runtime.activeRunId === runId) {
      generating.value = false
      runtime.activeRunId = null
      persistState()
    }
  }
}

function buildSavePayload() {
  request.value.factorTags = joinTagArray(factorTags.value)
  request.value.sceneTags = joinTagArray(sceneTags.value)
  request.value.aiModel = selectedAiModelName.value
  request.value.promptTemplateId = selectedPromptTemplateId.value || 0

  return {
    name: generationMeta.value.suggestedName || `AI脚本_${new Date().toLocaleDateString('zh-CN')}`,
    description: request.value.existingDescription || request.value.strategyDescription,
    code: generatedCode.value,
    brokerPlatform: request.value.brokerPlatform,
    strategyType: request.value.strategyType,
    scriptCategory: request.value.scriptCategory,
    factorTags: request.value.factorTags,
    scenarioTags: request.value.sceneTags,
    linkedStocks: request.value.stockCodes,
    isAiGenerated: true,
    aiPrompt: `${selectedPromptTemplateName.value}\n\n${usedPrompt.value || request.value.strategyDescription}`,
    aiModel: request.value.aiModel,
    status: 'draft',
  }
}

async function handleSaveAsNew() {
  if (!generatedCode.value || !window.go?.main?.App?.CreateQuantTemplate) {
    return
  }

  const created = await window.go.main.App.CreateQuantTemplate(buildSavePayload())
  if (created?.ID) {
    sourceTemplateId.value = created.ID
  }
  message.success('已保存到量化程序库')
}

async function handleUpdateExisting() {
  if (!canUpdateExisting.value || !window.go?.main?.App?.GetQuantTemplate || !window.go?.main?.App?.UpdateQuantTemplate) {
    return
  }

  const current = await window.go.main.App.GetQuantTemplate(sourceTemplateId.value)
  if (!current) {
    message.error('没有找到要更新的原脚本')
    return
  }

  const payload = {
    ...current,
    ...buildSavePayload(),
    ID: current.ID,
    name: request.value.existingScriptName || current.name,
    description: request.value.existingDescription || current.description || request.value.strategyDescription,
    status: current.status || 'draft',
  }

  const updated = await window.go.main.App.UpdateQuantTemplate(payload)
  if (!updated?.ID) {
    message.error('覆盖更新失败')
    return
  }

  message.success('原脚本已用 AI 结果更新')
  router.push(`/quant/editor/${updated.ID}`)
}

async function handleCopy() {
  await navigator.clipboard.writeText(generatedCode.value)
  message.success('已复制代码到剪贴板')
}

watch(
  [
    request,
    factorTags,
    sceneTags,
    selectedAiConfigId,
    selectedPromptTemplateId,
    generating,
    generatedCode,
    generationMessage,
    usedPrompt,
    rawModelOutput,
    finalModelOutput,
    thinkingContent,
    generationMeta,
    sourceTemplateId,
    mode,
  ],
  () => {
    persistState()
  },
  { deep: true }
)

onMounted(async () => {
  restoreState()
  loadDraftFromSession()
  if (window.go?.main?.App?.GetQuantTagTaxonomy) {
    taxonomy.value = (await window.go.main.App.GetQuantTagTaxonomy()) || []
  }
  await loadAiConfigs()
  await loadPromptTemplates()
  persistState()
  startRuntimeSync()
})

onUnmounted(() => {
  persistState()
  stopRuntimeSync()
})
</script>

<style scoped>
.hint-box {
  margin: 16px 0;
  padding: 16px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border-color);
}

.hint-title {
  font-weight: 700;
  margin-bottom: 8px;
}

.hint-list {
  margin: 0 0 0 18px;
  color: var(--text-secondary);
  line-height: 1.7;
}

.template-preview {
  white-space: pre-wrap;
  line-height: 1.7;
  color: var(--text-secondary);
  max-height: 260px;
  overflow: auto;
  padding: 8px 2px;
}

.output-card {
  min-height: 620px;
  display: flex;
  flex-direction: column;
}

.output-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 16px;
}

.output-title {
  font-weight: 700;
}

.output-subtitle,
.placeholder-subtitle {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-muted);
}

.placeholder-box {
  min-height: 420px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.placeholder-title {
  font-size: 18px;
  font-weight: 600;
}

.generation-message {
  margin-bottom: 16px;
}

.detail-collapse {
  margin-top: auto;
  padding-top: 20px;
}

.result-section {
  margin-bottom: 20px;
}

.section-title {
  font-weight: 700;
  margin-bottom: 10px;
}

.code-output,
.text-output {
  overflow: auto;
  max-height: 420px;
  padding: 16px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border-color);
}

.compact-output {
  max-height: 260px;
}

.code-output pre,
.text-output {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.7;
}

.reasoning-output {
  background: rgba(255, 196, 87, 0.06);
  border-color: rgba(255, 196, 87, 0.24);
}

.meta-box {
  padding: 14px 16px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border-color);
}

.meta-title {
  font-weight: 700;
  margin-bottom: 6px;
}

.meta-content {
  color: var(--text-secondary);
  line-height: 1.7;
  white-space: pre-wrap;
}
</style>
