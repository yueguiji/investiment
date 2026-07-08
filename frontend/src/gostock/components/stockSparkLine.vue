<script setup>
import {nextTick, onBeforeUnmount, onMounted, ref, watchEffect} from "vue";
import * as echarts from 'echarts';
import {GetStockMinutePriceLineData} from "../../../wailsjs/go/main/App"; // 如果您使用多个组件，请将此样式导入放在您的主文件中
const {idSuffix,stockCode,stockName,lastPrice,openPrice,darkTheme} = defineProps({
  idSuffix: {
    type: String,
    default: ""
  },
  stockCode: {
    type: String,
    default: ""
  },
  stockName: {
    type: String,
    default: ""
  },
  lastPrice: {
    type: Number,
    default: 0
  },
  openPrice: {
    type: Number,
    default: 0
  },
  darkTheme: {
    type: Boolean,
    default: true
  },
})

const chartRef = ref(null);
const chart = ref(null)
let disposed = false

function setChartData() {
  if (!chart.value || disposed || !stockCode) return
  //console.log("setChartData")
  GetStockMinutePriceLineData(stockCode, stockName).then(result => {
    if (!chart.value || disposed) return
    //console.log("GetStockMinutePriceLineData",result)
    const priceData = Array.isArray(result?.priceData) ? result.priceData : []
    if (priceData.length === 0) {
      chart.value.clear()
      return
    }
    let category = []
    let price = []
    let min = 0
    let max = 0
    for (let i = 0; i < priceData.length; i++) {
      category.push(priceData[i].time)
      price.push(priceData[i].price)
      if (min === 0 || min > priceData[i].price) {
        min = priceData[i].price
      }
      if (max < priceData[i].price) {
        max = priceData[i].price
      }
    }
    let option = {
      padding: [0, 0, 0, 0],
      grid: {
        top: 0,
        left: 0,
        right: 0,
        bottom: 0
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross',
          label: {
            backgroundColor: '#6a7985'
          }
        }
      },
      xAxis: {
        show: false,
        type: 'category',
        data: category
      },
      yAxis: {
        show: false,
        type: 'value',
        min: (min).toFixed(2),
        max: (max).toFixed(2),
        minInterval: 0.01,
      },
      // visualMap: {
      //   show: false,
      //   type: 'piecewise',
      //   pieces: [
      //     {
      //       min: Number(min),
      //       max: Number(openPrice),
      //       color: 'green'
      //     },
      //     {
      //       min: Number(openPrice),
      //       max: Number(max),
      //       color: 'red'
      //     }
      //   ]
      // },
      series: [
        {
          data: price,
          type: 'line',
          smooth: false,
          stack: '总量',
          showSymbol: false,
          lineStyle: {
            color: lastPrice > openPrice ? 'rgba(245, 0, 0, 1)' : 'rgb(6,251,10)'
          },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
              offset: 0,
              color: lastPrice > openPrice ? 'rgba(245, 0, 0, 1)' : 'rgba(6,251,10, 1)'
            }, {
              offset: 1,
              color: lastPrice > openPrice ? 'rgba(245, 0, 0, 0.25)' : 'rgba(6,251,10, 0.25)'
            }])
          },
        }
      ]
    };
    if (chart.value && !disposed) {
      chart.value.setOption(option);
    }
  }).catch((err) => {
    console.error('GetStockMinutePriceLineData error:', err)
  })
}

onMounted(async () => {
  disposed = false
  await nextTick()
  if (!chartRef.value || disposed) return
  chart.value = echarts.init(chartRef.value);
  setChartData();
})

onBeforeUnmount(() => {
  disposed = true
  if (chart.value) {
    chart.value.dispose()
    chart.value = null
  }
})


watchEffect(() => {
  console.log(stockName,'lastPrice变化为:', lastPrice,lastPrice > openPrice)
  if (!chart.value || disposed) return
  setChartData();
})


</script>
<template>
<div ref="chartRef" style="height: 20px;width: 100%"  :id="'sparkLine'+stockCode+idSuffix">
</div>
</template>
