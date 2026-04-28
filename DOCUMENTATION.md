# Документация проекта MailAdmin

**MailAdmin** — это современная, производительная и безопасная панель управления почтовыми серверами. Она разработана как Enterprise-альтернатива классическому PostfixAdmin, предлагая премиальный пользовательский интерфейс, строгую систему разграничения прав и высокую скорость работы благодаря бэкенду на Go.

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
- Автоматическая валидация и проверка статуса MX-записей домена в реальном времени.
- Мягкое управление (Active/Inactive) и каскадное удаление (удаление домена влечет бережное удаление всех его алиасов, ящиков и настроек автоответчиков).

### 2. Почтовые ящики и алиасы
- Полноценный CRUD для почтовых ящиков с возможностью управления лимитами индивидуальной квоты.
- Настройки автоответчика (Vacation): установка темы, текста и расписания активации.
- Генерация криптографически безопасных паролей, поддержка принудительного устаревания пароля (`password_expiry`).
- Управление алиасами (направление писем с одного адреса на группу адресов). Автосоздание Postmaster-алиасов при добавлении новых доменов.

### 3. Инструменты (Admin Tools)
- Инструмент отправки поверочных (тестовых) писем от имени администратора.
- Инструмент широковещательной рассылки (Broadcast) оповещений всем пользователям в домене или группе доменов (обрабатывается асинхронно для защиты от таймаута).

### 4. Журнал аудита (Audit Logs)
- Запись каждого критического действия в бэкенд-базу.
- Фиксация того, *кто*, *где* (в каком домене), *когда* и *какое* действие совершил (например: «create mailbox», «update domain»).
- Доступ обычным администраторам к журналу ограничен только событиями в разрешенных им доменах.

---

## 🛡 Безопасность

Проект спроектирован с упором на предотвращение популярных веб-уязвимостей. Основные механизмы защиты:
- **Аутентификация API:** Почти все конечные точки (кроме инициализации логина) закрыты `auth.JWTMiddleware`, предотвращая несанкционированный доступ (защита от неавторизованного использования или падения приложения по причине Nil Pointer-ов).
- **IDOR Protection:** Каждый запрос `POST`, `PUT` или `DELETE` на запись почтового ящика, автоответчика, алиаса или инструмента рассылки строго проверяет владение целевым доменом через базу `domain_admins`.

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

### Запуск Backend-части
Бэкенд использует конфигурацию, как правило, из файлов среды (ENV) или конфига конфигуратора `internal/config`.

```bash
cd backend
go mod tidy
go build -o mailadmin.exe ./main.go
# Запустите процесс
./mailadmin.exe
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

---

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
- Soft management (Active/Inactive) and cascading deletion (deleting a domain carefully deletes all its aliases, mailboxes, and auto-reply settings).

### 2. Mailboxes & Aliases
- Full CRUD for mailboxes with individual quota limit management.
- Vacation (Auto-reply) settings: setting subject, body, and activation schedule.
- Generation of cryptographically secure passwords, support for mandatory password expiration (`password_expiry`).
- Alias management (forwarding emails from one address to a group of addresses). Automatic creation of Postmaster aliases when new domains are added.

### 3. Admin Tools
- Tool for sending verification (test) emails on behalf of an administrator.
- Broadcast mailing tool to send notifications to all users in a domain or group of domains (processed asynchronously to prevent timeouts).

### 4. Audit Logs
- Logging of every critical action to the backend database.
- Recording *who*, *where* (in which domain), *when*, and *what* action was performed (e.g., "create mailbox", "update domain").
- Regular administrators' access to the log is restricted only to events within their permitted domains.

---

## 🛡 Security

The project was designed with a strong focus on preventing popular web vulnerabilities. Key protection mechanisms:
- **API Authentication:** Almost all endpoints (except login initialization) are protected by `auth.JWTMiddleware`, preventing unauthorized access (protection against unauthorized use or application crashes due to Nil Pointers).
- **IDOR Protection:** Every `POST`, `PUT`, or `DELETE` request for a mailbox, auto-reply, alias, or mailing tool strictly verifies ownership of the target domain via the `domain_admins` database.

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

### Running the Backend
The backend uses configuration, usually from environment files (ENV) or the `internal/config` setup.

```bash
cd backend
go mod tidy
go build -o mailadmin.exe ./main.go
# Start the process
./mailadmin.exe
```

### Running the Frontend
```bash
cd frontend
npm install
npm run dev     # For development
npm run build   # Build for production (outputs to the 'dist' folder)
```

For production use, it is recommended to compile the frontend (`npm run build`) and configure a reverse proxy server (Nginx or Caddy) where `/api` redirects to the Echo backend port, and the root route (`/`) serves static files from `dist`.
