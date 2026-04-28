# Настройка автоответчика (Vacation) для Postfix

Проект **MailAdmin** использует структуру базы данных, которая на 100% совместима с классическим PostfixAdmin. Это позволяет использовать стандартные, проверенные временем скрипты для организации автоответов. 

Данное руководство описывает современный и безопасный метод интеграции автоответчиков в Postfix через `recipient_bcc_maps`. В отличие от старых версий PostfixAdmin (которые добавляли фейковые адреса прямо в таблицу `alias`), наша реализация ничего не «загрязняет» и легко управляется.

## Принцип работы

Мы будем использовать встроенный в Postfix механизм скрытых копий (`recipient_bcc_maps`). Он прозрачно копирует письма для тех пользователей, у которых включен автоответчик, и отправляет их в специальный локальный скрипт `vacation.pl`. Это гарантирует, что пользователь получит исходное письмо, а отправителю автоматически уйдет ответ.

---

### Шаг 1: Подготовка скрипта vacation.pl

Мы используем оригинальный скрипт `vacation.pl` от PostfixAdmin, так как схема таблиц (таблица `vacation`) им полностью поддерживается.

1. **Создайте системного пользователя** для безопасного запуска скрипта:
   ```bash
   useradd -r -c "Virtual vacation" -d /var/spool/vacation -s /sbin/nologin vacation
   mkdir /var/spool/vacation
   chown vacation:vacation /var/spool/vacation
   ```

2. **Скачайте скрипт `vacation.pl`** и поместите его в рабочую директорию:
   ```bash
   wget -O /var/spool/vacation/vacation.pl https://raw.githubusercontent.com/postfixadmin/postfixadmin/master/VIRTUAL_VACATION/vacation.pl
   chmod +x /var/spool/vacation/vacation.pl
   chown vacation:vacation /var/spool/vacation/vacation.pl
   ```

3. **Отредактируйте параметры подключения** внутри `vacation.pl` (откройте файл редактором):
   ```perl
   our $db_type = 'Pg'; # или 'mysql' в зависимости от вашей базы
   our $db_username = 'mailadmin_user';
   our $db_password = 'your_password';
   our $db_name     = 'mailadmin';
   our $db_host     = '127.0.0.1';
   our $vacation_domain = 'autoreply.local'; # Служебный фейковый домен для роутинга
   ```

---

### Шаг 2: Настройка Postfix (BCC Routing)

Теперь необходимо научить Postfix понимать, когда нужно отправлять копию скрипту автоответчика. Для этого создадим SQL-запрос.

1. **Создайте файл** `/etc/postfix/mysql_vacation_bcc.cf`:
   ```ini
   user = mailadmin_user
   password = your_password
   hosts = 127.0.0.1
   dbname = mailadmin
   # Запрос проверяет флаг active и укладывается ли текущая дата в нужные рамки отпуска
   query = SELECT CONCAT(email, '@autoreply.local') FROM vacation WHERE email='%s' AND active=1 AND activefrom <= NOW() AND activeuntil >= NOW();
   ```

2. **Подключите эту карту** к Postfix. Откройте `/etc/postfix/main.cf` и добавьте (или дополните, если уже есть):
   ```ini
   recipient_bcc_maps = mysql:/etc/postfix/mysql_vacation_bcc.cf
   ```

*(Как это работает: если для ящика активно правило автоответа, запрос вернет `user@domain.com@autoreply.local`, и Postfix сделает на него скрытую копию (BCC)).*

---

### Шаг 3: Настройка Transport для служебного домена

Нам нужно перенаправить всю почту для домена `autoreply.local` в скрипт `vacation.pl`.

1. **В `/etc/postfix/main.cf` добавьте транспортную карту** (если ее еще нет):
   ```ini
   transport_maps = hash:/etc/postfix/transport
   ```

2. **Откройте файл** `/etc/postfix/transport` и добавьте туда строку:
   ```text
   autoreply.local    vacation:
   ```

3. **Скомпилируйте карту**:
   ```bash
   postmap /etc/postfix/transport
   ```

---

### Шаг 4: Регистрация сервиса в master.cf

Postfix должен понимать, что такое транспорт `vacation:`.

Откройте `/etc/postfix/master.cf` и добавьте новый сервис в самый конец файла:

```text
vacation    unix  -       n       n       -       -       pipe
  flags=Rq user=vacation argv=/var/spool/vacation/vacation.pl -f ${sender} -- ${recipient}
```

---

### Шаг 5: Применение настроек

Для применения всех изменений перезапустите сервис Postfix:
```bash
systemctl restart postfix
```

---

## Жизненный цикл автоответа

1. Ваш клиент (или кто-то сторонний) пишет письмо на `joe@example.com`.
2. Postfix при обработке адресата проверяет таблицу `recipient_bcc_maps`. 
3. Если Джо включил в **MailAdmin** форму отпуска и текущая дата совпадает, запрос возвращает адрес. Postfix делает копию для адресата `joe@example.com@autoreply.local`.
4. Эта копия попадает в `transport_maps`, где домен `autoreply.local` жестко направлен в локальный сервис `vacation:`.
5. Скрипт `vacation.pl` получает эту копию. Он:
   - Проверяет в таблице `vacation`, не истек ли `interval_time` (чтобы не зациклить переписку и не отправить 100 отбивок на 100 писем от одного человека). 
   - Берет из базы настроенные `subject` и `body`.
   - Отправляет ответное письмо обратно оригинальному отправителю.
6. В это же самое время оригинальное письмо безупречно доставляется Джо в его ящик силами Dovecot.
