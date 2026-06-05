// Catalogue price-filter data + helper.
// PRICE_RANGES defines the selectable price buckets. max === 0 is a sentinel meaning
// "no upper bound" (the open-ended top tier). The numeric bounds live here while the
// human labels are kept in ProductFilters.vue (see review note about that duplication).
export const PRICE_RANGES = [
  { id: 'p1', min: 0,    max: 3000 },
  { id: 'p2', min: 3000, max: 5000 },
  { id: 'p3', min: 5000, max: 7000 },
  { id: 'p4', min: 7000, max: 0    },  // open-ended: "from 7000"
]

// Collapse a set of selected range ids into a single {min, max} window for the query.
// Takes the lowest min and highest max across the chosen ranges; if any chosen range is
// open-topped (max === 0), the result is also open-topped (max === 0).
// Returns {min: 0, max: 0} when nothing valid is selected (interpreted as "no price filter").
export function priceBoundsFromIds(ids) {
  if (!ids || !ids.length) return { min: 0, max: 0 }
  const map = Object.fromEntries(PRICE_RANGES.map(r => [r.id, r]))
  const ranges = ids.map(id => map[id]).filter(Boolean)   // drop unknown ids
  if (!ranges.length) return { min: 0, max: 0 }
  const min = Math.min(...ranges.map(r => r.min))
  const hasOpenTop = ranges.some(r => r.max === 0)
  const max = hasOpenTop ? 0 : Math.max(...ranges.map(r => r.max))
  return { min, max }
}
