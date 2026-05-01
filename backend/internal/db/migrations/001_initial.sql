-- Sequences
CREATE SEQUENCE users_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE addresses_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE categories_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE products_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE product_sizes_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE product_images_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE promo_codes_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE orders_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE order_items_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE wishlist_seq START WITH 1 INCREMENT BY 1;

-- Users
CREATE TABLE users (
  id            NUMBER DEFAULT users_seq.NEXTVAL PRIMARY KEY,
  email         VARCHAR2(255) UNIQUE NOT NULL,
  password_hash VARCHAR2(255) NOT NULL,
  name          VARCHAR2(255),
  phone         VARCHAR2(20),
  role          VARCHAR2(10) DEFAULT 'customer',
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Addresses
CREATE TABLE addresses (
  id          NUMBER DEFAULT addresses_seq.NEXTVAL PRIMARY KEY,
  user_id     NUMBER NOT NULL REFERENCES users(id),
  city        VARCHAR2(100) NOT NULL,
  street      VARCHAR2(255) NOT NULL,
  house       VARCHAR2(20) NOT NULL,
  apartment   VARCHAR2(20),
  zip_code    VARCHAR2(10),
  is_default  NUMBER(1) DEFAULT 0
);

-- Categories
CREATE TABLE categories (
  id         NUMBER DEFAULT categories_seq.NEXTVAL PRIMARY KEY,
  name       VARCHAR2(100) NOT NULL,
  slug       VARCHAR2(100) UNIQUE NOT NULL,
  sort_order NUMBER DEFAULT 0
);

-- Products
CREATE TABLE products (
  id          NUMBER DEFAULT products_seq.NEXTVAL PRIMARY KEY,
  category_id NUMBER REFERENCES categories(id),
  name        VARCHAR2(255) NOT NULL,
  description CLOB,
  price       NUMBER(10,2) NOT NULL,
  is_active   NUMBER(1) DEFAULT 1,
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Product sizes
CREATE TABLE product_sizes (
  id         NUMBER DEFAULT product_sizes_seq.NEXTVAL PRIMARY KEY,
  product_id NUMBER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  size       VARCHAR2(10) NOT NULL,
  stock_qty  NUMBER DEFAULT 0
);

-- Product images
CREATE TABLE product_images (
  id         NUMBER DEFAULT product_images_seq.NEXTVAL PRIMARY KEY,
  product_id NUMBER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  image_path VARCHAR2(500) NOT NULL,
  is_primary NUMBER(1) DEFAULT 0,
  sort_order NUMBER DEFAULT 0
);

-- Promo codes
CREATE TABLE promo_codes (
  id                NUMBER DEFAULT promo_codes_seq.NEXTVAL PRIMARY KEY,
  code              VARCHAR2(50) UNIQUE NOT NULL,
  discount_type     VARCHAR2(10) NOT NULL,
  discount_value    NUMBER(10,2) NOT NULL,
  max_activations   NUMBER,
  activations_count NUMBER DEFAULT 0,
  expires_at        TIMESTAMP,
  is_active         NUMBER(1) DEFAULT 1,
  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders
CREATE TABLE orders (
  id              NUMBER DEFAULT orders_seq.NEXTVAL PRIMARY KEY,
  user_id         NUMBER NOT NULL REFERENCES users(id),
  address_id      NUMBER NOT NULL REFERENCES addresses(id),
  promo_code_id   NUMBER REFERENCES promo_codes(id),
  status          VARCHAR2(20) DEFAULT 'pending',
  total_price     NUMBER(10,2) NOT NULL,
  discount_amount NUMBER(10,2) DEFAULT 0,
  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order items
CREATE TABLE order_items (
  id              NUMBER DEFAULT order_items_seq.NEXTVAL PRIMARY KEY,
  order_id        NUMBER NOT NULL REFERENCES orders(id),
  product_id      NUMBER NOT NULL REFERENCES products(id),
  product_size_id NUMBER NOT NULL REFERENCES product_sizes(id),
  quantity        NUMBER NOT NULL,
  price_at_order  NUMBER(10,2) NOT NULL
);

-- Wishlist
CREATE TABLE wishlist (
  id         NUMBER DEFAULT wishlist_seq.NEXTVAL PRIMARY KEY,
  user_id    NUMBER NOT NULL REFERENCES users(id),
  product_id NUMBER NOT NULL REFERENCES products(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT wishlist_unique UNIQUE (user_id, product_id)
);
