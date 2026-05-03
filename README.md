# ClothesStore

Интернет-магазин одежды с тёмной темой. Полный стек: Go backend, Vue 3 frontend, PostgreSQL, Docker Compose.

## Стек

| Слой | Технологии |
|------|-----------|
| Frontend | Vue 3, Pinia, Vue Router, Element Plus, Vite |
| Backend | Go, Gin, PostgreSQL (lib/pq) |
| Деплой | Docker Compose, Nginx |

## Функциональность

- Каталог товаров с фильтрацией по категории, размеру, сортировке и поиском
- Страница товара с галереей изображений и выбором размера
- Корзина с промокодами и оформлением заказа
- Аккаунт: профиль, история заказов, адреса доставки, избранное
- Админ-панель: управление товарами, заказами, промокодами, статистика продаж
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
