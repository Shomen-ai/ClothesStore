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
    borderColor: '#cc2200',
    backgroundColor: 'rgba(204,34,0,0.1)',
    fill: true,
    tension: 0.3,
  }]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    x: { ticks: { color: '#888' }, grid: { color: '#2a2a2a' } },
    y: { ticks: { color: '#888' }, grid: { color: '#2a2a2a' } }
  }
}
</script>

<style scoped>
.chart-wrapper { height: 280px; }
</style>
