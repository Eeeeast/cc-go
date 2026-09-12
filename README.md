# compiler-template-go

Шаблон проекта курса компиляторов для Go

## Установка окружения

1. Установите Go: https://go.dev/doc/install
2. Установите golangci-lint: https://golangci-lint.run/docs/welcome/install/local/

## Сборка

Сборка с помощью GNU make:

```bash
# Обновление пакетов, сборка и запуск тестов
make

# Только обновление пакетов
make tidy

# Только сборка
make build

# Только запуск тестов
make test

# Автоформатирование
make format

# Проверка линтером
make lint

# Автоматическое исправление замечаний линтера
make fix
```

## Анализ покрытия кода тестами

Есть два способа генерации HTML отчёта:

```bash
# Создаёт простой HTML-отчёт с помощью go tool cover
# Путь к отчёту: coverage/report.html
make coverage

# Создаёт красивый HTML-отчёт с помощью ReportGenerator
make coverage-rich
# Путь к отчёту: coverage/report/index.html
```

Второй способ требует установки ReportGenerator:

1. Установите .NET 10: https://dotnet.microsoft.com/en-us/download/dotnet/10.0
2. Установите ReportGenerator: `dotnet tool install -g dotnet-reportgenerator-globaltool`

## Структура проекта на Go

Ликбез по структуре каталогов:

1. `bin/` — артефакты сборки
2. `cmd/` — исходный код CLI
3. `internal/` — основной каталог исходного кода

Файл `go.mod` содержит описание проекта и список зависимостей. Для добавления зависимости нужно:

1. Использовать нужный пакет в исходном коде
2. Затем выполнить `make tidy`
3. В конце зафиксировать изменения в Git

## Статический анализ

Шаблон содержит цели make, использующие статический анализатор golangci-lint:

1. `make lint` проверяет код
2. `make format` применяет автоформатирование
3. `make fix` применяет автоматические исправления

Вы можете настроить golangci-lint с помощью конфига — но только при условии, что это не снизит качество кода.
