# domroute

Keeps the route table up to date for a domain's IP addresses.

## Install

    go install github.com/DeloitteDigitalUK/domroute@main

## Usage:

```
add:
  Route traffic for a given domain via a destination gateway.

  domroute add <DOMAIN> <GATEWAY_IP>
  domroute add <DOMAIN> <INTERFACE_NAME>

delete:
  Delete a domain route added by this tool.

  domroute delete <DOMAIN> <GATEWAY_IP>
  domroute delete <DOMAIN> <INTERFACE_NAME>

keep:
  Periodically update resolution for domain and gateway,
  routing traffic for a given domain via a destination gateway.

  domroute keep <DOMAIN> <GATEWAY_IP>
  domroute keep <DOMAIN> <INTERFACE_NAME>

purge:
  Remove all routes added by this tool.

  domroute purge
```

## Add route for a domain

Resolves a domain to one or more IP addresses, then adds each IP to the route table, directing traffic to the provided gateway address.

```shell
$ sudo domroute add example.com 10.0.0.1

2022/01/25 14:37:54 resolved example.com to [34.117.59.81]
2022/01/25 14:37:54 creating route 34.117.59.81->10.0.0.1
2022/01/25 14:37:54 created route 34.117.59.81->10.0.0.1
```

### Using an interface name instead of a gateway IP

You can specify a network interface name, e.g. `utun3`, instead of an IP address for the gateway:

```shell
$ sudo domroute add example.com utun3

2022/01/25 14:37:54 resolved gateway interface utun3 to 10.0.0.1
2022/01/25 14:37:54 resolved example.com to [34.117.59.81]
2022/01/25 14:37:54 creating route 34.117.59.81->10.0.0.1
2022/01/25 14:37:54 created route 34.117.59.81->10.0.0.1
```

## Keep routes up to date for a domain

Same as above, but periodically repeats resolution and route table check.

```shell
$ sudo domroute keep example.com 10.0.0.1

2022/01/25 14:37:54 keeping example.com routed to 10.0.0.1 - checking every 60 seconds
2022/01/25 14:37:54 resolved example.com to [34.117.59.81]
2022/01/25 14:37:54 creating route 34.117.59.81->10.0.0.1
2022/01/25 14:37:54 created route 34.117.59.81->10.0.0.1

# ...repeats every 60 seconds
```

When the IP address for the domain changes, old IP addresses are removed from the route table and replaced with the latest.

Change the check interval by setting the environment variable (in seconds):

    CHECK_INTERVAL=300

## Delete routes for a domain

Resolves a domain to one or more IP addresses, then removes each IP/gateway combination from the route table if it exists. Any previously added routes held in state are also removed.

```shell
$ sudo domroute delete example.com 10.0.0.1

2022/01/25 14:37:41 resolved example.com to [34.117.59.81]
2022/01/25 14:37:41 deleting route 34.117.59.81->10.0.0.1
2022/01/25 14:37:41 deleted route 34.117.59.81->10.0.0.1
```

## Delete all previously added routes

Removes all routes previously added to the route table by this tool.

```shell
$ sudo domroute purge

2022/01/25 14:37:41 deleting route 34.117.59.81->10.0.0.1
2022/01/25 14:37:41 deleted route 34.117.59.81->10.0.0.1
```

## State

A record of entries added to the route table is kept in the state file:

    ~/.domroute

This file is used so that domroute can remove deprecated IPs for the domain from the route table.
