# Документация по API MailAdmin

В этом документе приведено описание всех доступных эндпоинтов API панели управления MailAdmin.

## Общая информация

- **Базовый URL**: `/api`
- **Формат данных**: JSON
- **Аутентификация**: Большинство эндпоинтов требуют передачи JWT токена в заголовке `Authorization`.
  Формат: `Authorization: Bearer <token>`

---

## 1. Аутентификация (`/api/auth`)

### POST `/api/auth/login`
Аутентификация пользователя и получение токена.

- **Аутентификация**: Не требуется.
- **Тело запроса**:
  ```json
  {
    "username": "admin@example.com",
    "password": "your_password"
  }
  ```
- **Успешный ответ (200 OK)**:
  ```json
  {
    "access_token": "eyJhbGciOi...",
    "user": {
      "username": "admin@example.com",
      "superadmin": true
    }
  }
  ```
  *Примечание*: Также устанавливается HttpOnly Cookie `refreshToken`.

### POST `/api/auth/refresh`
Обновление Access Token с использованием Refresh Token.

- **Аутентификация**: Использует Cookie `refreshToken`.
- **Успешный ответ (200 OK)**:
  ```json
  {
    "access_token": "eyJhbGciOi..."
  }
  ```

### POST `/api/auth/logout`
Выход из системы.

- **Аутентификация**: Не требуется. Очищает куку `refreshToken`.
- **Успешный ответ**: `200 OK` (без тела).

### GET `/api/auth/me`
Получение информации о текущем пользователе.

- **Аутентификация**: Требуется JWT.
- **Успешный ответ (200 OK)**: Данные пользователя (Claims).

### POST `/api/auth/change-password`
Смена пароля текущего пользователя.

- **Аутентификация**: Требуется JWT.
- **Тело запроса**:
  ```json
  {
    "new_password": "new_secure_password"
  }
  ```
- **Успешный ответ (200 OK)**:
  ```json
  {
    "message": "password changed successfully"
  }
  ```

---

## 2. Управление администраторами (`/api/admins`)
*Доступно только для SuperAdmin.*

### GET `/api/admins`
Получение списка всех администраторов.

- **Аутентификация**: Требуется JWT (SuperAdmin).
- **Успешный ответ (200 OK)**: Массив объектов администраторов.

### GET `/api/admins/:username`
Получение деталей конкретного администратора и управляемых им доменов.

- **Аутентификация**: Требуется JWT (SuperAdmin).
- **Успешный ответ (200 OK)**:
  ```json
  {
    "admin": { ... },
    "domains": ["example.com", "test.com"]
  }
  ```

### POST `/api/admins`
Создание нового администратора.

- **Аутентификация**: Требуется JWT (SuperAdmin).
- **Тело запроса**:
  ```json
  {
    "username": "newadmin@example.com",
    "password": "secure_password",
    "superadmin": false,
    "active": true,
    "domains": ["example.com"]
  }
  ```
- **Успешный ответ (201 Created)**: Созданный объект администратора.

### PUT `/api/admins/:username`
Редактирование администратора.

- **Аутентификация**: Требуется JWT (SuperAdmin).
- **Тело запроса**:
  ```json
  {
    "password": "optional_new_password",
    "active": true,
    "superadmin": false,
    "phone": "+123456789",
    "email_other": "backup@gmail.com",
    "domains": ["example.com", "newdomain.com"]
  }
  ```
- **Успешный ответ**: `200 OK` (без тела).

### DELETE `/api/admins/:username`
Удаление администратора.

- **Аутентификация**: Требуется JWT (SuperAdmin).
- **Успешный ответ**: `204 No Content`.

---

## 3. Управление доменами (`/api/domains`)

### GET `/api/domains`
Получение списка доменов с их статистикой.

- **Аутентификация**: Требуется JWT.
- **Параметры строки запроса (Query)**:
  - `search` (string, optional) — поиск по имени или описанию.
  - `active` (bool, optional) — фильтр по статусу (`true`/`false`).
