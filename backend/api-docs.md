# Portfolio Tracker API Specification

Полное техническое описание всех эндпоинтов REST API для системы трекинга портфелей, безрисковых ставок, леджера транзакций и аналитических модулей.

---

## Общие параметры

* **Базовый URL:** `http://localhost:8080`
* **Формат данных:** `application/json`
* **Авторизация:** Защищенные эндпоинты требуют передачи JWT-токена в заголовке `Authorization: Bearer <token>`.

---

## 1. Аутентификация и пользователи (`/auth`)

### Регистрация пользователя
* **Эндпоинт:** `POST /auth/register`
* **Тип запроса:** Публичный
* **Тело запроса (JSON):**
  ```json
  {
    "login": "user123",
    "email": "user@example.com",
    "password": "secure_password"
  }
  ```
* **Ответ (`201 Created`):**
  ```json
  {
    "status": "success",
    "message": "Successfully registered user"
  }
  ```

### Авторизация (Вход)
* **Эндпоинт:** `POST /auth/login`
* **Тип запроса:** Публичный
* **Тело запроса (JSON):**
  ```json
  {
    "login": "user123",
    "password": "secure_password"
  }
  ```
* **Ответ (`200 OK`):**
  ```json
  {
    "status": "success",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```

### Запрос на сброс пароля
* **Эндпоинт:** `POST /auth/request-password-reset`
* **Тип запроса:** Публичный
* **Тело запроса (JSON):**
  ```json
  {
    "email": "user@example.com"
  }
  ```
* **Ответ (`200 OK`)**

### Проверка одноразового кода (OTC)
* **Эндпоинт:** `POST /auth/verify-otc`
* **Тип запроса:** Публичный
* **Тело запроса (JSON):**
  ```json
  {
    "email": "user@example.com",
    "code": "123456"
  }
  ```
* **Ответ (`200 OK`)**

### Фиксация нового пароля
* **Эндпоинт:** `POST /auth/reset-password`
* **Тип запроса:** Публичный
* **Тело запроса (JSON):**
  ```json
  {
    "email": "user@example.com",
    "code": "123456",
    "new_password": "new_secure_password"
  }
  ```
* **Ответ (`200 OK`)**

---

## 2. Безрисковые ставки (`/risk-free-rates`)

### Получение всех ставок
* **Эндпоинт:** `GET /risk-free-rates`
* **Тип запроса:** Защищенный
* **Ответ (`200 OK`):**
  ```json
  [
    {
      "id": 1,
      "name": "USD-Tres Oct-2026",
      "value": 0.045
    }
  ]
  ```

### Получение ставки по ID
* **Эндпоинт:** `GET /risk-free-rates/{id}`
* **Тип запроса:** Защищенный
* **Ответ (`200 OK`):**
  ```json
  {
    "id": 1,
    "name": "USD-Tres Oct-2026",
    "value": 0.045
  }
  ```

### Создание ставки
* **Эндпоинт:** `POST /risk-free-rates`
* **Тип запроса:** Защищенный
* **Тело запроса (JSON):**
  ```json
  {
    "name": "USD-Tres Oct-2026",
    "value": 0.045
  }
  ```
* **Ответ (`201 Created`):**
  ```json
  {
    "status": "success",
    "message": "Risk free rate with 1 id created successfully"
  }
  ```

---

## 3. Инвестиционные портфели (`/portfolios`)

### Получение всех портфелей пользователя
* **Эндпоинт:** `GET /portfolios`
* **Тип запроса:** Защищенный
* **Описание:** Возвращает массив всех портфелей, принадлежащих текущему авторизованному пользователю.
* **Ответ (`200 OK`):**
  ```json
  [
    {
      "id": 1,
      "user_id": 1,
      "risk_free_rate_id": 1,
      "name": "Основной криптопортфель",
      "created_at": "2026-10-29T12:00:00Z"
    }
  ]
  ```

