# Delta SP Rate Limiter

## Setup
- Clone this repository
- Install dependencies with the command `go get -u gorm.io/gorm` `gorm.io/driver/sqlite` (or run `go mod tidy`)
- Run the command `go run main.go` to start the server on port 8080

## Usage
### Record Upload Limit
To record an hourly upload limit for a storage provider, send a POST request to the /record-upload-limit endpoint with the following query parameters:

- `address`: the address of the storage provider
- `limit`: the upload limit for the current hour
Example:

```
POST /record-upload-limit?address=example.com&limit=100
```

### Record Upload Size
To record an upload size for a storage provider, send a POST request to the /record-upload-size endpoint with the following query parameters:

- `address`: the address of the storage provider
- `size`: the size of the upload in byte

```
POST /record-upload-size?address=example.com&size=5000000
```

### Check Upload Limit
To check if a storage provider is within its hourly upload limit, send a GET request to the /check-upload-limit endpoint with the following query parameters:

- `address`: the address of the storage provider

```
GET /check-upload-limit?address=example.com
```

If the storage provider is within its limit, the API will return a JSON object with the storage provider data. If the storage provider has exceeded its limit, the API will return a 429 Too Many Requests error.

## Database
This web service uses a SQLite database to store storage provider data. The database is created and initialized automatically when the server is started. The StorageProvider model is defined in the main.go file, and the gorm.AutoMigrate() function is called in the initDB() function to automatically create the necessary tables and columns.

## Hook this to a Delta INstance