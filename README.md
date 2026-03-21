# lib-validation

[![CI](https://github.com/selfshop-dev/lib-validation/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-validation/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-validation/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-validation)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-validation)](https://goreportcard.com/report/github.com/selfshop-dev/lib-validation)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-validation)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-validation)](LICENSE)

Структурированные, field-level ошибки валидации для Go-сервисов — без привязки к фреймворку, без внешних зависимостей в production-коде.

## Overview

Большинство validation-библиотек либо возвращают одну строку с ошибкой, либо требуют глубокой интеграции с фреймворком. `lib-validation` даёт машиночитаемый [`*Error`](error.go), который несёт **все field-ошибки сразу** — HTTP-хендлер сериализует их в одном ответе, а клиент переключается по стабильным [`Code`](code.go)-значениям без парсинга строк.

Пакет намеренно узкий по scope: он определяет **словарь ошибок**, а не правила валидации. Предикаты пишешь ты — пакет даёт единообразный способ их сообщать.

### Installation

```bash
go get -u github.com/selfshop-dev/lib-validation
```

### Быстрый старт

```go
import validation "github.com/selfshop-dev/lib-validation"

func ValidateCreateUser(req CreateUserRequest) error {
    c := validation.NewCollector("invalid user")
    c.Check(req.Name != "", validation.Required("name"))
    c.Check(len(req.Name) <= 100, validation.TooLong("name", 100))
    c.Check(isEmail(req.Email), validation.Invalid("email", "must be a valid address"))
    c.Check(req.Age >= 18, validation.OutOfRange("age", 18, 120))
    c.Merge("address", validateAddress(req.Address))
    return c.Err()
}
```

На стороне получателя — например, в HTTP-хендлере:

```go
if err := svc.CreateUser(ctx, req); err != nil {
    if ve, ok := validation.As(err); ok {
        // ve.Fields — все FieldError-ы в порядке добавления
        // ve.Summary — человекочитаемое описание верхнего уровня
    }
}
```

## Code

[`Code`](code.go) — стабильная, машиночитаемая строка. Клиенты переключаются по ней — воспринимай как версионированный API-контракт. Переименование опубликованного `Code` — breaking change.

| Code | Значение |
|---|---|
| `required` | Поле отсутствует или имеет zero-value |
| `invalid` | Поле присутствует, но не проходит правило |
| `too_long` | Превышает максимальную длину |
| `too_short` | Меньше минимальной длины |
| `out_of_range` | Числовое значение вне диапазона `[min, max]` |
| `conflict` | Значение конфликтует с существующим состоянием |
| `immutable` | Поле нельзя изменить после создания |
| `type_mismatch` | Значение неправильного типа |
| `unknown` | Нераспознанный ключ (strict-режим конфига) |

## FieldError

[`FieldError`](field_error.go) описывает одну ошибку валидации. `Field` использует dot-notation — `"user.address.zip_code"` — что соответствует как путям в JSON body, так и ключам конфига. Пустой `Field` означает entity-level ошибку, не привязанную к конкретному полю; для таких случаев используй [`Entity`](builders.go).

Никогда не выставляй `Value` для чувствительных полей (пароли, токены). Используй [`WithValue`](field_error.go) явно только когда передача значения безопасна.

## Error

[`*Error`](error.go) агрегирует все `FieldError`-ы, собранные за один validation pass. Реализует интерфейс `error` — можно возвращать из любой функции и разворачивать через `errors.As` или package-level [`As`](error.go).

## Collector

[`Collector`](collector.go) — основная поверхность для построения ошибок. Накапливает `FieldError`-ы и возвращает `*Error` (или `nil`) в конце:

| Метод | Описание |
|---|---|
| `Check(ok bool, fe FieldError)` | Добавляет `fe` если `ok == false`; chainable |
| `Add(fes ...FieldError)` | Безусловно добавляет одну или несколько ошибок |
| `Merge(namespace string, src error)` | Поглощает все `FieldError`-ы из вложенного validator-а, добавляя `namespace`-префикс к каждому полю |
| `Err() error` | Возвращает `*Error` как `error`-интерфейс или `nil` |
| `Validation() *Error` | Возвращает `*Error` напрямую для инспекции полей или `nil` |

### Вложенная валидация

`Merge` позволяет легко компоновать validator-ы:

```go
func validateAddress(city, zip string) error {
    c := validation.NewCollector("invalid address")
    c.Check(city != "", validation.Required("city"))
    c.Check(len(zip) == 5, validation.Invalid("zip_code", "must be 5 digits"))
    return c.Err()
}

c := validation.NewCollector("invalid order")
c.Merge("shipping_address", validateAddress("", "123"))
// Fields: "shipping_address.city", "shipping_address.zip_code"
```

Если `src` — не `*Error`, а обычная ошибка, `Merge` оборачивает её как `CodeInvalid` с `namespace` в качестве поля.

### Инспекция ошибок

```go
ve, ok := validation.As(err)    // развернуть из любой error chain
fe, ok := ve.First("email")     // первая ошибка для поля
fes    := ve.FieldsFor("email") // все ошибки для поля
codes  := ve.Codes()            // уникальные code-значения по всем полям
```

## Безопасность относительно nil

Все функции, принимающие `error`-аргументы — `As`, `Is`, `Collector.Merge` — обрабатывают `nil` как "нет ошибки" и возвращают zero/false значения без паники. Это позволяет писать:

```go
c.Merge("address", mayReturnNil())
```

## Concurrency

`*Error` и `Collector` **не защищены от конкурентного использования**. Строй их в одной goroutine и передавай готовый `error` через goroutine-границы.

## Makefile

| Цель | Описание |
|---|---|
| `make code-gen` | Запустить `go generate ./...` |
| `make lint` | Запустить golangci-lint |
| `make test` | Генерация кода + тесты с coverage |
| `make prof` | Собрать профили (cpu, mem, block, mutex) |
| `make prof-view` | Открыть профиль в браузере (`FILE=cpu.out` по умолчанию) |

## Лицензия

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)