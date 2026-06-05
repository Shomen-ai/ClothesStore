# Будущие фиксы — ClothesStore

> Источник: код-ревью всего проекта (backend Go/Gin + frontend Vue 3), 2026-06-01.
> Это список находок для последующего исправления. Логика кода при ревью **не менялась** —
> добавлены только комментарии. Отмечай сделанное галочками.

Легенда приоритета: 🔴 высокий (безопасность/корректность) · 🟡 средний · ⚪ низкий/полировка.

---

## 🔴 Высокий приоритет — безопасность и корректность

- [ ] **Гонка на складе (oversell).** `backend/internal/repository/order_repo.go` (`Create`), `backend/internal/service/order_service.go`.
  `stock_qty` уменьшается без `WHERE stock_qty >= qty` и без блокировки строки → параллельные заказы уводят остаток в минус (нет и `CHECK (stock_qty>=0)` в `001_initial.sql`). Сток списывается при создании заказа, а не при оплате; при отмене не возвращается.
  *Фикс:* `UPDATE product_sizes SET stock_qty = stock_qty - $1 WHERE id=$2 AND stock_qty >= $1` внутри транзакции заказа, проверять `RowsAffected()==1`; рассмотреть списание при оплате и возврат при отмене; добавить `CHECK (stock_qty >= 0)`.

- [ ] **Лимит активаций промокода обходится.** `order_service.go` / `order_repo.go` (`Create`), `promo_service.go` (`ValidatePromo`).
  Чтение `activations_count` и инкремент не атомарны → `MaxActivations` можно превысить.
  *Фикс:* инкремент с условием в той же транзакции: `UPDATE ... SET activations_count = activations_count + 1 WHERE id=$1 AND (max_activations IS NULL OR activations_count < max_activations)`, откат если 0 строк.

- [ ] **JWT access и refresh взаимозаменяемы.** `backend/pkg/jwt/jwt.go`, `backend/internal/service/auth_service.go` (`RefreshToken`).
  Нет claim'а типа токена: access принимается на `/auth/refresh`, refresh проходит как access (через `AuthRequired`).
  *Фикс:* добавить claim `token_type` ("access"/"refresh"), выставлять в генераторах, требовать нужный тип в `AuthRequired` и `RefreshToken`.

- [ ] **SMTP header injection в mailer.** `backend/internal/mailer/mailer.go:109` (`buildMessage`).
  `to`/`from` вставляются в заголовки без проверки на CRLF; адрес приходит из пользовательского ввода регистрации.
  *Фикс:* валидировать через `net/mail.ParseAddress`, отклонять значения с `\r`/`\n`.

- [ ] **Публичный товар по id отдаёт неактивные.** `backend/internal/repository/product_repo.go` (`GetByID`) за `GET /api/products/:id`.
  Нет фильтра `is_active=true` (хотя `List`/`GetFeatured` его имеют) → угадавший id видит снятый товар.
  *Фикс:* добавить `AND is_active=true` (или отдельный админ-путь).

- [ ] **Токены в `localStorage`.** `frontend/src/stores/auth.js`, `frontend/src/api/axios.js`.
  Любой XSS крадёт access и долгоживущий refresh. Для ВКР — отметить как осознанный компромисс.
  *Фикс:* refresh в `HttpOnly; Secure; SameSite` cookie, access — только в памяти (Pinia).

---

## 🟡 Средний приоритет

- [ ] **Утечка `err.Error()` клиенту.** Многие хендлеры/сервисы возвращают сырой текст ошибок БД/драйвера (имена таблиц, SQL-state). *Фикс:* логировать на сервере, отдавать общий текст для 500.
- [ ] **`UpdateProduct` затирает поля.** `backend/internal/handler/admin_product_handler.go`; фронтовый `toggleActive` (`AdminProductsView.vue`) шлёт весь объект. Безусловный `UPDATE` всех колонок → частичный апдейт обнуляет `name`/`price`/`category_id`. *Фикс:* частичный апдейт (COALESCE) или load-then-merge; слать минимальный patch.
- [ ] **Загрузка файлов по расширению.** `backend/internal/service/upload_service.go`. Нет проверки MIME/размера, имя предсказуемо; `.svg`/`.html` из `/uploads` → stored XSS (смягчено: только админ). *Фикс:* `http.DetectContentType` + белый список, лимит размера, UUID-имя.
- [ ] **Гонка refresh в axios-интерсепторе.** `frontend/src/api/axios.js`. Несколько одновременных 401 дёргают `/auth/refresh`; ротация делает второй токен невалидным → разлогин. *Фикс:* один общий in-flight promise refresh + очередь запросов.
- [ ] **`ValidatePromo`: `Total` с `binding:"required"`.** `backend/internal/handler/order_handler.go`. Отклоняет легитимный `0` (Go-зеро). *Фикс:* убрать `required`, валидировать `>= 0`.
- [ ] **`JWT_SECRET` без проверки силы.** `backend/internal/config/config.go`, `backend/cmd/server/main.go` (только непустота). *Фикс:* требовать длину ≥ 32 байт, иначе fail-fast.
- [ ] **Wishlist `Add` не идемпотентен / все ошибки → 409.** `backend/internal/handler/wishlist_handler.go`, `wishlist_repo.go`. *Фикс:* `INSERT ... ON CONFLICT (user_id, product_id) DO NOTHING`; различать код `23505` от прочих ошибок.

