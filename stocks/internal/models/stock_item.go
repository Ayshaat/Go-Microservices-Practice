package models

type StockItem struct {
	SKU      uint32 `json:"sku"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Price    uint32 `json:"price"`
	Count    uint16 `json:"count"`
	Location string `json:"location"`
}

type SKUInfo struct {
	Name string
	Type string
}

var SKUDetails = map[uint32]SKUInfo{
	1001:  {"t-shirt", "apparel"},
	2020:  {"cup", "accessory"},
	3033:  {"book", "stationery"},
	4044:  {"pen", "stationery"},
	5055:  {"powerbank", "electronics"},
	6066:  {"hoody", "apparel"},
	7077:  {"umbrella", "accessory"},
	8088:  {"socks", "apparel"},
	9099:  {"wallet", "accessory"},
	10101: {"pink-hoody", "apparel"},
}
