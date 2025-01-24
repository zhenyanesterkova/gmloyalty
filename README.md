# gmloyalty

Сервис предназначен для управления учётными записями пользователей и их накопительными счетами в рамках программы лояльности интернет-магазина Гофермарт.

## Установка

1. Склонировать репозиторий:

```Git
   git clone https://github.com/zhenyanesterkova/gmloyalty.git
```

2. Установить переменные окружения

```bash
   export RUN_ADDRESS="localhost:8081"
   export DATABASE_URI="postgres://gopher:gopher@localhost:5432/gopherloyalty?sslmode=disable"
   export ACCRUAL_SYSTEM_ADDRESS="http://localhost:8083"
```

3. Запустить сервис

```Makefile
make run
```

или

```Dockerfile
docker-compose up
```

или

```Go
go mod tidy
```

```Go
go build -o gophermart cmd/gophermart/main.go
./gophermart
```

## API

```Go
POST /api/user/register - регистрация пользователя;
POST /api/user/login - аутентификация пользователя;
POST /api/user/orders - загрузка пользователем номера заказа для расчёта;
GET /api/user/orders - получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
GET /api/user/balance - получение текущего баланса счёта баллов лояльности пользователя;
POST /api/user/balance/withdraw - запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
GET /api/user/withdrawals - получение информации о выводе средств с накопительного счёта пользователем.
```

## Конфигурация

Сервис может быть сконфигурирован с использованием переменных окружения или флагов:

```
RUN_ADDRESS or -a - адрес и порт запуска сервиса
DATABASE_URI or -d - адрес подключения к базе данных
ACCRUAL_SYSTEM_ADDRESS or -r - адрес системы расчёта начислений
```
