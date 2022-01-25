# domroute

Keep route table up to date for a domain.

## Install

    go install github.com/pcornish/domroute@latest

## Usage:

```
domroute keep <DOMAIN> <GATEWAY>
domroute delete <DOMAIN> <GATEWAY>
```

## Keep route table up to date for a domain

    domroute keep example.com 10.0.0.1

## Delete route table entries for a domain

    domroute delete example.com 10.0.0.1
