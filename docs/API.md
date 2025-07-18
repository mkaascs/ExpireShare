# ExpireShare API Documentation

## Authentication
Some endpoints require a password if it was set during file upload.  
Password should be provided in the request body (JSON).

---

## 📂 File Operations

### 1. Get File Information
`GET /file/{alias}`

#### Request Body
```json
{
  "password": "string (optional)"
}
```

#### Response (200 OK)
```json
{
  "downloads_left": 5,
  "expires_in": "2h30m"
}
```

#### Response (404 Not Found)
```json
{
  "errors": ["File not found"]
}
```

#### Response (403 Forbidden)
```json
{
  "errors": ["Invalid password"]
}
```

---

### 2. Delete File
`DELETE /file/{alias}`

#### Request Body
```json
{
  "password": "string (optional)"
}
```

#### Response (200 OK)
```json
{
  "errors": ["File deleted successfully"]
}
```

#### Response (404 Not Found)
```json
{
  "errors": ["File not found"]
}
```

---

## ⬆️ Upload Operations

### 3. Upload File
`POST /upload`

#### Request (multipart/form-data)
| Parameter      | Type     | Required | Description                    |
|---------------|---------|----------|--------------------------------|
| `file`        | file    | Yes      | File to upload                 |
| `max_downloads` | int16   | No       | Maximum downloads (default: 1) |
| `ttl`         | string  | No       | Time-to-live (e.g. "1h")       |
| `password`    | string  | No       | Access password                |

#### Example Request Body:
```json
{
  "max_downloads": 5,
  "ttl": "24h",
  "password": "1234"
}
```

#### Response (201 Created)
```json
{
  "alias": "abc123"
}
```

#### Response (400 Bad Request)
```json
{
  "errors": ["Invalid file format"]
}
```

---

## ⬇️ Download Operations

### 4. Download File
`GET /download/{alias}`

#### Request Body
```json
{
  "password": "string (optional)"
}
```

#### Response (200 OK)
- Returns file binary data with headers:
  ```
  Content-Disposition: attachment; filename="original_name.ext"
  Content-Type: [appropriate MIME-type]
  ```

#### Response (403 Forbidden)
```json
{
  "errors": ["Download limit exceeded"]
}
```

---

## 🚨 Error Responses
All errors return JSON with format:
```json
{
  "errors": [
    "Error1 description",
    "Error2 description"
  ]
}
```

### Status Codes:
| Code | Description               |
|------|---------------------------|
| 400  | Bad Request               |
| 401  | Unauthorized              |
| 403  | Forbidden                 |
| 404  | Not Found                 |
| 500  | Internal Server Error     |

---

## Usage Examples

### cURL (Upload File)
```bash
curl -X POST "https://api.yourservice.com/v1/upload" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@document.pdf" \
  -F 'metadata={"max_downloads":3,"ttl":"12h","password":"1234"};type=application/json'
```
---

📌 **Notes:**
1. All time parameters (`ttl`, `expires_in`) use Go `time.Duration` format (e.g. "2h30m")
2. For protected endpoints, include password field even if empty: `"password": ""`
3. Maximum file size: 500MB (configurable server-side)
4. Passwords are case-sensitive
5. TTL examples: "24h", "1h30m", "30m"