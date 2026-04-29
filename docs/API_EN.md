# MailAdmin API Documentation

This document provides a description of all available API endpoints for the MailAdmin control panel.

## General Information

- **Base URL**: `/api`
- **Data Format**: JSON
- **Authentication**: Most endpoints require a JWT token passed in the `Authorization` header.
  Format: `Authorization: Bearer <token>`

---

## 1. Authentication (`/api/auth`)

### POST `/api/auth/login`
User authentication and token retrieval.

- **Authentication**: Not required.
- **Request Body**:
  ```json
  {
    "username": "admin@example.com",
    "password": "your_password"
  }
  ```
- **Successful Response (200 OK)**:
  ```json
  {
    "access_token": "eyJhbGciOi...",
    "user": {
      "username": "admin@example.com",
      "superadmin": true
    }
  }
  ```
  *Note*: An HttpOnly Cookie `refreshToken` is also set.

### POST `/api/auth/refresh`
Refresh Access Token using Refresh Token.

- **Authentication**: Uses Cookie `refreshToken`.
- **Successful Response (200 OK)**:
  ```json
  {
    "access_token": "eyJhbGciOi..."
  }
  ```

### POST `/api/auth/logout`
User logout.

- **Authentication**: Not required. Clears the `refreshToken` cookie.
- **Successful Response**: `200 OK` (no body).

### GET `/api/auth/me`
Get information about the current user.

- **Authentication**: JWT required.
- **Successful Response (200 OK)**: User data (Claims).

### POST `/api/auth/change-password`
Change the password of the current user.

- **Authentication**: JWT required.
- **Request Body**:
  ```json
  {
    "new_password": "new_secure_password"
  }
  ```
- **Successful Response (200 OK)**:
  ```json
  {
    "message": "password changed successfully"
  }
  ```

---

## 2. Administrator Management (`/api/admins`)
*Available only to SuperAdmin.*

### GET `/api/admins`
Retrieve a list of all administrators.

- **Authentication**: JWT required (SuperAdmin).
- **Successful Response (200 OK)**: Array of administrator objects.

### GET `/api/admins/:username`
Retrieve details of a specific administrator and the domains they manage.

- **Authentication**: JWT required (SuperAdmin).
- **Successful Response (200 OK)**:
  ```json
  {
    "admin": { ... },
    "domains": ["example.com", "test.com"]
  }
  ```

### POST `/api/admins`
Create a new administrator.

- **Authentication**: JWT required (SuperAdmin).
- **Request Body**:
  ```json
  {
    "username": "newadmin@example.com",
    "password": "secure_password",
    "superadmin": false,
    "active": true,
    "domains": ["example.com"]
  }
  ```
- **Successful Response (201 Created)**: Created administrator object.

### PUT `/api/admins/:username`
Edit an administrator.

- **Authentication**: JWT required (SuperAdmin).
- **Request Body**:
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
- **Successful Response**: `200 OK` (no body).

### DELETE `/api/admins/:username`
Delete an administrator.

- **Authentication**: JWT required (SuperAdmin).
- **Successful Response**: `204 No Content`.

---

## 3. Domain Management (`/api/domains`)

### GET `/api/domains`
Retrieve a list of domains with their statistics.

- **Authentication**: JWT required.
- **Query Parameters**:
  - `search` (string, optional) — search by name or description.
  - `active` (bool, optional) — filter by status (`true`/`false`).
