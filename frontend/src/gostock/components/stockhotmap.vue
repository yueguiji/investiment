<script setup lang="ts">
import { onBeforeMount, ref } from 'vue'
import { GetVersionInfo } from "../../../wailsjs/go/main/App";
import EmbeddedUrl from "./EmbeddedUrl.vue";

const messageWallUrl = ref("https://go-stock.sparkmemory.top:16667/go-stock")

onBeforeMount(() => {
  GetVersionInfo().then((res) => {
    if (res.messageWallUrl) {
      messageWallUrl.value = res.messageWallUrl
    }
  })
})
</script>

<template>
  <n-tabs  type="line" animated>
    <n-tab-pane name="选股通" tab="选股通">
      <embedded-url url="https://xuangutong.com.cn"  :height="'calc(100vh - 252px)'"/>
    </n-tab-pane>
    <n-tab-pane name="百度股市通" tab="百度股市通">
      <embedded-url url="https://gushitong.baidu.com"  :height="'calc(100vh - 252px)'"/>
    </n-tab-pane>
    <n-tab-pane name="东财大盘星图" tab="东财大盘星图">
      <embedded-url url="https://quote.eastmoney.com/stockhotmap/"  :height="'calc(100vh - 252px)'"/>
    </n-tab-pane>
    <n-tab-pane name="TopHub" tab="TopHub(今日热榜)">
      <embedded-url url="https://tophub.today/c/finance"  :height="'calc(100vh - 252px)'"/>
    </n-tab-pane>
    <n-tab-pane name="摸鱼" tab="摸鱼">
      <embedded-url url="https://996.ninja/"  :height="'calc(100vh - 252px)'"/>
    </n-tab-pane>
    <n-tab-pane name="财联社-行情数据" tab="财联社-行情数据">
      <embedded-url url="https://www.cls.cn/quotation"  :height="'calc(100vh - 252px)'"/>
    </n-tab-pane>
    <n-tab-pane name="消息墙" tab="消息墙">
      <embedded-url :url="messageWallUrl"  :height="'calc(100vh - 252px)'"/>
    </n-tab-pane>

    <n-tab-pane name="欢迎推荐更多有趣的财经网页" tab="欢迎推荐更多有趣的财经网页">
    </n-tab-pane>
  </n-tabs>
</template>

<style scoped>

</style>
