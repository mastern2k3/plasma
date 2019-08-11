
# Plasma

> Centralized, reactive configuration server built for highly complex and connected data

## Features

* Write your configuration in whatever structure you want
* Consume it with simple, intuitive HTTP requests
* Use Javascript to describe your configuration as you see fit
* Watches configuration folder files, reloading results and propagating changes
* Configuration changes get propagated to subscribers

## TODO

- [ ] Websocket endpoint
- [ ] Context awareness, e.g. Development Production (http://my.plasma/someconf?env=dev)
- [ ] Better js engine, possibly [go-duktape](https://github.com/olebedev/go-duktape)
- [ ] Better babel support (user configurable support, babel 7)
- [ ] Support for more input markups like yaml and toml
- [ ] Support for more output markups like yaml and toml

## Example

`main.js`:
```js
var sec = require("second.js");

module.exports = {
    fromRequire: sec.exampleSecond.plasmaIs,
    ...sec.exampleSecond,
    someString: "weeheww",
    someInt: 1231234,
}
```

`second.js`:
```js
module.exports = {
    exampleSecond: {
        plasmaIs: "awesome"
    }
}
```

### Result

```
$ curl localhost:1323/main

{
    "fromRequire": "awesome",
    "plasmaIs": "awesome",
    "someInt": 1231234,
    "someString": "weeheww"
}
```
