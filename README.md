# radar-bot

Telegram-бот «Радара»: привязка аккаунта и уведомления — срочные инсайты и сводка.

## Запуск

```sh
cp .env.example .env   # вписать BOT_TOKEN от @BotFather
set -a && . ./.env && set +a
make run
```

В PowerShell вместо `set -a`:

```powershell
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^#=]+?)\s*=\s*(.*)$') { Set-Item "env:$($matches[1])" $matches[2] }
}
go run ./cmd/bot
```

Нужен Go 1.24+.

## Переменные окружения

| Переменная  | По умолчанию            | Что это                                             |
| ----------- | ----------------------- | --------------------------------------------------- |
| `BOT_TOKEN` | —                       | токен бота от @BotFather, обязателен                |
| `SITE_URL`  | `http://localhost:5173` | адрес сайта, на него ведут ссылки в боте            |
| `LOG_LEVEL` | `info`                  | `debug`, `info`, `warn`, `error`                    |
| `DEV_ADDR`  | —                       | адрес для `POST /dev/deliveries`, пусто — выключено |

Кнопки «Открыть в Радаре» Telegram принимает только с публичным адресом. Если `SITE_URL`
указывает на `localhost`, бот ставит ссылку прямо в текст сообщения.

Договор с бэкендом — `docs/bot-api.md` в репозитории `radar-web`.

## Без бэкенда

Пока бэкенда нет, бот ходит в заглушку `internal/backend/fake.go`. Проверить привязку можно
ссылками `https://t.me/<бот>?start=valid-code`, `expired-code` и `used-code`.

Прислать себе уведомление (нужен `DEV_ADDR=localhost:8081` и привязанный чат):

```sh
curl -X POST localhost:8081/dev/deliveries                         # срочный инсайт
curl -X POST localhost:8081/dev/deliveries -d '{"kind":"digest"}'  # сводка
```

```powershell
Invoke-RestMethod -Method Post http://localhost:8081/dev/deliveries
Invoke-RestMethod -Method Post http://localhost:8081/dev/deliveries -Body '{"kind":"digest"}'
```

В теле можно передать свой `insight` или `digest` в формате `docs/bot-api.md`, а `chatId` —
чтобы отправить в один чат. Без `chatId` доставка уходит во все привязанные чаты. Бот
забирает очередь раз в 10 секунд.