- **Успешный ответ (200 OK)**:
  ```json
  [
    {
      "domain": "example.com",
      "description": "Основной домен",
      "aliases": 10,
      "mailboxes": 5,
      "maxquota": 10240,
      "quota": 2048,
      "transport": "virtual",
      "backupmx": false,
      "active": true,
      "mailboxes_count": 3,
      "aliases_count": 2,
      "quota_used": 1500
    }
  ]
  ```

### POST `/api/domains`
Создание нового домена.
*Примечание: Автоматически создаются RFC-алиасы (postmaster, abuse, hostmaster).*

- **Аутентификация**: Требуется JWT (SuperAdmin).
- **Тело запроса**:
  ```json
  {
    "domain": "newdomain.com",
    "description": "Описание",
    "aliases": 0,
    "mailboxes": 0,
    "maxquota": 0,
    "quota": 0,
    "transport": "virtual",
    "backupmx": false,
    "active": true
  }
  ```
- **Успешный ответ (201 Created)**: Созданный объект домена.

### PUT `/api/domains/:id`
Редактирование домена.

- **Аутентификация**: Требуется JWT (SuperAdmin).
- **Тело запроса**: Те же поля, что и при создании (кроме `domain`).
- **Успешный ответ (200 OK)**: Обновленный объект домена.

### DELETE `/api/domains/:id`
Каскадное удаление домена и всех связанных с ним ящиков, алиасов и настроек.

- **Аутентификация**: Требуется JWT (SuperAdmin).
- **Успешный ответ**: `204 No Content`.

### GET `/api/domains/:id/dns`
Генерация рекомендуемых DNS-записей (MX, SPF, DKIM, DMARC) для домена.

- **Аутентификация**: Требуется JWT (Доступ к домену).
- **Успешный ответ (200 OK)**:
  ```json
  {
    "domain": "example.com",
    "spf": "v=spf1 ip4:1.2.3.4 mx a -all",
    "dkim": "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
    "dmarc": "v=DMARC1; p=quarantine; rua=mailto:postmaster@example.com"
  }
  ```

---

## 4. Управление почтовыми ящиками (`/api/mailboxes`)

### GET `/api/mailboxes`
Получение списка ящиков.

- **Аутентификация**: Требуется JWT.
- **Параметры строки запроса (Query)**:
  - `domain` (string) — обязателен, если не указан `search`.
  - `search` (string, optional) — поиск по username или имени.
  - `page` (int, default 1).
  - `limit` (int, default 50).
  - `active` (bool, optional).
- **Заголовки ответа**: `X-Total-Count` — общее количество.
- **Успешный ответ (200 OK)**: Массив объектов `Mailbox` с полями `quota_used` и `messages`.

### POST `/api/mailboxes`
Создание почтового ящика.
*Примечание: Автоматически создается зеркальный алиас.*

- **Аутентификация**: Требуется JWT (Доступ к домену).
- **Тело запроса**:
  ```json
  {
    "username": "user@example.com",
    "password": "secure_password",
    "name": "Иван Иванов",
    "quota": 1024,
    "active": true,
    "phone": "+123456789",
    "email_other": "backup@gmail.com"
  }
  ```
- **Успешный ответ (201 Created)**: Созданный объект.

### PUT `/api/mailboxes/:username`
Редактирование ящика.

- **Аутентификация**: Требуется JWT (Доступ к домену).
- **Тело запроса**:
  ```json
  {
    "password": "optional_new_password",
    "name": "Новое Имя",
    "quota": 2048,
    "active": true,
    "phone": "+123456789",
    "email_other": "backup@gmail.com"
  }
  ```
- **Успешный ответ (200 OK)**: Обновленный объект.

### DELETE `/api/mailboxes/:username`
Удаление ящика и связанных алиасов/автоответчиков.

