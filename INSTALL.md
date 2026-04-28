# Инструкция по установке MailAdmin

**MailAdmin** — современная панель управления почтовым сервером на базе связки Postfix + MariaDB (совместима со схемой данных PostfixAdmin).

Данный документ содержит инструкции по развертыванию панели как внутри Docker-контейнера, так и в виде классического системного сервиса.

---

## Вариант 1. Установка через Docker Compose (Рекомендуемый)

Самый быстрый и удобный способ развертывания — использование Docker Compose. Все компоненты (фронтенд и бэкенд) собираются в единый контейнер.

### Шаг 1. Настройка окружения

В корневой директории проекта находится файл `docker-compose.yml`. Отредактируйте его параметры при необходимости:

```yaml
version: '3.8'

services:
  mailadmin:
    build: .
    container_name: mailadmin
    restart: always
    ports:
      - "8081:80" # Внешний порт панели
    environment:
      - DB_DSN=postfix:password@tcp(127.0.0.1:3306)/postfix?charset=utf8mb4&parseTime=True&loc=Local
      - JWT_SECRET=your_secure_random_jwt_secret_key_here
      - LISTEN_ADDR=:8080
    volumes:
      - /var/log/mail.log:/var/log/mail.log:ro
      - /data/mail:/data/mail
      - /data/sieve:/data/sieve
```

Обязательно обновите данные для подключения к БД (`DB_DSN`) и задайте безопасный `JWT_SECRET`.

### Шаг 2. Запуск

Выполните команду для сборки и старта:

```bash
docker compose up -d --build
```

После этого панель будет доступна по адресу `http://ваш-ip:8081`.


---

## Вариант 2. Ручная установка (Standalone)

Если почтовый сервер работает прямо на хост-машине (без Docker).

### Требования
* **Go** версии 1.20 или выше
* **Node.js** версии 16+ и пакетный менеджер **npm**
* Настроенный веб-сервер **Nginx**

### Шаг 1. Клонирование репозитория
```bash
git clone https://github.com/kirpicheff/mailadmin.git
cd mailadmin
```

### Шаг 2. Настройка и сборка бэкенда
1. Перейдите в директорию бэкенда:
   ```bash
   cd backend
   ```
2. Создайте и отредактируйте файл `.env`:
   ```bash
   cp .env.example .env  # при наличии, или создайте вручную
   ```
   **Параметры `.env`:**
   * `DB_DSN`: строка подключения к БД (например, `postfix:password@tcp(127.0.0.1:3306)/postfix`)
   * `JWT_SECRET`: секретная строка для подписи токенов
   * `LISTEN_ADDR`: порт прослушивания (по умолчанию `:8080`)

3. Скомпилируйте приложение:
   ```bash
   go mod tidy
   go build -o mailadmin ./main.go
   ```

4. (Опционально) Настройте системную службу `systemd` для автозапуска:
   Создайте файл `/etc/systemd/system/mailadmin.service`:
   ```ini
   [Unit]
   Description=MailAdmin Go Backend
   After=network.target mariadb.service

   [Service]
   Type=simple
   WorkingDirectory=/opt/mailadmin/backend
   ExecStart=/opt/mailadmin/backend/mailadmin
   Restart=always
   Environment=PORT=8080

   [Install]
   WantedBy=multi-user.target
   ```

### Шаг 3. Сборка фронтенда
1. Перейдите в папку фронтенда:
   ```bash
   cd ../frontend
   ```
2. Установите зависимости и соберите статику:
   ```bash
   npm install
   # Сборка в подкаталог /mailadmin/
   VITE_API_URL=/mailadmin/api npm run build -- --base=/mailadmin/
   ```
   Готовые файлы появятся в директории `frontend/dist`.

### Шаг 4. Настройка Nginx

Для корректной работы объединяем статику и API через Reverse Proxy:

```nginx
server {
    listen 9999 ssl;
    server_name mail.example.com;

    # SSL настройки опущены...

    # Статические файлы панели MailAdmin
    location /mailadmin {
        alias /var/www/mailadmin;
        index index.html;
        try_files $uri $uri/ /mailadmin/index.html;
    }

    # API Бэкенда (Go)
    location /mailadmin/api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

После перезапуска Nginx панель станет доступна по адресу `https://ваш-домен:9999/mailadmin/`.
