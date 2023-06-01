# Техническое задание
---
## Централизованное управление iptables
### Общие требования

Система представляет собой HTTP API со следующими требованиями к бизнес-логике:

* добавление, удаление правил iptables через единую консоль;
* ведение статистики packets и bytes по каждому правилу;

### Абстрактная схема взаимодействия с системой

Ниже представлена абстрактная бизнес-логика взаимодействия пользователя с системой:
1. На каждый физический или виртуальный сервер устанавливается агент.
2. Администратор добавляет правила для каждого хоста сети.
3. Правила iptables обновляются с заданной периодичностью.
4. Агент отсылает статистику на сервер по каждому правилу.

### Сводное HTTP API
Сервис должен предоставлять следующие HTTP-хендлеры:
* `POST /api/v1/rule` — получение списка цепочек и правил по заданным фильтрам, включая накопленную статистику;
* `POST /api/v1/rule/add` — добавление одного или нескольких правил;
* `POST /api/v1/rule/delete` — удаление одного или нескольких правил;
* `POST /api/v1/rule/stat` — обновление статистики по одному или нескольким правилм;

### (в данный момент требует доработки)
* `POST /api/v1/address-list` — получение списка ipset;
* `POST /api/v1/address-list/add` — добавление одного или нескольких списков ipset;
* `POST /api/v1/address-lists/delete` — удаление одного или нескольких списков ipset;

### Общие ограничения и требования
* агент должен работать с привилегиями root;
* хранилище данных — PostgreSQL;
* клиент может поддерживать HTTP-запросы/ответы со сжатием данных;
* реализация консольного клиента для управления (befwctl)

#### **Получение списка правил**

Хендлер: `POST /api/v1/rule`

Пример консольной команды: 
`befwctl rule show --tenant="host1" --table=filter`
`befwctl rule show --tenant="host1" --table=filter --chain=BEFW`

Формат запроса:

```
POST /api/v1/rule HTTP/1.1
Content-Type: application/json
...
{
    "tenant": "host1",
    "rules": [
        {
            "table": "filter",
            "chain": "BEFW"
        }
    ]
}
```

Возможные коды ответа:
- `200` — успешный запрос;
- `403` — неверный формат запроса;
- `500` — внутренняя ошибка сервера.

 Формат ответа:
```
{
  "tenant": "host1",
  "rules": [
    {
      "id": "f835a787-1f21-45b8-b606-7098673ce236",
      "table": "filter",
      "chain": "BEFW",
      "rulenum": 1,
      "src-address": "192.168.2.0/24",
      "protocol": "tcp",
      "dst-port": "8080",
      "jump": "ACCEPT"
    }
  ],
  "stats": [
    {
      "id": "f835a787-1f21-45b8-b606-7098673ce236",
      "pkts": 10,
      "bytes": 1024
    }
  ]
}
```

#### **Добавление правил**

Хендлер: `POST /api/v1/rule/add`

Пример консольной команды: 
`befwctl rule-add --tenant="host1" --table=filter --chain=BEFW --src-address=192.168.1.100/32 --protocol=tcp --dst-port=8080 --action=ACCEPT`

Формат запроса:

```
POST /api/v1/rule/add HTTP/1.1
Content-Type: application/json
...
{
    "tenant": "host1",
    "rules": [
        {
            "table": "filter",
            "chain": "BEFW",
            "src-address": "192.168.1.100/32",
            "protocol": "tcp",
            "dst-port": "8080",
            "jump": "ACCEPT"
        }
    ]
}
```

Возможные коды ответа:
- `200` — успешный запрос;
- `403` — неверный формат запроса;
- `500` — внутренняя ошибка сервера.

#### **Удаление правил**

Хендлер: `POST /api/v1/rule/delete`

Пример консольной команды: 
`befwctl rule-add --tenant="host1" --id=f835a787-1f21-45b8-b606-7098673ce236`
`befwctl rule-add --tenant="host1" --table=filter --chain=BEFW --rulenum=1`

