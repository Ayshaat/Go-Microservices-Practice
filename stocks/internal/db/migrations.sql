CREATE TABLE IF NOT EXISTS stock_items (
    sku       INTEGER PRIMARY KEY,
    name      TEXT NOT NULL,
    type      TEXT NOT NULL,
    price     NUMERIC(10, 2) NOT NULL,
    count     INTEGER NOT NULL DEFAULT 0,
    location  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sku_info (
    sku     INTEGER PRIMARY KEY,
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
