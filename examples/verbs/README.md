# JavaScript verb examples

This directory contains user-visible JavaScript verb repositories for
`goja-site verbs`. The embedded built-in verbs are always available; these local
copies are useful when editing or learning the verb format.

Run the local copies with an explicit repository:

```bash
go run ./cmd/goja-site verbs \
  --repository examples/verbs/builtin \
  examples local-builtin hello --name Manuel
```

List all available verbs, including embedded and local repositories:

```bash
go run ./cmd/goja-site verbs \
  --repository examples/verbs/builtin \
  list
```