Формат запроса:
```
POST /api/v1/rule/add HTTP/1.1
Content-Type: application/json
...
{
    "tenant": "host1",
    "rules": [
        {
            "id": "f835a787-1f21-45b8-b606-7098673ce236",
        }
    ]
}
```

```
POST /api/v1/rule/add HTTP/1.1
Content-Type: application/json
...
{
    "tenant": "host1",
    "rules": [
        {
            "table": "filter",
            "chain": "BEFW",
            "rulenum": 1
        }
    ]
}
```

Возможные коды ответа:
- `200` — успешный запрос;
- `403` — неверный формат запроса;
- `500` — внутренняя ошибка сервера.

#### **Обновление счетчиков правил**

Хендлер: `POST /api/v1/rule/stat`

Формат запроса:
```
POST /api/v1/rule/add HTTP/1.1
Content-Type: application/json
...
{
    "tenant": "host1",
    "stats": [
        {
            "id": "f835a787-1f21-45b8-b606-7098673ce236",
            "packets": 10,
            "bytes": 1024
        }
    ]
}
```

Возможные коды ответа:
- `200` — успешный запрос;
- `403` — неверный формат запроса;
- `500` — внутренняя ошибка сервера.

Общее количество запросов информации о начислении не ограничено.

### Конфигурирование агента
Сервис должен поддерживать конфигурирование следующими методами:
- адрес и порт сервера: переменная окружения ОС `ADDRESS` или флаг `-address`
- лог файл `LOG_FILE` или флаг `-log_file`
- уровень логирования `LOG_LEVEL` или флаг `-log_level`
- периодичность запросов к серверу  `POLL_INTERVAL` или флаг `-poll_interval`
- периодичность обвновления счетчиков  `REPORT_INTERVAL` или флаг `-report_interval`
- tenant `TENANT` или флаг `-tenant`

### Конфигурирование сервера
Сервис должен поддерживать конфигурирование следующими методами:
- адрес и порт сервера: переменная окружения ОС `ADDRESS` или флаг `-address`
- лог файл `LOG_FILE` или флаг `-log_file`
- уровень логирования `LOG_LEVEL` или флаг `-log_level`
- DSN для подключения к БД `DATABASE_DSN` или флаг `-database`

### Консольный клиент
```
befwctl --help
Usage:
  befwctl [OPTIONS] <rule-add | rule-del | rule-show>

Application Options:
  -a, --address= set address of firewall server (default: http://127.0.0.1:8080/api/v1) [$ADDRESS]
  -t, --tenant=  set current tenant [$TENANT]

Help Options:
  -h, --help     Show this help message

Available commands:
  rule-add   Add rule
  rule-del   Delete rule
  rule-show  Show rule
```

```
befwctl rule-add --help
Usage:
  befwctl [OPTIONS] rule-add [rule-add-OPTIONS]

[rule-add command options]
          --table=         set table (default: filter)
          --chain=         set chain
          --rulenum=       set rulenum
          --in-interface=  set in-interface
          --out-interface= set out-interface
          --src-address=   set src-address
          --dst-address=   set dst-address
          --protocol=      set protocol
          --src-port=      set src-port
          --dst-port=      set dst-port
          --action=        set action
          --jump=          set jump
          --comment=       set comment
```

```
befwctl rule-show --help
Usage:
  befwctl [OPTIONS] rule-show [rule-show-OPTIONS]

[rule-show command options]
          --id=      set ruleID
          --table=   set table (default: filter)
          --chain=   set chain
          --rulenum= set rulenum
```

```
befwctl rule-del --help
Usage:
  befwctl [OPTIONS] rule-del [rule-del-OPTIONS]

[rule-del command options]
          --id=      set ruleID
          --table=   set table (default: filter)
          --chain=   set chain
          --rulenum= set rulenum
```
