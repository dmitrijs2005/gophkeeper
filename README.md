# GophKeeper

Лёгкий менеджер секретов с синхронизацией через gRPC. Сервер на Go, хранение в Postgres, бинарные файлы — в S3-совместимом хранилище (MinIO/AWS S3). Проект оснащён миграциями, тестами и удобным Makefile.

## Особенности

gRPC API (GophKeeperService)

JWT-аутентификация: access/refresh токены

Postgres как основное хранилище

Presigned PUT/GET для файлов в S3/MinIO

Миграции через pressly/goose (встроены через go:embed)

Покрытие тестами с порогом (по умолчанию 80%)

Makefile: test, vet, fmt, кроссплатформенная сборка

## Архитектура
internal/
  server/
    grpc/            # gRPC-сервер, интерсепторы, хендлеры
    services/        # бизнес-логика (UserService, EntryService)
    repositories/    # users, entries, files, refreshtokens + repomanager
    migrations/      # SQL-миграции (go:embed + goose)
    config/          # конфигурация сервера
  common/            # общие ошибки, утилиты
  auth/              # JWT-хелперы


Services работают с хранилищем через RepositoryManager — это упрощает мокирование в тестах.

EntryService генерирует presigned URL, чтобы не проксировать бинарные данные через сервер.

Интерсептор проверяет access-токен на методе Sync и прокидывает userID в context.

Быстрый старт
Требования

Go 1.21+

Postgres 14+

MinIO (локально) или AWS S3

Конфигурация

Параметры читаются из internal/server/config.Config (или переменных окружения). Пример:

## Postgres
DATABASE_DSN="postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable"

## gRPC
ENDPOINT_ADDR_GRPC="127.0.0.1:50051"

## JWT
SECRET_KEY="super-secret"
ACCESS_TOKEN_VALIDITY_DURATION="15m"
REFRESH_TOKEN_VALIDITY_DURATION="720h"   # 30 дней

## S3 / MinIO
S3_ROOT_USER="minioadmin"
S3_ROOT_PASSWORD="minioadmin"
S3_BUCKET="gophkeeper"
S3_REGION="us-east-1"
S3_BASE_ENDPOINT="http://127.0.0.1:9000"


Формат длительностей — как у time.ParseDuration (например, 15m, 24h).

## Миграции

Миграции применяются при старте через repomanager.RunMigrations и goose.UpContext c go:embed. Отдельных команд обычно не требуется.

Makefile: основные команды
vet            # go vet ./...
fmt            # go fmt ./...
test           # тесты по всем пакетам (кроме /proto), отчёт и проверка порога покрытия
build.linux    # сборка для linux/amd64 -> ./bin/cli-linux
build.darwin   # сборка для darwin/arm64 -> ./bin/cli-darwin
build.win      # сборка для windows/amd64 -> ./bin/cli.exe


Параметры по умолчанию можно переопределять:

## Порог покрытия и имя профайла
COVER_THRESHOLD=85 make test
COVER_PROFILE=coverage.out make test

# Имя приложения и директория с main
APP_NAME=server CMD_DIR=./cmd/server make build.linux


Сборка прокидывает атрибуты версии через -ldflags:

internal/buildinfo.buildVersion

internal/buildinfo.buildDate

internal/buildinfo.buildCommit

Тестирование

Репозитории тестируются через sqlmock (регулярные выражения с (?s) для многострочных SQL).

gRPC-слой покрывается юнит-тестами хендлеров и интерсепторов без запуска реального сервера.

Services тестируются с фейковым RepositoryManager; транзакции dbx.WithTx имитируются через sqlmock.ExpectBegin/Commit/Rollback.

Примеры:

## Все тесты с покрытием
go test ./internal/server/... -race -cover

## Цель Makefile с порогом покрытия
make test


Если go tool cover ругается на отсутствующие файлы (артефакт старых профилей), очистите кэш:

go clean -cache -testcache
rm -f coverage.out

## Аутентификация

UserService.generateAccessToken — JWT (HS256) с валидностью из конфигурации.

UserService.generateRefreshToken — случайный hex; хранится в таблице refresh_tokens.

Обновление токенов выполняется в транзакции: удаление старого, генерация нового, запись.

Работа с файлами (S3/MinIO)

EntryService.GetPresignedPutUrl — генерация ключа вида users/YYYY/M/D/<uuid> и presigned URL на PUT.

Клиент загружает файл напрямую в S3/MinIO.

EntryService.MarkUploaded — помечает файл как загруженный.

EntryService.GetPresignedGetURL — presigned URL на скачивание.

Отладка и советы

Если нужно подменять внешние вызовы (например, goose.UpContext или S3 Presign), удобно ввести «швы»:

var gooseUpContext = func(ctx context.Context, db *sql.DB, dir string, opts ...goose.OptionsFunc) error {
    return goose.UpContext(ctx, db, dir, opts...)
}


В тестах переменную можно подменять stub-функцией.

Структура репозитория (пример)
cmd/
  server/                 # main сервера
  cli/                    # main клиента (если есть)
internal/
  auth/
  common/
  dbx/
  server/
    config/
    grpc/
    migrations/
    models/
    repositories/
      entries/
      files/
      refreshtokens/
      repomanager/
      users/
    services/
  proto/                  # сгенерированные файлы gRPC (исключены из покрытия в Makefile)

