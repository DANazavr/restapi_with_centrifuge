# Real-Time Notification System

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Centrifugo](https://img.shields.io/badge/Centrifugo-v3.1.0-orange.svg)](https://centrifugal.dev/)

Высокопроизводительная система уведомлений в реальном времени с использованием Go и Centrifugo. Система обеспечивает мгновенную доставку сообщений пользователям с подтверждением прочтения и историей уведомлений.

## Ключевые особенности

- ⚡ Мгновенная доставка сообщений (<100 мс задержки)
- 🔐 Аутентификация JWT с access/refresh токенами
- 📝 История уведомлений с фильтрацией (прочитанные/непрочитанные)
- 📊 Панель администратора для управления уведомлениями
- 🔔 Персонализированные каналы уведомлений
- 📦 Готовые Docker-образы для быстрого развертывания

## Технологический стек

### Бэкенд

| Компонент             | Технология                |
| --------------------- | ------------------------- |
| Язык программирования | Go 1.21+                  |
| gRPC сервисы          | gRPC-Go, Protocol Buffers |
| HTTP Gateway          | Gorilla mux               |
| База данных           | PostgreSQL 14+            |
| Брокер сообщений      | Centrifugo v3             |
| Аутентификация        | JWT                       |
| Контейнеризация       | Docker, Docker Compose    |

### Фронтенд

```plaintext
Vanilla JavaScript, Centrifuge.js, HTML5, CSS3
```

## API Endpoints

### Аутентификация

| Метод | Эндпоинт       | Описание                 |
| ----- | -------------- | ------------------------ |
| POST  | /login         | Вход в систему           |
| POST  | /register      | Регистрация пользователя |
| GET   | /token_refresh | Обновление токена        |

### Уведомления

| Метод | Эндпоинт                | Описание                      |
| ----- | ----------------------- | ----------------------------- |
| GET   | /user/getnotifications  | Получить уведомления          |
| POST  | /user/markasread        | Отметить как прочитанное      |
| POST  | /notification/publish   | Создать уведомление           |
| POST  | /notification/broadcast | Опубликовать через Centrifugo |
