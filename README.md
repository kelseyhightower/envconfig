# envconfig

[![Build Status](https://travis-ci.org/kelseyhightower/envconfig.svg)](https://travis-ci.org/kelseyhightower/envconfig)

```go
import "github.com/kelseyhightower/envconfig"
```

## Documentation

See [godoc](http://godoc.org/github.com/kelseyhightower/envconfig)

## Usage

Set some environment variables:

```bash
export MYAPP_DEBUG=false
export MYAPP_PORT=8080
export MYAPP_USER=Kelsey
export MYAPP_RATE="0.5"
export MYAPP_TIMEOUT="3m"
export MYAPP_USERS="rob,ken,robert"
export MYAPP_COLORCODES="red:1,green:2,blue:3"
export MYAPP_DB_0_HOST="mysql-1.default"
export MYAPP_DB_1_HOST="mysql-2.default"
```

Write some code:

```go
import (
	"fmt"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Specification struct {
	Debug      bool
	Port       int
	User       string
	Users      []string
	Rate       float32
	Timeout    time.Duration
	ColorCodes map[string]int
	DB         []struct {
		Host string
		Port int `default:"3306"`
	}
}

func main() {
	var s Specification
	err := envconfig.Process("myapp", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	format := "Debug: %v\nPort: %d\nUser: %s\nRate: %f\nTimeout: %s\n"
	_, err = fmt.Printf(format, s.Debug, s.Port, s.User, s.Rate, s.Timeout)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("Users:")
	for _, u := range s.Users {
		fmt.Printf("  %s\n", u)
	}

	fmt.Println("Color codes:")
	for k, v := range s.ColorCodes {
		fmt.Printf("  %s: %d\n", k, v)
	}

	fmt.Println("Databases:")
	for i, db := range s.DB {
		fmt.Printf("  %d: %s:%d\n", i, db.Host, db.Port)
	}
}
```

Results:

```bash
Debug: false
Port: 8080
User: Kelsey
Rate: 0.500000
Timeout: 3m0s
Users:
  rob
  ken
  robert
Color codes:
  green: 2
  blue: 3
  red: 1
Databases:
  0: mysql-1.default:3306
  1: mysql-2.default:3306
```

## Struct Tag Support

Envconfig supports the use of struct tags to specify alternate, default, and required
environment variables.

For example, consider the following struct:

```go
type Specification struct {
	ManualOverride1         string `envconfig:"manual_override_1"`
	DefaultVar              string `default:"foobar"`
	RequiredVar             string `required:"true"`
	IgnoredVar              string `ignored:"true"`
	AutoSplitVar            string `split_words:"true"`
	RequiredAndAutoSplitVar string `required:"true" split_words:"true"`
}
```

Envconfig has automatic support for CamelCased struct elements when the
`split_words:"true"` tag is supplied. Without this tag, `AutoSplitVar` above
would look for an environment variable called `MYAPP_AUTOSPLITVAR`. With the
setting applied it will look for `MYAPP_AUTO_SPLIT_VAR`. Note that numbers
will get globbed into the previous word. If the setting does not do the
right thing, you may use a manual override.

Envconfig will process value for `ManualOverride1` by populating it with the
value for `MYAPP_MANUAL_OVERRIDE_1`. Without this struct tag, it would have
instead looked up `MYAPP_MANUALOVERRIDE1`. With the `split_words:"true"` tag
it would have looked up `MYAPP_MANUAL_OVERRIDE1`.

```bash
export MYAPP_MANUAL_OVERRIDE_1="this will be the value"

# export MYAPP_MANUALOVERRIDE1="and this will not"
```

If envconfig can't find an environment variable value for `MYAPP_DEFAULTVAR`,
it will populate it with "foobar" as a default value.

If envconfig can't find an environment variable value for `MYAPP_REQUIREDVAR`,
it will return an error when asked to process the struct.  If
`MYAPP_REQUIREDVAR` is present but empty, envconfig will not return an error.

If envconfig can't find an environment variable in the form `PREFIX_MYVAR`, and there
is a struct tag defined, it will try to populate your variable with an environment
variable that directly matches the envconfig tag in your struct definition:

```bash
export SERVICE_HOST=127.0.0.1
export MYAPP_DEBUG=true
```

```go
type Specification struct {
	ServiceHost string `envconfig:"SERVICE_HOST"`
	Debug       bool
}
```

Envconfig won't process a field with the "ignored" tag set to "true", even if a corresponding
environment variable is set.

## Supported Struct Field Types

envconfig supports these struct field types:

  * string
  * int8, int16, int32, int64
  * bool
  * float32, float64
  * structs
  * slices of any supported type, including structs
  * maps (keys and values of any supported type), except structs
  * [encoding.TextUnmarshaler](https://golang.org/pkg/encoding/#TextUnmarshaler)
  * [encoding.BinaryUnmarshaler](https://golang.org/pkg/encoding/#BinaryUnmarshaler)
  * [time.Duration](https://golang.org/pkg/time/#Duration)

Embedded structs using these fields are also supported.

## Slices of structs

Envconfig supports slices of structs in the following form:

```bash
export MYAPP_INNER_0_VALUE=hello
export MYAPP_INNER_1_VALUE=world
```

```go
type Specification struct {
	Inner []struct {
		Value string
	}
}
```

The slice itself can be tagged with `envconfig` tag and works in the usual way, 
however, `envconfig` tag will be used only to overwrite the key name and won't be 
used to fill the value from a matching non-prefixed global variable (because all the 
slice elements share the same tag)

For example:

```bash
export MYAPP_ALTERNATIVE_0_KEY=hello
export MYAPP_ALTERNATIVE_1_KEY=world
```

```go
type Specification struct {
	Inner []struct {
		Value string `envconfig:"key"`
	} `envconfig:"alternative"`
}
```

and:

```bash
export ALTERNATIVE_0_KEY=hello
export ALTERNATIVE_1_KEY=world
```

```go
type Specification struct {
	Inner []struct {
		Value string `envconfig:"key"`
	} `envconfig:"alternative"`
}
```

But in the following example, KEY will be ignored, although it's an unprefixed version of an `envconfig` tag:
```bash
export KEY=THIS_WILL_BE_IGNORED
```

```go
type IgnoredTagSpecification struct {
	Inner []struct {
		Value string `envconfig:"key"`
	}
}
```


## Custom Decoders

Any field whose type (or pointer-to-type) implements `envconfig.Decoder` can
control its own deserialization:

```bash
export DNS_SERVER=8.8.8.8
```

```go
type IPDecoder net.IP

func (ipd *IPDecoder) Decode(value string) error {
	*ipd = IPDecoder(net.ParseIP(value))
	return nil
}

type DNSConfig struct {
	Address IPDecoder `envconfig:"DNS_SERVER"`
}
```

Also, envconfig will use a `Set(string) error` method like from the
[flag.Value](https://godoc.org/flag#Value) interface if implemented.
