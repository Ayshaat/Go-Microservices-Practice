# Stocks Microservice

This microservice manages stock items and exposes a REST API for adding stock data.

---

## Docker Image

- **Image name:** `ayshaat/stocks:hw7`
- **Exposed port:** `8080`

---

## Environment Variables

| Variable      | Description                           | Example               |
|---------------|-------------------------------------|-------------------------|
| `DB_HOST`     | Hostname or IP of PostgreSQL server | `postgres`              |
| `DB_PORT`     | PostgreSQL port                     | `5432`                  |
| `DB_NAME`     | Database name                       | `stocks`                |
| `DB_USER`     | Database username                   | `postgres`              |
| `DB_PASSWORD` | Database password                   | `lavender@30.06.04`     |
| `HTTP_PORT`   | Port for the Stocks HTTP server     | `8080`                  |
---

## API Endpoints

### Add Stock Item

- **Method:** POST  
- **URL:** `/stocks/item/add`  
- **Content-Type:** `application/json`

**Request body example:**

```json
{
  "userId": 123,
  "sku": 1001,
  "price": 15.50,
  "count": 2,
  "location": "Helsinki"
}
