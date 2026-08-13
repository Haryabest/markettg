module github.com/markettg/markettg/services/gateway

go 1.23

require (
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/gofiber/adaptor/v2 v2.2.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/markettg/markettg/packages/go-shared v0.0.0
	github.com/redis/go-redis/v9 v9.7.0
	go.uber.org/zap v1.27.0
)

replace github.com/markettg/markettg/packages/go-shared => ../../packages/go-shared
