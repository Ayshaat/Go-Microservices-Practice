CREATE TABLE IF NOT EXISTS sku_info (
  sku BIGINT PRIMARY KEY,
  name    TEXT NOT NULL,
  type    TEXT NOT NULL
);

INSERT INTO sku_info (sku, name, type) VALUES
(1001, 't-shirt', 'apparel'),
(2020, 'cup', 'accessory'),
(3033, 'book', 'stationery'),
(4044, 'pen', 'stationery'),
(5055, 'powerbank', 'electronics'),
(6066, 'hoody', 'apparel'),
(7077, 'umbrella', 'accessory'),
(8088, 'socks', 'apparel'),
(9099, 'wallet', 'accessory'),
(10101, 'pink-hoody', 'apparel')
ON CONFLICT (sku) DO NOTHING;

CREATE TABLE IF NOT EXISTS cart_items (
  user_id BIGINT NOT NULL,
  sku BIGINT NOT NULL,
  count SMALLINT NOT NULL CHECK (count >= 0),
  PRIMARY KEY (user_id, sku),
  CONSTRAINT fk_cart_sku FOREIGN KEY (sku) REFERENCES sku_info(sku)
);
