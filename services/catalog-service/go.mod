module github.com/markettg/markettg/services/catalog-service

go 1.23

require (
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/markettg/markettg/packages/go-shared v0.0.0
	github.com/minio/minio-go/v7 v7.0.80
	github.com/redis/go-redis/v9 v9.7.0
	go.uber.org/zap v1.27.0
)

replace github.com/markettg/markettg/packages/go-shared => ../../packages/go-shared
