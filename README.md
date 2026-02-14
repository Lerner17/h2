# VPN CLI (Hysteria 2)

CLI-утилита на Go для управления пользователями `auth.userpass` в `hysteria` конфиге.

## Что есть

- Чистая архитектура:
  - `internal/hysteria/domain` - доменная модель и ошибки
  - `internal/hysteria/app` - unit-of-work use cases (каждый в своей директории с `deps/usecase/wire`)
  - `internal/hysteria/infra` - YAML-репозиторий и service control
  - `internal/config` - YAML-конфиг CLI + ENV overrides

## Запуск CLI

```bash
go run ./cmd/cli --help
```

## Makefile

```bash
make deps
make build
make test
sudo make install-cli
```

CLI будет установлен как:
- `/usr/local/bin/vpn-cli`

## CLI

```bash
go run ./cmd/cli add-user --username newuser
```

Конфиг CLI теперь YAML и создается автоматически при первом запуске:
- приоритет пути: `VPN_CONFIG_PATH` → `/etc/vpn/config.yaml` → `~/.config/vpn/config.yaml`
- если файла нет, он будет создан с дефолтами
- ENV (`HYSTERIA_*`) остаются как override поверх YAML

Ротация пароля:

```bash
go run ./cmd/cli rotate-password --username newuser
```

Удаление пользователя:

```bash
go run ./cmd/cli remove-user --username olduser
```

Список пользователей:

```bash
go run ./cmd/cli list-users
```

Инициализация сервера через Ansible:

```bash
go run ./cmd/cli init --host 1.2.3.4 --user root --ssh-key ~/.ssh/id_rsa
```

Playbook лежит в:
- `resources/ansible/hysteria_init.yml`
- `resources/ansible/vars.yaml` (основные переменные для тонкой настройки)

Сценарий:
1. Базовый пользователь просто запускает `init`.
2. Продвинутый пользователь редактирует `resources/ansible/vars.yaml` и запускает тот же `init`.

Интерактивный режим:

```bash
go run ./cmd/cli add-user
```

Для CI/скриптов:

```bash
go run ./cmd/cli add-user --username newuser --yes --output json
```

Пароль всегда генерируется автоматически (криптографический random, только буквы и цифры, без спецсимволов).

URL + QR для подключения:

```bash
go run ./cmd/cli connection --username valera
```

Для рендера QR в терминале нужен `qrencode` в системе.

После успешного добавления пользователя сервис `hysteria` будет перезапущен автоматически.
Порядок:
- если задан `HYSTERIA_RESTART_COMMAND`, выполняется он;
- иначе используется доступный менеджер сервисов (`systemctl`, `service`, на macOS `brew services`).

## Wire

Wire размещен отдельно в каждом use case-пакете (`internal/hysteria/app/*/wire.go`, `wire_gen.go`).
