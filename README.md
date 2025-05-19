# 🌦️ Weather Subscriber API

Цей проєкт — API-сервіс для підписки на прогноз погоди в різних містах. Побудовано на Golang з використанням MySQL та Docker.

## 📦 Функціональність

- Створення підписки (email + місто + частота)
- Перегляд усіх підписок
- Перегляд списку міст з пагінацією
- Перевірка стану сервера (`/health`)

## 🛠️ Стек технологій

- Go (Golang)
- MySQL
- Docker + Docker Compose
- Swagger (OpenAPI)

---

##  Інсталяція

### 1. Клонування репозиторію

```bash
git clone https://github.com/WOLFnik5/weather_subscription.git
cd weather_subscription
```

### 2. Налаштування `.env`

```bash
cp .env.dist .env
```
Заповни необхідні credentials

### 3. Запуск міграцій

> Увага: переконайся, що каталог `./db/migrations` містить файли міграцій.

```bash
docker compose up -d db
docker compose run --rm migrate up
```

### 4. Збірка та запуск API додатку

```bash
docker compose up --build -d app
```

Додаток буде доступний на: [http://localhost:8080](http://localhost:8080)

### 5. Запуск тестів

```bash
docker compose up -d db_test
docker compose --env-file .env_test run --rm migrate_test up
docker compose run tester
```

---

##  API Документація

Swagger UI доступний за адресою: [swagger.yaml](https://editor.swagger.io/?url=https://raw.githubusercontent.com/WOLFnik5/weather_subscription/refs/heads/main/swagger.yaml)

Якщо посилання не доступне відкрий `swagger.yaml` на https://editor.swagger.io


---

## 📁 Структура проєкту

```
.
├── db/              # Підключення до бази даних
├── handler/         # HTTP-обробники
├── model/           # Моделі та SQL-логіка
├── router/          # Налаштування маршрутів
├── main.go          # Точка входу
├── Dockerfile
├── docker-compose.yaml
├── .env
└── README.md
```
---

## 📄 Ліцензія

MIT