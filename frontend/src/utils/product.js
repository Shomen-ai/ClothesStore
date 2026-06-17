// Shared helpers for product presentation.

// Display title for a product: "<тип> <название>" (e.g. "Футболка Valor").
// `type_name` is the singular form of the product's category, returned by the API.
// Falls back to the bare name when no type is available.
export function productTitle(p) {
  if (!p) return ''
  return p.type_name ? `${p.type_name} ${p.name}` : (p.name || '')
}
