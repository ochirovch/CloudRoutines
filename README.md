# cloudroutines
cloudroutines launch multiple vps with one click and execute your parsing application on each one. 

Client application get tasks from your server, download pages, parse and send results to server.

Your vps's works like goroutines.

## What problem does it solve?

Many sites limit parsing by ip address, using proxy servers to solve this problem. This application allows you to offer an alternative to using proxy servers.

|               | Proxy servers | CloudRoutines|
| ------------- |  :---:   | :---:   |
| offer ip address  | V         | V            |
| can perform calculations | X  | V            |
| can buy by hour | X           | V            |
| can be the proxy server itself| X  | V       |
| can be private (not shared)| XV    | V       |
| how many ip for 5 usd? | much less but more time    | 720 per hour |

![Scheme](https://raw.githubusercontent.com/ochirovch/CloudRoutines/master/img/scheme.png)


## Cloud support
Firstly, it implements DigitalOcean droplets 5 usd/month

Secondly, Google Compute Engine ~5 usd/month (preemptible)

## Quick launch
rename _settings.json to settings.json and paste your DigitalOcean token

cd $GOPATH/src/github.com/ochirovch/cloudroutines/cmd/server

go run cmd/server/main.go

visit http:/localhost:8099/

## TODO
- [X] add bootstrap

- [X] implements adding vps from dashboard

- [ ] add payload handler

- [ ] add projects

- [ ] add processing results handler