- **Successful Response (200 OK)**:
  ```json
  [
    {
      "domain": "example.com",
      "description": "Main domain",
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
Create a new domain.
*Note: RFC aliases (postmaster, abuse, hostmaster) are created automatically.*

- **Authentication**: JWT required (SuperAdmin).
- **Request Body**:
  ```json
  {
    "domain": "newdomain.com",
    "description": "Description",
    "aliases": 0,
    "mailboxes": 0,
    "maxquota": 0,
    "quota": 0,
    "transport": "virtual",
    "backupmx": false,
    "active": true
  }
  ```
- **Successful Response (201 Created)**: Created domain object.

### PUT `/api/domains/:id`
Edit a domain.

- **Authentication**: JWT required (SuperAdmin).
- **Request Body**: Same fields as creation (except `domain`).
- **Successful Response (200 OK)**: Updated domain object.

### DELETE `/api/domains/:id`
Cascading deletion of a domain and all associated mailboxes, aliases, and settings.

- **Authentication**: JWT required (SuperAdmin).
- **Successful Response**: `204 No Content`.

### GET `/api/domains/:id/dns`
Generate recommended DNS records (MX, SPF, DKIM, DMARC) for the domain.

- **Authentication**: JWT required (Domain Access).
- **Successful Response (200 OK)**:
  ```json
  {
    "domain": "example.com",
    "spf": "v=spf1 ip4:1.2.3.4 mx a -all",
    "dkim": "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
    "dmarc": "v=DMARC1; p=quarantine; rua=mailto:postmaster@example.com"
  }
  ```

---

## 4. Mailbox Management (`/api/mailboxes`)

### GET `/api/mailboxes`
Retrieve a list of mailboxes.

- **Authentication**: JWT required.
- **Query Parameters**:
  - `domain` (string) — required if `search` is not specified.
  - `search` (string, optional) — search by username or name.
  - `page` (int, default 1).
  - `limit` (int, default 50).
  - `active` (bool, optional).
- **Response Headers**: `X-Total-Count` — total count.
- **Successful Response (200 OK)**: Array of `Mailbox` objects with `quota_used` and `messages` fields.

### POST `/api/mailboxes`
Create a mailbox.
*Note: A mirror alias is created automatically.*

- **Authentication**: JWT required (Domain Access).
- **Request Body**:
  ```json
  {
    "username": "user@example.com",
    "password": "secure_password",
    "name": "John Doe",
    "quota": 1024,
    "active": true,
    "phone": "+123456789",
    "email_other": "backup@gmail.com"
  }
  ```
- **Successful Response (201 Created)**: Created object.

### PUT `/api/mailboxes/:username`
Edit a mailbox.

- **Authentication**: JWT required (Domain Access).
- **Request Body**:
  ```json
  {
    "password": "optional_new_password",
    "name": "New Name",
    "quota": 2048,
    "active": true,
    "phone": "+123456789",
    "email_other": "backup@gmail.com"
  }
  ```
- **Successful Response (200 OK)**: Updated object.

### DELETE `/api/mailboxes/:username`
Delete a mailbox and associated aliases/vacation auto-replies.

- **Authentication**: JWT required (Domain Access).
- **Successful Response**: `204 No Content`.

### Batch Operations

#### POST `/api/mailboxes/batch/create`
Mass create mailboxes with a standard password.
- **Request Body**:
  ```json
  {
    "domain": "example.com",
    "prefixes": ["user1", "user2"],
    "password": "common_password",
    "quota": 1024,
    "active": true
  }
  ```
- **Response (201 Created)**: `{"created": 2}`

#### POST `/api/mailboxes/batch/delete`
Mass delete.
- **Request Body**: `{"usernames": ["user1@example.com", "user2@example.com"]}`
- **Response**: `204 No Content`

#### POST `/api/mailboxes/batch/status`
Mass change status (active/blocked).
- **Request Body**: `{"usernames": ["user1@example.com"], "active": false}`
- **Response**: `204 No Content`

---

## 5. Alias Management (`/api/aliases`)

### GET `/api/aliases`
List forwarding aliases.

- **Authentication**: JWT required.
- **Query Parameters**: `domain`, `search`, `page`, `limit`.

### POST `/api/aliases`
Create an alias.
- **Body**: `{"address": "info@example.com", "goto": "user@example.com,boss@example.com", "domain": "example.com", "active": true}`
- **Response**: `201 Created`.

### PUT `/api/aliases/:address`
Edit an alias.
- **Body**: `{"goto": "new@example.com", "active": true}`

### DELETE `/api/aliases/:address`
Delete an alias.

### Domain Aliases (`/api/aliases/domain-aliases`)
*SuperAdmin only.*
- `GET /api/aliases/domain-aliases?domain=...`
- `POST /api/aliases/domain-aliases` (Body: `AliasDomain` object)
- `DELETE /api/aliases/domain-aliases/:alias_domain`

---

## 6. Logs and Statistics

### GET `/api/logs`
View audit logs of administrator actions.
- **Parameters**: `page`, `limit` (max 500), `domain`.

### GET `/api/stats/dashboard`
Summary statistics for the dashboard (number of domains, mailboxes, quota used, recent logs).

---

## 7. System Monitoring (`/api/system`)
*Available only to SuperAdmin.*

- `GET /api/system/health` — Status of RAM, CPU, Disk, services (Postfix, Dovecot, MySQL, etc.).
- `GET /api/system/queue` — View Postfix mail queue.
- `POST /api/system/queue/flush` — Force mail queue flush.
- `DELETE /api/system/queue/:id` — Delete a message from the queue (`id` or `all`).
- `GET /api/system/logs?lines=200&search=...` — View system logs `/var/log/mail.log`.
- `GET /api/system/fail2ban` — List banned IPs.
- `DELETE /api/system/fail2ban/unban?ip=...&jail=...` — Unban an IP.

---

## 8. Tools and Vacation Auto-Reply

- `GET /api/tools/check-mx/:domain` — Quick check of MX records.
- `POST /api/tools/send-email` — Send a test email.
- `POST /api/tools/broadcast` — Broadcast announcement to all mailboxes in a domain.

### Vacation Auto-Reply
- `GET /api/mailboxes/:username/vacation`
- `PUT /api/mailboxes/:username/vacation`
  - Body: `{"subject": "...", "body": "...", "active": true, "activefrom": "...", "activeuntil": "...", "interval_time": 1}`

---

## 9. Sieve Filters (`/api/sieve`)

- `GET /api/sieve/:username` — Retrieve rules (including `GLOBAL` for superadmin).
- `POST /api/sieve/:username` — Save rules (JSON).
- `POST /api/sieve/:username/import` — Import raw `.sieve` file from server to web interface.
