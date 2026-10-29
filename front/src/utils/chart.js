export function buildLineChart(seriesMap, { width = 1000, height = 460, padding = 52, ticks = 5 } = {}) {
  const normalized = Object.fromEntries(Object.entries(seriesMap).map(([key, values]) => [key, (values || []).map(Number).filter(Number.isFinite)]))
  const allValues = Object.values(normalized).flat()
  const rawMin = Math.min(...allValues, 0)
  const rawMax = Math.max(...allValues, 0)
  const rawSpan = rawMax - rawMin
  const margin = rawSpan ? rawSpan * 0.12 : Math.max(Math.abs(rawMax) * 0.1, 1)
  const min = rawMin - margin
  const max = rawMax + margin
  const span = max - min || 1
  const innerWidth = width - padding * 2
  const innerHeight = height - padding * 2
  const maxLength = Math.max(...Object.values(normalized).map((values) => values.length), 0)
  const x = (index) => padding + (maxLength <= 1 ? innerWidth / 2 : index * innerWidth / (maxLength - 1))
  const y = (value) => padding + ((max - value) / span) * innerHeight
  const series = Object.fromEntries(Object.entries(normalized).map(([key, values]) => {
    const coords = values.map((value, index) => ({ x: x(index), y: y(value), value }))
    return [key, { points: coords.map((point) => `${point.x},${point.y}`).join(' '), last: coords.at(-1) }]
  }))
  const yTicks = Array.from({ length: ticks }, (_, index) => {
    const value = max - index * span / (ticks - 1)
    return { value, y: y(value) }
  })
  const historicalLength = Math.max(...Object.entries(normalized).filter(([key]) => key !== 'riskFree').map(([, values]) => values.length), 0)
  const historyEndX = historicalLength > 0 ? x(historicalLength - 1) : null
  return { series, yTicks, zeroY: y(0), historyEndX, startLabel: 'Начало', endLabel: maxLength > historicalLength ? `Прогноз +${maxLength - historicalLength} дней` : (maxLength > 1 ? `Точка ${maxLength}` : 'Сейчас') }
}
