ALTER TABLE products
  ADD COLUMN IF NOT EXISTS is_on_sale BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_products_is_on_sale ON products(is_on_sale) WHERE is_on_sale = true;
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