### Создание портфеля
* **Эндпоинт:** `POST /portfolios`
* **Тип запроса:** Защищенный
* **Тело запроса (JSON):**
  ```json
  {
    "name": "Основной криптопортфель",
    "risk_free_rate_id": 1
  }
  ```
* **Ответ (`201 Created`):**
  ```json
  {
    "status": "success",
    "message": "Portfolio with 1 id created successfully"
  }
  ```

### Обновление портфеля
* **Эндпоинт:** `PUT /portfolios/{id}`
* **Тип запроса:** Защищенный
* **Описание:** Обновляет имя портфеля и базовую глобальную безрисковую ставку (только для владельца).
* **Тело запроса (JSON):**
  ```json
  {
    "name": "Новое имя портфеля",
    "risk_free_rate_id": 2
  }
  ```
* **Ответ (`200 OK`):**
  ```json
  {
    "status": "success",
    "message": "Portfolio updated successfully"
  }
  ```

### Удаление портфеля
* **Эндпоинт:** `DELETE /portfolios/{id}`
* **Тип запроса:** Защищенный
* **Описание:** Удаляет портфель и все каскадно связанные элементы и транзакции (только для владельца).
* **Ответ (`200 OK`):**
  ```json
  {
    "status": "success",
    "message": "Portfolio deleted successfully"
  }
  ```

---

## 4. Элементы портфеля и аллокация (`/portfolios/{id}/items` и `/portfolio-items`)

### Добавление элемента (актива) в конкретный портфель
* **Эндпоинт:** `POST /portfolios/{id}/items`
* **Тип запроса:** Защищенный
* **Описание:** Поля `ticker` и `risk_free_rate_id` являются взаимоисключающими на уровне CHECK-constraint базы данных.
* **Тело запроса (JSON):**
  ```json
  {
    "ticker": "BTCUSDT",
    "risk_free_rate_id": null,
    "total_amount": 2.2
  }
  ```
* **Ответ (`201 Created`):**
  ```json
  {
    "status": "success",
    "message": "Portfolio item with 5 id created successfully"
  }
  ```

### Удаление элемента из портфеля по ID элемента
* **Эндпоинт:** `DELETE /portfolio-items/{id}`
* **Тип запроса:** Защищенный
* **Описание:** Удаляет конкретную позицию из портфеля по её уникальному идентификатору (проверяется владение родительским портфелем).
* **Ответ (`200 OK`):**
  ```json
  {
    "status": "success",
    "message": "Portfolio item deleted successfully"
  }
  ```

---

## 5. Леджер изменений и транзакции (`/portfolios/{id}/transactions`)

### Получение истории транзакций портфеля
* **Эндпоинт:** `GET /portfolios/{id}/transactions`
* **Тип запроса:** Защищенный
* **Ответ (`200 OK`):**
  ```json
  [
    {
      "id": 42,
      "portfolio_id": 1,
      "portfolio_item_id": 5,
      "type": "deposit",
      "amount": 0.25,
      "processed_at": "2026-10-29T14:55:00Z"
    }
  ]
  ```

### Создание транзакции (Атомарное изменение объема)
* **Эндпоинт:** `POST /portfolios/{id}/transactions`
* **Тип запроса:** Защищенный
* **Описание:** Выполняет операцию в рамках `pgx.Tx`. Изменяет `total_amount` позиции в `portfolio_items` и пишет неизменяемый лог.
* **Тело запроса (JSON):**
  ```json
  {
    "portfolio_item_id": 5,
    "type": "deposit",
    "amount": 0.25,
    "price_usd": 81463.8
  }
  ```
* **Ответ (`201 Created`):**
  ```json
  {
    "status": "success",
    "message": "Transaction recorded and balance updated successfully. Transaction ID: 42"
  }
  ```

---

## 6. Математический анализ и метрики (`/portfolios/{id}/...`)

