# Cart Microservice

This microservice manages user cart operations such as adding items, deleting items, listing cart contents, and clearing the cart. It interacts with the **Stocks** microservice to validate SKU and retrieve item details.

---

## Docker Image

- **Image name:** `ayshaat/cart:hw7`
- **Exposed port:** `8090`

---

## Environment Variables

| Variable            | Description                            | Example                   |
|---------------------|----------------------------------------|---------------------------|
| `DB_HOST`           | Hostname or IP of PostgreSQL server    | `cart-db`                |
| `DB_PORT`           | PostgreSQL port                        | `5432`                    |
| `DB_NAME`           | Database name                          | `cart`                    |
| `DB_USER`           | Database username                      | `postgres`                |
| `DB_PASSWORD`       | Database password                      | `lavender@30.06.04`       |
| `HTTP_PORT`         | Port for the Cart HTTP server          | `8090`                    |
| `STOCK_SERVICE_URL` | Base URL of the Stocks service         | `http://stock-service:8080` |

---

## API Endpoints

### Add Item to Cart

- **Method:** POST  
- **URL:** `/cart/item/add`  
- **Content-Type:** `application/json`

**Request body:**
```json
{
  "userId": 123,
  "sku": 1001,
  "count": 2
}