- **Аутентификация**: Требуется JWT (Доступ к домену).
- **Успешный ответ**: `204 No Content`.

### Массовые операции (Batch)

#### POST `/api/mailboxes/batch/create`
Массовое создание ящиков со стандартным паролем.
- **Тело запроса**:
  ```json
  {
    "domain": "example.com",
    "prefixes": ["user1", "user2"],
    "password": "common_password",
    "quota": 1024,
    "active": true
  }
  ```
- **Ответ (201 Created)**: `{"created": 2}`

#### POST `/api/mailboxes/batch/delete`
Массовое удаление.
- **Тело запроса**: `{"usernames": ["user1@example.com", "user2@example.com"]}`
- **Ответ**: `204 No Content`

#### POST `/api/mailboxes/batch/status`
Массовое изменение статуса (активен/заблокирован).
- **Тело запроса**: `{"usernames": ["user1@example.com"], "active": false}`
- **Ответ**: `204 No Content`

---

## 5. Управление алиасами (`/api/aliases`)

### GET `/api/aliases`
Список алиасов пересылки.

- **Аутентификация**: Требуется JWT.
- **Параметры Query**: `domain`, `search`, `page`, `limit`.

### POST `/api/aliases`
Создание алиаса.
- **Тело**: `{"address": "info@example.com", "goto": "user@example.com,boss@example.com", "domain": "example.com", "active": true}`
- **Ответ**: `201 Created`.

### PUT `/api/aliases/:address`
Редактирование.
- **Тело**: `{"goto": "new@example.com", "active": true}`

### DELETE `/api/aliases/:address`
Удаление.

### Алиасы Доменов (`/api/aliases/domain-aliases`)
*Только для SuperAdmin.*
- `GET /api/aliases/domain-aliases?domain=...`
- `POST /api/aliases/domain-aliases` (Тело: `AliasDomain` объект)
- `DELETE /api/aliases/domain-aliases/:alias_domain`

---

## 6. Логи и статистика

### GET `/api/logs`
Просмотр логов аудита действий администраторов.
- **Параметры**: `page`, `limit` (max 500), `domain`.

### GET `/api/stats/dashboard`
Сводная статистика для главной страницы (кол-во доменов, ящиков, занятая квота, последние логи).

---

## 7. Системный мониторинг (`/api/system`)
*Доступно только для SuperAdmin.*

- `GET /api/system/health` — Статус RAM, CPU, Диска, сервисов (Postfix, Dovecot, MySQL и др.).
- `GET /api/system/queue` — Просмотр почтовой очереди Postfix.
- `POST /api/system/queue/flush` — Принудительная отправка очереди.
- `DELETE /api/system/queue/:id` — Удаление письма из очереди (`id` или `all`).
- `GET /api/system/logs?lines=200&search=...` — Просмотр системных логов `/var/log/mail.log`.
- `GET /api/system/fail2ban` — Список забаненных IP.
- `DELETE /api/system/fail2ban/unban?ip=...&jail=...` — Разбан IP.

---

## 8. Инструменты и Автоответчик

- `GET /api/tools/check-mx/:domain` — Быстрая проверка MX-записей.
- `POST /api/tools/send-email` — Отправка тестового письма.
- `POST /api/tools/broadcast` — Рассылка объявлений по всем ящикам домена.

### Автоответчик (Vacation)
- `GET /api/mailboxes/:username/vacation`
- `PUT /api/mailboxes/:username/vacation`
  - Тело: `{"subject": "...", "body": "...", "active": true, "activefrom": "...", "activeuntil": "...", "interval_time": 1}`

---

## 9. Sieve-фильтры (`/api/sieve`)

- `GET /api/sieve/:username` — Получение правил (в т.ч. `GLOBAL` для суперадмина).
- `POST /api/sieve/:username` — Сохранение правил (JSON).
- `POST /api/sieve/:username/import` — Импорт сырого `.sieve` файла с сервера в веб-интерфейс.
