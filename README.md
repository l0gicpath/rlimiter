# Rate Limiter

A request rate limiter that sits between a consumer and a cluster of web nodes limiting
the request rate across the entire cluster per request PATH.

Please check [Implementation Details](doc/implementation.org) for more information.


## Development

You will need Golang 1.9+, ag, entr(1) and glide, a linux box with
build-essential and GNUmake

### Setup

My current development mahcine is running Ubuntu, I'm no longer using
Mac so I won't be including any macOS setup instructions. Please check
each tool's corresponding link for detailed setup instructions.

1. Glide, for dependency management and vendoring. Used for reproducable builds

```
$ curl https://glide.sh/get | sh
```
You can get more information here: [Glide.sh](https://glide.sh/)

2. [entr(1)](http://entrproject.org/) and [ag](https://github.com/ggreer/the_silver_searcher) for a nice development workflow. I use this to rebuild the
project when a file changes.

```
$ apt-get install entr
$ apt-get install silversearcher-ag
```

_You might want to be running this with sudo_

### Running

`make run` will run the service.

### Building

Just run `make` 
