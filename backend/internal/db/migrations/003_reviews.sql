-- Product reviews
CREATE TABLE product_reviews (
  id         BIGSERIAL PRIMARY KEY,
  user_id    BIGINT NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
  product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  rating     SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
  text       TEXT NOT NULL DEFAULT '',
  is_hidden  BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, product_id)
);

CREATE INDEX idx_product_reviews_product
  ON product_reviews(product_id) WHERE is_hidden = FALSE;

CREATE INDEX idx_product_reviews_user
  ON product_reviews(user_id);
