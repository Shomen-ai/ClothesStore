<template>
  <div class="chart-wrapper">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler } from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler)

const props = defineProps({ points: Array })

const chartData = computed(() => ({
  labels: props.points?.map(p => p.date) || [],
  datasets: [{
    label: 'Выручка',
    data: props.points?.map(p => p.revenue) || [],
    borderColor: '#e11900',
    backgroundColor: 'rgba(225, 25, 0, 0.10)',
    pointBackgroundColor: '#e11900',
    pointBorderColor: '#08080c',
    pointRadius: 3,
    pointHoverRadius: 5,
    borderWidth: 2,
    fill: true,
    tension: 0.35,
  }]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: 'rgba(8,8,12,0.95)',
      borderColor: '#e11900',
      borderWidth: 1,
      titleColor: '#f5f5f7',
      bodyColor: '#c8c8d2',
      padding: 10,
      titleFont: { family: 'JetBrains Mono', size: 11 },
      bodyFont: { family: 'Manrope', size: 13 },
    }
  },
  scales: {
    x: {
      ticks: { color: '#7a7a85', font: { family: 'JetBrains Mono', size: 10 } },
      grid: { color: 'rgba(255,255,255,0.04)' }
    },
    y: {
      ticks: { color: '#7a7a85', font: { family: 'JetBrains Mono', size: 10 } },
      grid: { color: 'rgba(255,255,255,0.04)' }
    }
  }
}
</script>

<style scoped>
.chart-wrapper { height: 280px; }
</style>
