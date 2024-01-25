# Field

[![Go Reference](https://pkg.go.dev/badge/github.com/go-haru/field.svg)](https://pkg.go.dev/github.com/go-haru/field)
[![License](https://img.shields.io/github/license/go-haru/field)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/go-haru/field.svg?style=flat-square)](https://github.com/go-haru/field/releases)
[![Go Test](https://github.com/go-haru/field/actions/workflows/go.yml/badge.svg)](https://github.com/go-haru/field/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-haru/field)](https://goreportcard.com/report/github.com/go-haru/field)

provides key-value list for instant json generation, which drives data collecting feature in [`log`](https://github.com/go-haru/log) and [`errors`](https://github.com/go-haru/errors).

## Usage

When observing or debug with complicated transaction with changeable data, we always need on-site data for metric or analysing purpose.

With this repo, we can describe named data as `field` and append id to log or error

### field

consists of a string name and arbitrary type content, it can be encoded as json bytes or write to a buffer:

```go
type Field struct {
	Key string // name
	Content    // wrapped value
}

func (f Field) EncodeJSON(buffer Buffer) error { ... }

func (f Field) MarshalJSON() ([]byte, error) { ... }
```

We made different functions for common data types, so you need care noting about wrapping value

```go
func Bool(key string, val bool) Field

func Bools(key string, valArr []bool) Field

func Error(key string, err error) Field

func Float64(key string, val float64) Field

// ... refer to go doc for more impl
```
Now, we can add field to logger or error (refer to their repo for more info):

```go
// data for logger
logger.With(field.String("traceId", id)).Debug("data created")

// data for error
errors.Note(err, field.String("traceId", id))

```

### fields

is list of `field`, which provides simple map-like operation and can be encoded to json object 

```go
type Fields []Field

func (f Fields) Unique() []Field 
func (f Fields) Has(key string) bool
func (f Fields) Get(key string) (Field, bool) 
func (f Fields) Export() map[string]any 
func (f Fields) EncodeJSON(buf Buffer) (err error) 
func (f Fields) MarshalJSON() (dst []byte, err error) 
```

## Testing

All types of field are supposed to be finely tested in [field_test.go](./field_test.go). you ca run test with command:

```shell
go test ./...
```
[Github Action](https://github.com/go-haru/field/actions) is also enabled for main branch, every release shall pass all tests.

## Contributing

For convenience of PM, please commit all issue to [Document Repo](https://github.com/go-haru/go-haru/issues).

## License

This project is licensed under the `Apache License Version 2.0`.

Use and contributions signify your agreement to honor the terms of this [LICENSE](./LICENSE).

Commercial support or licensing is conditionally available through organization email.
