#!/bin/sh

# Запуск системного агента (Privileged) от имени root
/usr/local/bin/mailadmin --agent &

# Запуск веб-узла (Unprivileged) от имени пользователя mailadmin
su-exec mailadmin /usr/local/bin/mailadmin --web &

# Запуск веб-сервера Nginx на переднем плане
echo "=== Запуск MailAdmin (Nginx + Agent + Web) ==="
nginx -g "daemon off;"

