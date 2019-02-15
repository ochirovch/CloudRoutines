# cloudroutines
cloudroutines runs multiple vps and executes your application on each one.

Your vps's as goroutines!

![Scheme](https://raw.githubusercontent.com/ochirovch/CloudRoutines/master/img/scheme.png)


## Cloud support
Firstly, it implements DigitalOcean droplets 5 usd/month

Secondly, Google Compute Engine ~5 usd/month (preemptible)

## Quick launch
rename _settings.json to settings.json and paste your DigitalOcean token

go run cmd/server/main.go

visit http:/localhost:8099/

## TODO
- [X] add bootstrap

- [X] implements adding vps from dashboard

- [ ] add payload handler

- [ ] add projects

- [ ] add processing results handler
