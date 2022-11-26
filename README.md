# gorpeh (gopher anagram)

This is a basic implementation of a Gopher server in Go/Golang. See my blog post on Gopher for a bit more information on the protocol [here](https://koryporter.com/posts/gopher-the-father-of-the-world-wide-web). The [RFC](https://datatracker.ietf.org/doc/html/rfc1436) is also great sunday reading.

## Consuming

1. Clone this repo
2. Run: `go build && ./gorpeh` and you're good to go! That will serve the sample folder.

### CLI API

`gorpeh` supports: `--port`, `--host`, and `--directory` flags.

Use the `--directory` flag to point the gopher serve anywhere on your machine.

**Example**

```bash
gorpeh \
  --port 1234 \
  --host localhost \
  --directory="/Users/me/my-cool-repo/"
```
