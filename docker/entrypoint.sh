#!/bin/sh

# Запуск бэкенда на Go в фоновом режиме
/usr/local/bin/mailadmin &

# Запуск веб-сервера Nginx на переднем плане
echo "=== Запуск MailAdmin (Nginx + Go) на порту 80 ==="
nginx -g "daemon off;"
