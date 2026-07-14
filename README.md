
## 🛠 Technology Stack

### Backend
- **Development Language:** Go
- **Framework:** [Echo (v4)](https://echo.labstack.com/) for handling HTTP requests.
- **ORM:** [GORM](https://gorm.io/) for working with relational databases.
- **Security:**
  - JWT (JSON Web Tokens) for authentication.
  - Password hashing via SHA512-CRYPT standard (backward compatible with MD5-CRYPT for legacy databases).
  - Strict IDOR protection: rights to access and manipulate resources are isolated (Multi-tenancy).

### Frontend
- **Language/Framework:** [Vue 3](https://vuejs.org/) (Composition API) + [Vite](https://vitejs.dev/).
- **Styling:** Tailwind CSS. The interface is built with a unified, strict design (Premium Enterprise).
- **Localization:** Vue I18n (built-in support for Russian and English languages).

---

## 🔑 Authorization & Roles

The system implements strict access control using role segregation. There are two access levels:

1. **SuperAdmin:**
   - Full access to all system resources.
   - Rights to create, edit, and delete other administrators.
   - Global management of any domain and system-level settings (e.g., Domain Aliases).

2. **Domain Admin:**
   - Access only to domains explicitly assigned to the specific administrator (linked via the `domain_admins` table).
   - Data isolation: attempting to request a list of mailboxes, aliases, or make modifications to someone else's domain will be aborted with a `403 Forbidden` error.

---

## 🚀 Key Features

### 1. Domain Management
- View current summary statistics (number of mailboxes, aliases, quotas) using optimized SQL queries (JOINs).
- Manage limits (mailbox count and total quota volume).
- Automatic validation and real-time checking of domain MX record statuses.
- Soft management (Active/Inactive) and cascading deletion (deleting a domain carefully deletes all its aliases, mailboxes, and Sieve filters).
- **DNS Diagnostics & Generator:** Comprehensive domain audits (SPF, DKIM, DMARC, MX, SSL, RBL) with one-click DNS record generation and recommendations.

### 2. Mailboxes & Aliases
- Full CRUD for mailboxes with individual quota limit management.
- **Alphabetical Pagination:** A-Z/А-Я index filter for lightning-fast navigation in large mailbox lists.
- **Password Strength Meter:** Real-time visual feedback on password complexity during creation/modification.
- **Sieve Filter Builder:** A powerful visual constructor for message processing rules (Move to folder, Forward, Discard, Reject, Mark as Read).
- **Advanced Vacation:** Set up modern auto-replies via Sieve with custom subjects, multi-line messages, and interval control.
- Generation of cryptographically secure passwords, support for mandatory password expiration (`password_expiry`).
- Email notifications sent to administrative recipients when a new mailbox is created or when a password is changed.
- Alias management (forwarding emails from one address to a group of addresses) and **Catch-All Domain Aliases** support. Automatic creation of Postmaster aliases when new domains are added.

### 3. Admin & Monitoring Tools
- **System Health Dashboard:** Real-time monitoring of CPU, RAM, Disk, SSL certificate validity, and system service statuses.
- **Postfix Log Analyzer:** Interactive log parser that groups events into complete mail transactions by Queue ID (supports ISO 8601 & traditional dates, up to 10k lines). Displays delivery status distribution, top senders/recipients/clients, incoming rejects (NOQUEUE), and expandible delivery logs.
- **Fail2Ban Management:** View active bans across all jails (SSH, Postfix, Dovecot) and unban malicious IPs with a single click.
- **Mail Queue Management:** View, search, and delete messages in the Postfix mail queue directly from the web interface.
- Tool for sending verification (test) emails on behalf of an administrator.
- **Server-wide Broadcasts:** Send secure global notifications to all active mailboxes across all domains with a built-in Postfix integration guide.

### 4. Audit Logs
- Logging of every critical action to the backend database (localized action types).
- Recording *who*, *where* (in which domain), *when*, and *what* action was performed (e.g., "create mailbox", "update domain").
- Regular administrators' access to the log is restricted only to events within their permitted domains.

---

## 🛡 Security

The project was designed with a strong focus on preventing popular web vulnerabilities. Key protection mechanisms:
- **API Authentication:** Almost all endpoints (except login initialization) are protected by `auth.JWTMiddleware`, preventing unauthorized access (protection against unauthorized use or application crashes due to Nil Pointers).
- **IDOR Protection:** Every `POST`, `PUT`, or `DELETE` request for a mailbox, alias, Sieve filter, or mailing tool strictly verifies ownership of the target domain via the `domain_admins` database.
- **Privilege Separation (Agent Architecture):** To minimize the attack surface, the system is split into two processes:
  - **Agent Node (`mailadmin --agent`):** Runs as `root`, performs privileged tasks (Postfix, Fail2Ban), and listens only on a local Unix socket. *Note: In Docker, ensure the image contains the necessary tools (postqueue, fail2ban-client) or mount them from the host.*
  - **Web Node (`mailadmin --web`):** Runs as an unprivileged user, handles the API/UI, and communicates with the Agent via IPC.

---

## 🎨 Design & UI/UX (Design Guidelines)

The frontend interface follows a unified set of rules:
- **Alignment:** All elements (headers, form fields) are left-aligned to improve information scanning structure.
- **Border Radii:** `rounded-xl` (12px) for control buttons, `rounded-[32px]` for panel cards.
- **Inputs:** Aggressive bright colors for value text in fields are avoided. Only semantic tags (green/red) are used for statuses.
- **Action Placement:** Table Action buttons are strictly centered in their respective columns.

---

## 📦 Deployment

### Requirements
- Installed Go version `1.20` or newer.
- Installed Node.js (for frontend building).
- Database (MySQL / MariaDB or compatible) with `utf8mb4` encoding.

### Configuration (.env)
Before running the application, create a `.env` file in the `backend` directory. You can use `.env.example` as a template:

```bash
cp .env.example .env
```

Key variables:
- `DB_DSN`: Database connection string (DSN). Ensure `parseTime=True` is included.
- `JWT_SECRET`: A secure key for signing tokens (minimum 32 characters recommended).
- `LISTEN_ADDR`: The address and port for the Go server (e.g., `:8080`).
- `CORS_ORIGIN`: The URL of your frontend (required for cross-origin requests).
- `MAIL_ROOT`: Base directory for mailbox data (e.g., `/data/mail`).
- `SIEVE_ROOT`: Directory for storing Sieve scripts.
- `LOG_PATH`: Path to the main mail system log for real-time monitoring (e.g., `/var/log/mail.log`).

### Running the Backend
The backend uses configuration, usually from environment files (ENV) or the `internal/config` setup.

```bash
cd backend
go mod tidy
go build -o mailadmin-bin ./main.go

# Start the Agent (as root)
sudo ./mailadmin-bin --agent &

# Start the Web Node (as unprivileged user)
./mailadmin-bin --web
```

### Running the Frontend
```bash
cd frontend
npm install
npm run dev     # For development
npm run build   # Build for production (outputs to the 'dist' folder)
```

For production use, it is recommended to compile the frontend (`npm run build`) and configure a reverse proxy server (Nginx or Caddy) where `/api` redirects to the Echo backend port, and the root route (`/`) serves static files from `dist`.

---


# Документация проекта MailAdmin

**MailAdmin** — это современная, производительная и безопасная панель управления почтовыми серверами. Она разработана как альтернатива классическому PostfixAdmin, предлагая современный пользовательский интерфейс, строгую систему разграничения прав и высокую скорость работы благодаря бэкенду на Go.

[**Документация по API**](docs/API.md)

---

## 🛠 Технологический стек

### Бэкенд
- **Язык разработки:** Go
- **Фреймворк:** [Echo (v4)](https://echo.labstack.com/) для обработки HTTP-запросов.
- **ORM:** [GORM](https://gorm.io/) для работы с реляционными базами данных.
- **Безопасность:**
  - JWT (JSON Web Tokens) для аутентификации.
  - Хеширование паролей по стандарту SHA512-CRYPT (обратная совместимость с MD5-CRYPT для легаси баз данных).
  - Строгая защита от IDOR: права на доступ и манипуляцию ресурсами изолированы (Multi-tenancy).

### Фронтенд
- **Язык/Фреймворк:** [Vue 3](https://vuejs.org/) (Composition API) + [Vite](https://vitejs.dev/).
- **Стилизация:** Tailwind CSS. Интерфейс выполнен в едином строгом дизайне (Premium Enterprise).
- **Локализация:** Vue I18n (встроенная поддержка Русского и Английского языков).

---

## 🔑 Авторизация и роли

В системе реализовано жесткое разграничение прав доступа с помощью системы ролей. Существует два уровня доступа:

1. **Суперадмин (SuperAdmin):**
   - Полный доступ ко всем ресурсам системы.
   - Права на создание, редактирование и удаление других администраторов.
   - Глобальное управление любым доменом и настройками системного уровня (например, Alias Domain).

2. **Администратор домена (Domain Admin):**
   - Доступ только к доменам, назначенным конкретному администратору (связки через таблицу `domain_admins`).
   - Изоляция данных: попытка запросить список ящиков, алиасов или внести изменения в чужой домен будет прервана с ошибкой `403 Forbidden`.

---

## 🚀 Основной функционал

### 1. Управление доменами
- Просмотр актуальной сводной статистики (количество почтовых ящиков, алиасов, квоты) с помощью оптимизированных SQL-запросов (JOIN).
- Управление лимитами (на количество ящиков и объем квоты).
- **Диагностика и генератор DNS:** Полный аудит домена (SPF, DKIM, DMARC, MX, SSL, RBL) и автоматическая генерация правильных DNS-записей.
- Автоматическая валидация и проверка статуса MX-записей домена в реальном времени.
- Мягкое управление (Active/Inactive) и каскадное удаление (удаление домена влечет бережное удаление всех его алиасов, ящиков и фильтров Sieve).

### 2. Почтовые ящики и алиасы
- Полноценный CRUD для почтовых ящиков с возможностью управления лимитами индивидуальной квоты.
- **Алфавитная пагинация**: Фильтр по алфавиту (A-Z, А-Я) для мгновенного поиска в больших списках ящиков.
- **Индикатор сложности пароля**: Визуальный контроль надежности пароля при создании и изменении аккаунтов.
- **Конструктор фильтров Sieve:** Мощный визуальный конструктор правил обработки почты (Перемещение в папку, Пересылка, Удаление, Отклонение, Пометка прочитанным).
- **Продвинутый Автоответчик:** Настройка современных автоответов через Sieve с поддержкой многострочного текста, интервалов и условий.
- Генерация криптографически безопасных паролей, поддержка принудительного устаревания пароля (`password_expiry`).
- Отправка email-уведомлений администраторам при создании нового ящика или смене пароля.
- Управление пересылками и поддержка **Catch-All алиасов** для домена. Автосоздание Postmaster-алиасов при добавлении новых доменов.

### 3. Инструменты и Мониторинг
- **Дашборд здоровья системы:** Мониторинг RAM, Диска, статуса SSL-сертификатов и сессий IMAP в реальном времени.
- **Анализатор логов Postfix**: Интерактивный разбор до 10000 строк логов. Собирает события в транзакции по Queue ID (поддерживает ISO 8601 и традиционные даты), отображает распределение статусов, топ отправителей/получателей/клиентов, отклоненные письма (NOQUEUE) и детальный лог доставки.
- **Управление Fail2Ban:** Мониторинг активных блокировок по всем тюрьмам (SSH, Postfix, Dovecot) и мгновенный разбан (Unban) вредоносных IP в один клик.
- **Управление почтовой очередью:** Просмотр, поиск и удаление сообщений в очереди Postfix прямо из панели управления.
- Инструмент отправки поверочных (тестовых) писем от имени администратора.
- **Глобальные рассылки (Broadcast)**: Безопасная отправка рассылок на весь сервер (все домены) с интерактивной инструкцией по настройке ограничений в Postfix.

### 4. Журнал аудита (Audit Logs)
- Запись каждого критического действия в бэкенд-базу (русифицированные типы событий).
- Фиксация того, *кто*, *где* (в каком домене), *когда* и *какое* действие совершил (например: «create mailbox», «update domain»).
- Доступ обычным администраторам к журналу ограничен только событиями в разрешенных им доменах.

---

## 🛡 Безопасность

Проект спроектирован с упором на предотвращение популярных веб-уязвимостей. Основные механизмы защиты:
- **Аутентификация API:** Почти все конечные точки (кроме инициализации логина) закрыты `auth.JWTMiddleware`, предотвращая несанкционированный доступ (защита от неавторизованного использования или падения приложения по причине Nil Pointer-ов).
- **IDOR Protection:** Каждый запрос `POST`, `PUT` или `DELETE` на запись почтового ящика, алиаса, фильтра Sieve или инструмента рассылки строго проверяет владение целевым доменом через базу `domain_admins`.
- **Разделение привилегий (Agent Architecture):** Для минимизации рисков система разделена на два процесса:
    - **Agent Node (`mailadmin --agent`):** Работает от `root`, выполняет системные команды (Postfix, Fail2Ban) и слушает только локальный Unix-сокет. *Примечание: При использовании Docker убедитесь, что в образе установлены необходимые утилиты (postqueue, fail2ban-client) или они проброшены с хоста.*
    - **Web Node (`mailadmin --web`):** Работает от бесправного пользователя, обрабатывает API и UI, взаимодействуя с Агентом через IPC.

---

## 🎨 Дизайн и UI/UX (Design Guidelines)

 Интерфейс фронтенда следует единому своду правил:
- **Выравнивание:** Все элементы (заголовки, поля форм) выровнены по левому краю для улучшения структуры сканирования информации.
- **Округления:** Радиус `rounded-xl` (12px) для кнопок контроля, `rounded-[32px]` для карточек панелей.
- **Инпуты:** Избегаются агрессивные яркие цвета у шрифтов значений в полях. Для статусов присутствуют только семантические тэги (зеленые/красные).
- **Размещение действий:** Табличные Action-кнопки строго центрированы в своей колонке. 

---

## 📦 Развертывание

### Требования
- Установленный Go версии `1.20` или новее.
- Установленный Node.js (для сборки фронтенда).
- База данных (MySQL / MariaDB или совместимая) с кодировкой `utf8mb4`.

### Конфигурация (.env)
Перед запуском приложения создайте файл `.env` в директории `backend`. Вы можете использовать `.env.example` как шаблон:

```bash
cp .env.example .env
```

Основные переменные:
- `DB_DSN`: Строка подключения к БД (DSN). Убедитесь, что параметр `parseTime=True` включен.
- `JWT_SECRET`: Секретный ключ для подписи токенов (рекомендуется минимум 32 символа).
- `LISTEN_ADDR`: Адрес и порт для запуска Go-сервера (например, `:8080`).
- `CORS_ORIGIN`: URL вашего фронтенда (необходим для разрешения кросс-доменных запросов).
- `MAIL_ROOT`: Корневая директория для данных почтовых ящиков (например, `/data/mail`).
- `SIEVE_ROOT`: Путь к директории со скриптами Sieve.
- `LOG_PATH`: Путь к основному логу почтовой системы для мониторинга (например, `/var/log/mail.log`).

### Запуск Backend-части
Бэкенд использует конфигурацию, как правило, из файлов среды (ENV) или конфига конфигуратора `internal/config`.

```bash
cd backend
go mod tidy
go build -o mailadmin-bin ./main.go

# Запустите Агента (от root)
sudo ./mailadmin-bin --agent &

# Запустите Веб-узел (от обычного пользователя)
./mailadmin-bin --web
```

### Запуск Frontend-части
```bash
cd frontend
npm install
npm run dev     # Для разработки
npm run build   # Сборка для продакшена (в папку dist)
```

Для использования в продакшене рекомендуется скомпилировать фронтенд (`npm run build`) и настроить обратный прокси-сервер (Nginx или Caddy), где `/api` перенаправляется на порт Echo бэкенда, а корневой маршрут (`/`) раздает статические файлы из `dist`.

---
<br/>

# MailAdmin Project Documentation

**MailAdmin** is a modern, high-performance, and secure control panel for email servers. It is designed as an Enterprise alternative to the classic PostfixAdmin, offering a premium user interface, a strict role-based access control system, and high speed thanks to its Go-based backend.

[**API Documentation**](docs/API_EN.md)

---
