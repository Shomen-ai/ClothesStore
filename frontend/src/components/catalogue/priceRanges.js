export const PRICE_RANGES = [
  { id: 'p1', min: 0,    max: 3000 },
  { id: 'p2', min: 3000, max: 5000 },
  { id: 'p3', min: 5000, max: 7000 },
  { id: 'p4', min: 7000, max: 0    },
]

export function priceBoundsFromIds(ids) {
  if (!ids || !ids.length) return { min: 0, max: 0 }
  const map = Object.fromEntries(PRICE_RANGES.map(r => [r.id, r]))
  const ranges = ids.map(id => map[id]).filter(Boolean)
  if (!ranges.length) return { min: 0, max: 0 }
  const min = Math.min(...ranges.map(r => r.min))
  const hasOpenTop = ranges.some(r => r.max === 0)
  const max = hasOpenTop ? 0 : Math.max(...ranges.map(r => r.max))
  return { min, max }
}
