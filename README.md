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