---

## ⚪ Низкий приоритет / полировка

- [ ] **Деньги во `float64`** во всех слоях (накопление ошибок округления). Округлять на стороне SQL (`ROUND(...,2)`) или хранить копейки int.
- [ ] **Нет индексов** `orders(status)` и `orders(created_at)` — по ним фильтруют админ-список/отчёты/статистика. Добавить (можно композитный `(status, created_at)`).
- [ ] **Дублирование кода:** `formatPrice`, `statusMap/statusLabel/statusType`, `formatDate`, обработка ошибок логина/регистрации, маркизные шапки админки. Вынести в `frontend/src/utils/`.
- [ ] **Самописный base64 в mailer** (`b64UTF8`, `mailer.go`) — заменить на `encoding/base64`.
- [ ] **`isUniqueEmailErr`** (`auth_service.go`) ловит ошибку строковым матчингом с мёртвой Oracle-веткой. Проверять `pq.Error.Code == "23505"`.
- [ ] **Нет confirm-диалога** на части деструктивных админ-действий (`AdminPromoView.vue` remove/deactivate, `AddressesView.vue` remove) — в отличие от товаров/отзывов.
- [ ] **Корзина для товаров без размера** ключует по `product_size_id=0` (`frontend/src/stores/cart.js`, `ProductView.vue`) → два разных товара слипаются. Ключевать композитно.
- [ ] **`priceBoundsFromIds`** (`frontend/src/components/catalogue/priceRanges.js`) схлопывает несмежные диапазоны в один (включает «дыру» между ними). Слать дискретные диапазоны или запрещать несмежный выбор.
- [ ] **`?sale=1` при навигации внутри каталога** меняет фильтр, но не перезапрашивает (`CatalogueView.vue`: `watch(route.query, applyQuery)` без `load()`).
- [ ] **Заказ/оплата показывают сырые ID** вместо названий товаров (`OrderDetailView.vue`, `PaymentView.vue`). Денормализовать названия в ответе заказа (админка/отзывы их уже получают).
- [ ] **Нет сброса cart/wishlist при logout** (`stores/auth.js` logout не чистит другие стораджи) — данные прошлого пользователя живут до перезагрузки.
- [ ] **PaymentView — клиентская заглушка:** `confirm-payment` без серверной проверки оплаты. Осознанно для демо ВКР; проговорить на защите. Реально — подтверждать через webhook платёжного провайдера.
- [ ] **`cart.js` без guard'а:** `JSON.parse(localStorage 'cart')` без try/catch (битый JSON ломает старт); `updateQty` без нижней границы/валидации (0/отрицательные).
- [ ] **`RevenueByCategory` vs `RevenueByCity`** в `reports_repo.go` используют разные определения выручки (`price_at_order*quantity` vs `total_price`) → суммы по листам не сходятся. Унифицировать или задокументировать.
- [ ] **gofmt:** ~6 backend-файлов не gofmt-чисты по порядку импортов (предсуществующее, было в HEAD). Прогнать `gofmt -w ./...` (это переупорядочит импорты — отдельной правкой, не вместе с комментариями).

---

## ✔️ Проверено и хорошо (не требует правок)

- Серверная авторизация админки реальна: группа `/api/admin` под `AuthRequired + AdminRequired` (`main.go`), клиентский guard роутера — лишь UX.
- Владение ресурсами проверяется по user_id из JWT; возвращается 404 (не 403) — не палит существование.
- `PasswordHash` скрыт `json:"-"`; пароли — bcrypt; сравнение в постоянное время; «invalid credentials» без перечисления.
- SQL параметризован (`$N`), `ORDER BY` — по белому списку; `created_at <= NOW()` клампит сидовые «будущие» заказы.
- JWT: алгоритм подписи (HS256) форсится в keyfunc — защита от alg-confusion.

---

## Статус email-верификации (для контекста)

Отправка кода по почте **работает** (подтверждено 2026-06-01). Ранее не уходила из-за блокировки исходящего SMTP на сети Timeweb (порты 25/465/587/2525); поддержка по тикету их разблокировала. Код/конфиг были исправны. Подробности — в памяти проекта (`domain-email`).
