# Logbook
## 4th October
- Created a one-page project description document
- Set up a GitHub repository for the project and cloned
- Used Mermaid to create a System Architecture diagram of how I see the system being used

## 6th October
- Make a skeleton file structure
- Created sample NGINX nginxConfig for api gateway

## 22nd October
- Had a look at microservices libraries - found some interesting ones
    - Go-kit: build from scratch, not ideal for this project
    - Go-micro: might be what I'm looking for
- Had a look at NodeRed
- Took at look at existing API Gateways
  - Key Features: 
    - handle many types of requests (HTTP, REST, WebSockets)
    - rate limiting
    - caching
    - Auth (OAuth, keys)
  - Current Limitations:
    - AWS: configs are complex, can't handle large payloads
- NGINX will be data plane, use go for control plane

## 27th October
- Need to decide between hosting a single API gateway where users can register microservices;
  - or a runnable instance of an API gateway

### Runnable API Gateway
- Initially going to focus on getting this to run locally, if time, make it cloud platform (v2)
- User should have the gateway installed, and will only need a nginxConfig file and the set of commands to interact with the gateway
- Features
  - NGINX data plane
    - Request Routing
    - Handle policies specified
    - Emit logs and metrics
    - Support HTTP at first, with possible HTTPS, gRPC and REST support
  - Go Control Plane
    - Parse configs
    - Validate configs are correct
    - Generate NGINX configs
    - Reload the NGINX nginxConfig is there are changes
    - Register new microservices
    - Enforce policies (Auth, rate limit, etc)
    - Maybe:
      - Send logs to loki?
      - Interactive CLI?
- User Workflow (* = "not done by user")
  - Download the api gateway
  - Create a nginxConfig file (register the microservices)
  - *GatewayConfig generation
  - *NGINX restarts with new nginxConfig (logging)
  - Update frontend API URLs
  - *Gateway should route correctly
  - Can check logs on Grafana, may need to run another command to display grafana dashboard

### v2 Cloud Implementation Additions
- Multi-tenancy: Use Lua scripting to dynamically assign microservices to different NGINX instances
  - Each tenant has their own hosted nginxConfig file
  - Each tenant must have their own API URL
- Web management dashboard
  - Logs could be displayed here
- Could put the gateway behind a load balancer?
- API Keys

### Tools
- Data plane
  - NGINX
  - Lua to route dynamically
  - Loki/Prometheus and Grafana for logging
- Control plane
  - Go to create methods
    - Will need to be able to parse configs in YAML (gopkg.in/yaml.v3 maybe)
    - Cobra for CLI stuff
  - Need to run shell command to parse NGINX, possibly done in Go, but likely Bash
  - Make - to run simple automated commands
- Running
  - Postman to test the APIs
  - `nginx -t` to validate the nginxConfig
  - GitHub Actions for CI
  - Maybe docker for APIs?
- Cloud
  - AWS/EC2

## 5th November
### Requirements
#### Must Haves
- Request Routing
- Authentication - key-based
- Logging
- Rate Limiting
- Configuration Management
- Ease of Use - user should only have to change a nginxConfig, gateway takes care of everything else
- High Availability
  - Health checks
  - Graceful retries - if a backend connection, fails to n times, x seconds apart to reconnect, don't completely kill the program over one failure
- Good Performance
  - High throughput - 1000s of connections per second
  - Low Latency
  - NGINX should be able to handle both issues

#### Nice to Haves
- Authentication - Basic Auth or JWT
- Encrypted transmission
- Scalability - Online platform could dynamically create new NGINX instances when the nginxConfig file gets to a certain size (or based on region)

## 12th November
- Remade architecture

## 19th November
### Month 1
- Create Control Plane
- Start on the hosted Gateway with docker

### Month 2
- Finish Auth
- Start Rate Limiting
- Finish Logging

### Month 3
- Health Checks
- Scale on Cloud
- Start the dashboard

### Month 4
- Finish dashboard
- Add tests

## 26th November
Requirement: Gateway has to be able to handle high load
- Rust, Go and C++ are the best at high-performance concurrency https://medium.com/@lpramithamj/the-race-to-1m-tasks-35018c35e347
- Express is mostly for Node.js coding, while javascript is good, it's not what I'm looking for in terms of concurrency
- Kong is built on top of NGINX
- NGINX has an API Management Module, that performs better than Kong
- Kong comes preloaded with a lot of the features I want to use
- I will be limited to Kong's requirements if I use it, with NGINX, I could get better performance if I configure correctly https://www.f5.com/company/blog/nginx/nginx-controller-api-management-module-vs-kong-performance-comparison

Requirement: Requests should be processed quickly

## 23rd December
- Learned how to test apis in Golang https://speedscale.com/blog/testing-golang-with-httptest/
- Initially decided to use Go version 1.25.5 (latest release), decided to downgrade to Go version 1.24.11, as it the most stable version of Go that works with GitHub actions
- Created some basic microservices using net/http
- Created some unit tests for the microservices to make sure there were working as expected
- Makefile to automate running the net/http microservices and the unit tests
- Updated the architecture in Mermaid

## 24th December
- Started on creating the data plane with NGINX (why NGINX?)
- Created a simple dataplane
- Changed the microservices to run on different ports (defeats the purpose of using NGINX if all same port)
- Learned how to write an NGINX nginxConfig
- Containerised the Microservices and the Dataplane in docker
- Added docker make targets
- Challenges: Setting up dockerfiles, NGINX nginxConfig

## 27th December
- Created a YAML nginxConfig file (Why YAML?)
- Created a test directory to store files needed for unit/integration tests
- Started on the control plane
  - Created a method to register the nginxConfig file
  - Needed to make an interface to load the nginxConfig file into
  - Validated the location and syntax of the nginxConfig file
- Unit tests for loading the nginxConfig file

## 28th December
- Tried to make a function for updating the NGINX config based on the users gateway config
  - Had to come up with a new approach
  - Have decided to make a separate NGINX config for each user and "include" it in the main NGINX config
    - Trade-off: not very scalable for a lot of users, dynamic approach would take too long, may be able to come back to it, but this is the compromise

## 30th December
- Discovered that NGINX Plus exposes APIs that allow you to dynamically edit the NGINX config
  - Not feasible as it costs money to use
- Alternative solution: Tried using a template nginx config
  - Trade-off: need to remake the config from scratch every time something changes
  - Best compromise I could find that didn't cost money
  - Checked if there was a config parser for NGINX, one exists in Python, which would have make the architecture too complicated
- Started writing a small test for the new method

## 31st December
- Finished unit test for updating the NGINX config

## 1st January
- Create a file watcher for the gateway config, when event happens, it triggers the update NGINX function
- Currently, this works for a single user, I must implement database to make it multi-user compatible

## 3rd January
- Did something, kinda forgot what tho, probably really cool though

## 4th January
- Containerised the control plane

## 5th January
- Figuring out how to send the updated NGINX config from the control plane to the data plane
- Created a ConfigStore object which will allow me to serve an API
- Will have the dataplane poll?? the API in search of a newer NGINX config

## 7th January
- Fix problem where data plane could not connect to the control plane container
- Created new file for hosting the API for the data plane to get the updated config
- Created a polling system in the data plane
- Need to figure out why NGINX is not reloading correctly with `nginx -s reload`