# qsite

`qsite` provides a very minimal server that renders markdown as HTML and performs some basic
convention-over-configuration logic for things like handling 404s. This piece of software was
written to scratch my _very particular_ itch, where I occasionally want to spin up a small website
for a directory of markdown documents with a single command, sometimes for things like D&D campaign
notes or my personal blog. As such, do be warned that this is a rather inflexible piece of software:

- Markdown is parsed at boot time and _does not_ get treated as compiled output
- The server special-cases a couple of paths on the server

I may add static output in the future so that the content can be served from a CDN, but for the time
being, this software is exactly as complex as I need it to be and no more. In particular, handling
the content statically means that everything is encoded into the mux up-front and I don't need to
worry about opening the path-traversal can of worms.

# Usage

## Site Structure

The site structure is pretty minimal:

- `/base.html.tmpl`: This is the HTML page template that will be used. The `.Content` property
  contains the output of the current page, and `.DocumentPath` contains the cleaned-up relative path
  of the document if you'd like to perform anything fancy (see `qsite.TemplateInput`)
- `/pages/`: This is a directory of markdown documents. **This will be served as /**
    - `index.md`: This is where `/` will redirect
    - `404.md` This is the 404 page. You MUST provide this document, or you may see some cryptic
      infinite redirects.
- `/static/`: This is a directory containing all static assets (images, CSS, JavaScript) for the
  site. This is all optional, but obviously necessary if you'd like to have any amount of style
  customization. **This will be served under /static/**.

The `example/` directory in this repo contains a basic example of all of these pieces.

## Server

`qsite` provides a CLI to run the server

```
Usage of qsite:
  -addr string
    	server listen address (default "localhost:8000")
  -root string
    	content root (default ".")
  -static-ttl int
    	static content TTL (default 600)
```

Alternatively, you can import `pkg/qsite` into an existing go program (see
`qsite.BootstrapDefault`).

## Library Documentation

See https://pkg.go.dev/github.com/16bitt/qsite/pkg/qsite
