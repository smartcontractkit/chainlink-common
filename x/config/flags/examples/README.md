# examples

Each directory is a runnable program. `cd` into one and follow its README; every example prints
what its settings resolved to, so the fallback chain (flag > env var > config file > compiled-in
default) shows up in the output.

| example                      | shows |
|------------------------------|-------|
| [simple](simple)             | one config struct on one command; the whole fallback chain, list flags, `config.Duration` |
| [namespaced](namespaced)     | two independent structs on one command, kept apart by `Options.Namespace` |
| [nested](nested)             | nested sections: value vs pointer, why `required` needs no pointer, and how an unset pointer section stays `nil` |

```sh
cd simple && go run .
```
