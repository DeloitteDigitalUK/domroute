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

```
$ domroute keep example.com 10.0.0.1

2022/01/25 14:37:54 keeping example.com routed to 10.0.0.1 - checking every 30 seconds
2022/01/25 14:37:54 resolved example.com to [34.117.59.81]
2022/01/25 14:37:54 creating route 34.117.59.81->10.0.0.1
2022/01/25 14:37:54 created route 34.117.59.81->10.0.0.1

# ...repeats every 30 seconds
```

Change the check interval by setting the environment variable (in seconds):

    CHECK_INTERVAL=60

## Delete route table entries for a domain

```
$ domroute delete example.com 10.0.0.1

2022/01/25 14:37:41 resolved example.com to [34.117.59.81]
2022/01/25 14:37:41 deleting route 34.117.59.81->10.0.0.1
2022/01/25 14:37:41 deleted route 34.117.59.81->10.0.0.1
```
