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

## Advanced Usage

### Alternative Variable Names

A specification struct may be annotated with the `alt:"<name>"` tag
to change the name of the environmental variable `envconfig` accepts.

For example,

```Go
type Specification struct {
    MultiWordVar `alt:"multi_word_var"`
}
```

will now be read in as follows:

```Bash
export MYAPP_MULTI_WORD_VAR="this will be the value"

# export MYAPP_MULTIWORDVAR="and this will not"
```

### Backwards Compatibility

If you were using `envconfig` before this feature, adding the
`alt:"<name>"` tag would break backwards compatibility, because the
original ("smushy" - because the word are smushed together) variable
will be ignored.  To address this, add the `accept_smushy_name:"yes"`
tag.

For example,

```Go
type Specification struct {
    MultiWordVar `alt:"multi_word_var" accept_smushy_name:"yes"`
}
```

will now be read in as follows:

```Bash
# export MYAPP_MULTI_WORD_VAR="if this value is not provided,"

export MYAPP_MULTIWORDVAR="but this one is, this one will get used"
```


### Minor Gotcha

It is worth noting that if *both* variables are supplied, and the struct
member is annotated with *both* tags, **the alternate name will be
used**.

For example:

```Go
type Specification struct {
    MultiWordVar `alt:"multi_word_var" accept_smushy_name:"yes"`
}
```

will now be read in as follows:

```Bash
export MYAPP_MULTIWORDVAR="this value is supplied but ignored"
export MYAPP_MULTI_WORD_VAR="this value takes precedence"
```
