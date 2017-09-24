# gohan/config

The gohan/config package gets your application config values from different sources
allowing you to run your app in different environments. 

Config properties are considered in the following order:

   1. Env variables (Upper case format, e.g. FIRST_NAME)
   2. json file (Standard camel case syntax, e.g. firstName)

Each item takes precedence over the item below it.

The json file is loaded from CONFIG_JSON_FILE env var or from the `config.json` file.

## Usage

```
package main

import (
	"fmt"
	"github.com/appnaconda/gohan/config"
	"log"
)

func main() {
    // Supported tags: default, ignored, required and alias
	c := struct {
		Debug bool
		Db    struct {
			Host       string `default: "localhost"`
			Username   string `default:"root" alias:"db_user"`
			Password   string `required:"true"`
			Port       uint   `default:"3306"`
			AutoCommit bool `ignored:"true"`
		}
	}{}

	if err := config.Load(&c); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("config: %#v", c)
}
```