### 6.1 Риски и волатильность (`GET /portfolios/{id}/risks`)
* **Описание:** Возвращает квантили VaR Паркинсона для всего портфеля и риск-метрики (MCR, CCTR) для каждого актива, деаннуализированные до дневного масштаба.
* **Ответ (`200 OK`):**
  ```json
  {
    "portfolioVaR_quantiles": {
      "Mu": 1587.6963122610846,
      "Q90": 2749.991785290907,
      "Q95": 3382.458371161213,
      "Q99": 4987.679675784959
    },
    "partial_risks": [
      ["BTCUSDT", [2.2284, 843.298], [3.8916, 1472.735], [4.8079, 1819.495], [7.1484, 2705.212], 38.942, 1.748],
      ["ETHUSDT", [3.1904, 744.397], [5.4743, 1277.256], [6.6988, 1562.963], [9.7826, 2282.467], 67.125, 64.112]
    ],
    "volatility_drag": 24.489,
    "diversification_ratio": 1.005,
    "effective_number_of_assets": 1.093
  }
  ```
* **Контракт кортежа в `partial_risks`:** `[Ticker, [Mu%, MuUSD], [Q90%, Q90USD], [Q95%, Q95USD], [Q99%, Q99USD], daily_mcr, daily_cctr]`

### 6.2 Текущая структура портфеля (`GET /portfolios/{id}/structure`)
* **Ответ (`200 OK`):**
  ```json
  {
    "portfolio_structure": [
      ["ETHUSDT", 2356.9, 10, 23569, 12.57, 450.25],
      ["BTCUSDT", 81463.8, 2, 162927.6, 86.89, 1200.0],
      ["USD-Tres Oct-2026", 1, 1000, 1000, 0.53, 0]
    ]
  }
  ```
* **Контракт кортежа в `portfolio_structure`:** `[Ticker, CurrentPrice, Amount, TotalUSDValue, WeightPercentage, PnLUSD]`

### 6.3 Доходность и кривая FOMO (`GET /portfolios/{id}/yield`)
* **Описание:** Рассчитывает исторические таймлайны доходностей на основе слепов крон-воркера, вычисляет ковариацию/корреляцию с FOMO (BTCUSDT) и выдает динамический Топ-3 активов по P&L.
* **Ответ (`200 OK`):**
  ```json
  {
    "p&l": {
      "portfolio": [0, 12, -20, 203],
      "portfolio_FOMO": [0, 15, -45, 310],
      "risk_free": [0, 1, 2, 3.5]
    },
    "yield_top_3": {
      "1": ["BTCUSDT", 1200.0, 4.8],
      "2": ["ETHUSDT", 450.0, 2.1]
    },
    "portfolio_structure": [...],
    "sharpes_ratio": 0.6298,
    "sortino_ratio": 0.45,
    "cov_fomo_pnl": 0.124,
    "corr_fomo_pnl": 0.85
  }
  ```

### 6.4 Сводный экран управления (`GET /portfolios/{id}/overview`)
* **Описание:** Главная панель. Рассчитывает интегральный уровень риска `risks_metrick` (`0`–`100`), список триггеров аномалий (`reasons`) и краткосрочный/долгосрочный скоринг P&L.
* **Ответ (`200 OK`):**
  ```json
  {
    "p&l_history": {
        "portfolio": [0, 12, -20, 203],
        "portfolio_FOMO": [0, 15, -45, 310],
        "risk_free": [0, 1, 2, 3.5]
    },
    "risks": {
        "risks_metrick": 52,
        "reasons": ["volatility drag", "High VaR", "Low Diversification"]
    },
    "p&l": {
        "total_pnl": [-405.234, -25.434],
        "daily_pnl": [200.544, 6.0],
        "historical_pnl": [450.234, 3.07]
    }
    }
  ```
## 7. Сервисный мониторинг (`/health`)

### Проверка работоспособности системы
* **Эндпоинт:** `GET /health`
* **Тип запроса:** Публичный
* **Описание:** Используется внешними системами оркестрации, в частности Docker или мониторинга для верификации того, что инстанс приложения успешно запущен и обрабатывает входящие HTTP-запросы.
* **Ответ (`200 OK`):**
  ```json
  {
    "status": "up"
  }
  ```