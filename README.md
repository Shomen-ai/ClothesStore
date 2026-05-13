# ClothesStore

Интернет-магазин одежды с тёмной темой. Полный стек: Go backend, Vue 3 frontend, PostgreSQL, Docker Compose.

## Стек

| Слой | Технологии |
|------|-----------|
| Frontend | Vue 3, Pinia, Vue Router, Element Plus, Vite |
| Backend | Go, Gin, PostgreSQL (lib/pq) |
| Деплой | Docker Compose, Nginx |

## Функциональность

- Каталог товаров с фильтрацией по категории, размеру, диапазону цены, флагу распродажи и поиском
- Страница товара с галереей изображений и выбором размера
- Корзина с промокодами и оформлением заказа
- Аккаунт: профиль, история заказов, адреса доставки, избранное
- Админ-панель: управление товарами (в т.ч. флаг SALE), заказами, промокодами, статистика продаж
- JWT-аутентификация (access + refresh токены)
- Адаптивный дизайн (mobile-first)

## Запуск

```bash
cp .env.example .env
# Отредактируй JWT_SECRET в .env

docker compose up --build
```

Сайт откроется на `http://localhost`.

## Структура

```
├── backend/
│   ├── cmd/server/          # Точка входа
│   ├── internal/
│   │   ├── config/          # Конфигурация
│   │   ├── db/              # Подключение к БД и миграции
│   │   ├── handler/         # HTTP-обработчики
│   │   ├── middleware/       # Auth, RBAC
│   │   ├── model/           # Структуры данных
│   │   ├── repository/      # Слой работы с БД
│   │   └── service/         # Бизнес-логика
│   └── pkg/jwt/             # JWT-утилиты
├── frontend/
│   └── src/
│       ├── api/             # Axios-клиенты
│       ├── components/      # Переиспользуемые компоненты
│       ├── stores/          # Pinia-сторы
│       └── views/           # Страницы
├── docker-compose.yml
└── .env.example
```

## Переменные окружения

| Переменная | Описание | По умолчанию |
|-----------|----------|-------------|
| `JWT_SECRET` | Секрет для подписи JWT | — |
| `DB_CONN_STR` | Строка подключения к PostgreSQL | задаётся в compose |
| `PORT` | Порт backend | `8080` |
| `UPLOADS_DIR` | Папка для загружаемых файлов | `./uploads` |

## Дефолтный admin

После первого запуска доступен аккаунт администратора:

- Email: `admin@clothesstore.ru`
- Пароль: `admin123`

Смени пароль после первого входа.

## Миграции

SQL-файлы лежат в `backend/internal/db/migrations/`. PostgreSQL применяет их при первом запуске volume `pgdata` (через `/docker-entrypoint-initdb.d/`).

Если БД уже инициализирована, применить новые миграции вручную:

```bash
docker compose exec db psql -U store -d clothesstore -f /docker-entrypoint-initdb.d/002_sale.sql
```

## Импорт каталога с внешнего сайта

В `scripts/import_torch.py` лежит парсер каталога [torch-fff.com](https://torch-fff.com) — собирает товары и заливает их через админ-API. Используется для наполнения демо-БД.

```bash
cd scripts
pip install -r requirements.txt

# 1) Парсит сайт → products.json + папка _images/
python import_torch.py scrape

# 2) Заливает в локальный backend (backend должен быть запущен на :8080)
python import_torch.py upload

# Или одним шагом
python import_torch.py all
```

Параметры (через env):
- `TORCH_MAX_IMAGES` — лимит картинок на товар (по умолчанию `8`)
- `API_URL` — адрес backend (`http://localhost:8080/api`)
- `ADMIN_EMAIL`, `ADMIN_PASSWORD` — учётка админа

Скрипт идемпотентен: повторный `upload` пропустит товары, чьё имя уже есть в БД.
