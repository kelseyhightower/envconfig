# envconfig

[![Build Status](https://travis-ci.org/kelseyhightower/envconfig.png)](https://travis-ci.org/kelseyhightower/envconfig)

```Go
import "github.com/kelseyhightower/envconfig"
```

## Documentation

See [godoc](http://godoc.org/github.com/kelseyhightower/envconfig)

## Usage

Set some environment variables:

```Bash
export MYAPP_DEBUG=false
export MYAPP_PORT=8080
export MYAPP_USER=Kelsey
export MYAPP_RATE="0.5"
```

Write some code:

```Go
package main

import (
    "fmt"
    "log"

    "github.com/kelseyhightower/envconfig"
)

type Specification struct {
    Debug bool
    Port  int
    User  string
    Rate  float32
}

func main() {
    var s Specification
    err := envconfig.Process("myapp", &s)
    if err != nil {
        log.Fatal(err.Error())
    }
    format := "Debug: %v\nPort: %d\nUser: %s\nRate: %f\n"
    _, err = fmt.Printf(format, s.Debug, s.Port, s.User, s.Rate)
    if err != nil {
        log.Fatal(err.Error())
    }
}
```

Results:

```Bash
Debug: false
Port: 8080
User: Kelsey
Rate: 0.500000
```

Register and use custom decoders:

```Bash
export MYAPP_KEYS=key1,key2,key3
```

```Go
package main

import (
    "fmt"
    "log"
    "reflect"
    "strings"

    "github.com/kelseyhightower/envconfig"
)

type Specification struct {
    Keys []string
}

func main() {
    var s Specification
    envconfig.RegisterDecoder("Keys", func(value string, fieldValue, struc reflect.Value) error {
        items := strings.Split(value, ",")
        n := len(items)

        if fieldValue.Len() < n {
            fieldValue.Set(reflect.MakeSlice(fieldValue.Type(), n, n))
        }

        for i, v := range items {
            fieldValue.Index(i).SetString(v)
        }

        return nil
    })
    defer envconfig.ClearDecoders()
    err := envconfig.Process("myapp", &s)
    if err != nil {
        log.Fatal(err.Error())
    }
    format := "Keys: %v"
    _, err = fmt.Printf(format, s.Keys)
    if err != nil {
        log.Fatal(err.Error())
    }
}
```

```Bash
Keys: [key1  key2  key3]
```
