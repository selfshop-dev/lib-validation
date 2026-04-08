# lib-validation

[![CI](https://github.com/selfshop-dev/lib-validation/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-validation/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-validation/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-validation)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-validation)](https://goreportcard.com/report/github.com/selfshop-dev/lib-validation)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-validation)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-validation)](LICENSE)

Структурированные ошибки валидации с машиночитаемыми кодами для Go. Без внешних зависимостей. Проект организации [selfshop-dev](https://github.com/selfshop-dev).

### Installation

```bash
go get -u github.com/selfshop-dev/lib-validation
```

## Overview

`lib-validation` даёт единый тип `Error`, который несёт все ошибки полей сразу — API-обработчик может сериализовать все проблемы в одном ответе, а клиент переключается на стабильные значения `Code`, не разбирая строки сообщений.

```go
func ValidateUser(req CreateUserRequest) error {
    c := validation.NewCollector("invalid user")
    c.Check(req.Name != "", validation.Required("name"))
    c.Check(len(req.Name) <= 100, validation.TooLong("name", 100))
    c.Check(isEmail(req.Email), validation.Invalid("email", "must be a valid address"))
    c.Merge("address", validateAddress(req.Address))
    return c.Err()
}

// На принимающей стороне:
if ve, ok := validation.As(err); ok {
    for _, fe := range ve.Fields {
        fmt.Printf("%s: [%s] %s\n", fe.Field, fe.Code, fe.Message)
    }
}
```

Ключевые свойства:

- **Нет внешних зависимостей** — только стандартная библиотека Go.
- **Стабильные коды** — `Code` — это часть публичного API-контракта; клиенты могут зависеть от их строковых значений.
- **Вложенная валидация** — `Merge` автоматически расставляет dot-notation префиксы полей.
- **Безопасность по умолчанию** — значение поля никогда не попадает в ошибку автоматически; нужен явный вызов `WithValue`.

### Быстрый старт

```go
import "github.com/selfshop-dev/lib-validation"

c := validation.NewCollector("invalid user")
c.Check(req.Name != "", validation.Required("name"))
c.Check(len(req.Name) <= 50, validation.TooLong("name", 50))

if err := c.Err(); err != nil {
    return err
}
```

## Коды ошибок

Коды — стабильная часть публичного API. Переименование или удаление кода — это breaking change, требующий мажорного релиза. Добавление новых кодов — безопасно.

| Код | Константа | Описание |
|---|---|---|
| `required` | `CodeRequired` | Поле отсутствует или пустое |
| `invalid` | `CodeInvalid` | Значение присутствует, но некорректно |
| `too_long` | `CodeTooLong` | Строка или срез превышает максимальную длину |
| `too_short` | `CodeTooShort` | Строка или срез короче минимальной длины |
| `out_of_range` | `CodeOutOfRange` | Числовое значение вне допустимого диапазона |
| `conflict` | `CodeConflict` | Значение конфликтует с существующим состоянием |
| `immutable` | `CodeImmutable` | Поле нельзя изменить после создания |
| `type_mismatch` | `CodeTypeMismatch` | Неверный тип значения |
| `unknown` | `CodeUnknown` | Нераспознанный ключ (используется в lib-config) |

## Билдеры

Готовые конструкторы покрывают наиболее частые сценарии и автоматически заполняют `Code`, `Message` и `Meta`.

```go
validation.Required("email")                       // поле обязательно
validation.Invalid("email", "not a valid address") // некорректное значение
validation.TooLong("username", 50)                 // Meta: {"max": 50}
validation.TooShort("password", 8)                 // Meta: {"min": 8}
validation.OutOfRange("age", 18, 120)              // Meta: {"min": 18, "max": 120}
validation.Conflict("email", "already taken")
validation.Immutable("user_id")
validation.TypeMismatch("count", "integer") // Meta: {"expected_type": "integer"}
validation.Unknown("extra_field")

validation.Entity(validation.CodeConflict, "duplicate entry") // ошибка уровня сущности, без поля
```

## Collector

`Collector` накапливает ошибки в ходе прохода по полям и возвращает `*Error` (или `nil`) в конце. Все методы возвращают `*Collector` и поддерживают цепочки вызовов.

```go
c := validation.NewCollector("invalid user")

// Добавить ошибку если условие ложно
c.Check(req.Name != "", validation.Required("name"))

// Добавить ошибку если условие истинно — удобно когда условие описывает нарушение
c.Fail(len(req.Name) < minLen, validation.TooShort("name", minLen))
c.Fail(len(req.Name) > maxLen, validation.TooLong("name", maxLen))

// Добавить безусловно
c.Add(validation.Required("email"))

// Вложенная валидация с автоматическим префиксом
c.Merge("shipping_address", validateAddress(req.Address))
// поле "city" внутри → "shipping_address.city"

// Получить результат
err := c.Err()        // error или nil
ve  := c.Validation() // *Error или nil — для инспекции полей
```

## Инспекция ошибок

После получения `*Error` доступны несколько методов для поиска по полям.

```go
ve, ok := validation.As(err) // достать *Error из цепочки ошибок
ve.Fields                    // все FieldError

ve.First("email")                                     // (FieldError, bool) — первая ошибка поля
ve.FirstWithCode("password", validation.CodeTooShort) // по полю и коду

ve.FieldsFor("email") // все ошибки поля
ve.Codes()            // уникальные коды по всем полям

validation.Is(err) // есть ли *Error в цепочке (без инспекции полей)
```

## FieldError

`FieldError` описывает одну ошибку валидации. Поле `Field` использует dot-notation и совместимо как с конфигурационными ключами (`database.host`), так и с путями JSON-тела (`user.address.zip_code`). Пустой `Field` означает ошибку уровня сущности.

```go
fe := validation.Invalid("status", "unrecognised value")

// Прикрепить безопасное значение для отладки — только для не-чувствительных полей
fe = fe.WithValue("PENDING_APPROVAL")

// Добавить метаданные
fe = fe.WithMetaPair("allowed", []string{"active", "inactive"})
```

`WithValue` и `WithMetaPair` возвращают копию — оригинальный `FieldError` не мутируется.

## Вложенная валидация

`Merge` позволяет вызывать отдельные функции валидации для вложенных структур и собирать их ошибки в единый результат с правильными путями полей.

```go
c := validation.NewCollector("invalid order")
c.Merge("shipping_address", validateAddress(req.ShippingAddress))
c.Merge("billing_address", validateAddress(req.BillingAddress))

// Итоговые пути: "shipping_address.city", "billing_address.zip_code" и т.д.
```

Глубина вложенности не ограничена — каждый уровень просто добавляет свой префикс.

## Лицензия

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)