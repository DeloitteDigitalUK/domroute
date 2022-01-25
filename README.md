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

Periodically resolves a domain to one or more IP addresses, then adds each IP to the route table, directing traffic to the provided gateway address.

```
$ sudo domroute keep example.com 10.0.0.1

2022/01/25 14:37:54 keeping example.com routed to 10.0.0.1 - checking every 30 seconds
2022/01/25 14:37:54 resolved example.com to [34.117.59.81]
2022/01/25 14:37:54 creating route 34.117.59.81->10.0.0.1
2022/01/25 14:37:54 created route 34.117.59.81->10.0.0.1

# ...repeats every 30 seconds
```

When the IP address for the domain changes, old IP addresses are removed from the route table and replaced with the latest IP address returned by DNS resolution.

Change the check interval by setting the environment variable (in seconds):

    CHECK_INTERVAL=60

## Delete route table entries for a domain

Resolves a domain to one or more IP addresses, then removes each IP/gateway combination from the route table if it exists.

```
$ sudo domroute delete example.com 10.0.0.1

2022/01/25 14:37:41 resolved example.com to [34.117.59.81]
2022/01/25 14:37:41 deleting route 34.117.59.81->10.0.0.1
2022/01/25 14:37:41 deleted route 34.117.59.81->10.0.0.1
```

## State

A record of entries added to the route table is kept in the state file:

    ~/.domroute

This file is used so that domroute can remove deprecated IPs for the domain from the route table.
