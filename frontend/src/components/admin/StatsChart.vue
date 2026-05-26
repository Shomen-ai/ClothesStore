<template>
  <div class="chart-wrapper">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler } from 'chart.js'
import { useThemeStore } from '@/stores/theme.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler)

const props = defineProps({ points: Array })
const theme = useThemeStore()

const palette = computed(() => {
  const dark = theme.theme === 'dark'
  return {
    accent: '#e11900',
    accentSoft: dark ? 'rgba(225, 25, 0, 0.16)' : 'rgba(225, 25, 0, 0.12)',
    pointBorder: dark ? '#08080c' : '#f6f3ee',
    tooltipBg: dark ? 'rgba(8,8,12,0.95)' : 'rgba(246,243,238,0.96)',
    title: dark ? '#f5f5f7' : '#14141a',
    body: dark ? '#c8c8d2' : '#34343e',
    tick: dark ? '#7a7a85' : '#6c6c78',
    grid: dark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.06)',
  }
})

const chartData = computed(() => ({
  labels: props.points?.map(p => p.date) || [],
  datasets: [{
    label: 'Выручка',
    data: props.points?.map(p => p.revenue) || [],
    borderColor: palette.value.accent,
    backgroundColor: palette.value.accentSoft,
    pointBackgroundColor: palette.value.accent,
    pointBorderColor: palette.value.pointBorder,
    pointRadius: 3,
    pointHoverRadius: 5,
    borderWidth: 2,
    fill: true,
    tension: 0.35,
  }]
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: palette.value.tooltipBg,
      borderColor: palette.value.accent,
      borderWidth: 1,
      titleColor: palette.value.title,
      bodyColor: palette.value.body,
      padding: 10,
      titleFont: { family: 'JetBrains Mono', size: 11 },
      bodyFont: { family: 'Manrope', size: 13 },
    }
  },
  scales: {
    x: {
      ticks: { color: palette.value.tick, font: { family: 'JetBrains Mono', size: 10 } },
      grid: { color: palette.value.grid }
    },
    y: {
      ticks: { color: palette.value.tick, font: { family: 'JetBrains Mono', size: 10 } },
      grid: { color: palette.value.grid }
    }
  }
}))
</script>

<style scoped>
.chart-wrapper { height: 280px; }
</style>
