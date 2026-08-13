module github.com/markettg/markettg/services/delivery-service

go 1.23

require (
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/markettg/markettg/packages/go-shared v0.0.0
	github.com/redis/go-redis/v9 v9.7.0
	go.uber.org/zap v1.27.0
)

replace github.com/markettg/markettg/packages/go-shared => ../../packages/go-shared
