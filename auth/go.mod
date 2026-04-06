module github.com/EugeneNail/vox/auth

go 1.26.1

require (
	github.com/EugeneNail/vox/lib-common v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/samborkent/uuidv7 v0.0.0-20231110121620-f2e19d87e48b
	golang.org/x/crypto v0.49.0
)

replace github.com/EugeneNail/vox/lib-common => ../lib-common
