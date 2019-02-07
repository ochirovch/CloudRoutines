# CloudRoutines
CloudRoutines runs multiple vps and executes a colly application on each one which receive tasks on the server.

Your vps's as goroutines!

## Cloud support
Firstly, it implements DigitalOcean droplets 5 usd/month

Secondly, Google Compute Engine 2,5 usd/month (preemptible)

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
