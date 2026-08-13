module github.com/markettg/markettg/services/user-service

go 1.23

require (
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/markettg/markettg/packages/go-shared v0.0.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.31.0
)

replace github.com/markettg/markettg/packages/go-shared => ../../packages/go-shared
