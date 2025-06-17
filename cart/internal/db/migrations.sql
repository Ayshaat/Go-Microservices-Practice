CREATE TABLE IF NOT EXISTS cart_items (
  user_id BIGINT NOT NULL,
  sku BIGINT NOT NULL,
  count SMALLINT NOT NULL CHECK (count >= 0),
  PRIMARY KEY (user_id, sku)
  CONSTRAINT fk_cart_sku FOREIGN KEY (sku) REFERENCES sku_info(sku)
);
