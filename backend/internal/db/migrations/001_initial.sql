-- Users
CREATE TABLE users (
  id            BIGSERIAL PRIMARY KEY,
  email         VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  name          VARCHAR(255),
  phone         VARCHAR(20),
  role          VARCHAR(10) DEFAULT 'customer',
  created_at    TIMESTAMP DEFAULT NOW()
);

-- Addresses
CREATE TABLE addresses (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL REFERENCES users(id),
  city        VARCHAR(100) NOT NULL,
  street      VARCHAR(255) NOT NULL,
  house       VARCHAR(20) NOT NULL,
  apartment   VARCHAR(20),
  zip_code    VARCHAR(10),
  is_default  BOOLEAN DEFAULT false
);

-- Categories
CREATE TABLE categories (
  id         BIGSERIAL PRIMARY KEY,
  name       VARCHAR(100) NOT NULL,
  slug       VARCHAR(100) UNIQUE NOT NULL,
  sort_order INTEGER DEFAULT 0
);

-- Products
CREATE TABLE products (
  id          BIGSERIAL PRIMARY KEY,
  category_id BIGINT REFERENCES categories(id),
  name        VARCHAR(255) NOT NULL,
  description TEXT,
  price       NUMERIC(10,2) NOT NULL,
  is_active   BOOLEAN DEFAULT true,
  created_at  TIMESTAMP DEFAULT NOW()
);

-- Product sizes
CREATE TABLE product_sizes (
  id         BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  size       VARCHAR(10) NOT NULL,
  stock_qty  INTEGER DEFAULT 0,
  UNIQUE (product_id, size)
);

-- Product images
CREATE TABLE product_images (
  id         BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  image_path VARCHAR(500) NOT NULL,
  is_primary BOOLEAN DEFAULT false,
  sort_order INTEGER DEFAULT 0
);

-- Promo codes
CREATE TABLE promo_codes (
  id                BIGSERIAL PRIMARY KEY,
  code              VARCHAR(50) UNIQUE NOT NULL,
  discount_type     VARCHAR(10) NOT NULL,
  discount_value    NUMERIC(10,2) NOT NULL,
  max_activations   INTEGER,
  activations_count INTEGER DEFAULT 0,
  expires_at        TIMESTAMP,
  is_active         BOOLEAN DEFAULT true,
  created_at        TIMESTAMP DEFAULT NOW()
);

-- Orders
CREATE TABLE orders (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT NOT NULL REFERENCES users(id),
  address_id      BIGINT NOT NULL REFERENCES addresses(id),
  promo_code_id   BIGINT REFERENCES promo_codes(id),
  status          VARCHAR(20) DEFAULT 'pending',
  total_price     NUMERIC(10,2) NOT NULL,
  discount_amount NUMERIC(10,2) DEFAULT 0,
  created_at      TIMESTAMP DEFAULT NOW()
);

-- Order items
CREATE TABLE order_items (
  id              BIGSERIAL PRIMARY KEY,
  order_id        BIGINT NOT NULL REFERENCES orders(id),
  product_id      BIGINT NOT NULL REFERENCES products(id),
  product_size_id BIGINT NOT NULL REFERENCES product_sizes(id),
  quantity        INTEGER NOT NULL,
  price_at_order  NUMERIC(10,2) NOT NULL
);

-- Wishlist
CREATE TABLE wishlist (
  id         BIGSERIAL PRIMARY KEY,
  user_id    BIGINT NOT NULL REFERENCES users(id),
  product_id BIGINT NOT NULL REFERENCES products(id),
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE (user_id, product_id)
);

-- Indexes
CREATE INDEX idx_addresses_user_id ON addresses(user_id);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_product_sizes_product_id ON product_sizes(product_id);
CREATE INDEX idx_product_images_product_id ON product_images(product_id);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_promo_code_id ON orders(promo_code_id);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_order_items_product_size_id ON order_items(product_size_id);
CREATE INDEX idx_wishlist_user_id ON wishlist(user_id);
CREATE INDEX idx_wishlist_product_id ON wishlist(product_id);
