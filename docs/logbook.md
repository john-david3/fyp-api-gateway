# Logbook
## 4th October
- Created a one-page project description document
- Set up a GitHub repository for the project and cloned
- Used Mermaid to create a System Architecture diagram of how I see the system being used

## 6th October
- Make a skeleton file structure
- Created sample NGINX config for api gateway

## 22nd October
- Had a look at microservices libraries - found some interesting ones
    - Go-kit: build from stratch, not ideal for this project
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
- User should have the gateway installed, and will only need a config file and the set of commands to interact with the gateway
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
    - Reload the NGINX config is there are changes
    - Register new microservices
    - Enforce policies (Auth, rate limit, etc)
    - Maybe:
      - Send logs to loki?
      - Interactive CLI?
- User Workflow (* = "not done by user")
  - Download the api gateway
  - Create a config file (register the microservices)
  - *Config generation
  - *NGINX restarts with new config (logging)
  - Update frontend API URLs
  - *Gateway should route correctly
  - Can check logs on Grafana, may need to run another command to display grafana dashboard

### v2 Cloud Implementation Additions
- Multi-tenancy: Use Lua scripting to dynamically assign microservices to different NGINX instances
  - Each tenant has their own hosted config file
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
  - `nginx -t` to validate the config
  - Github Actions for CI
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
- Ease of Use - user should only have to change a config, gateway takes care of everything else
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
- Scalability - Online platform could dynamically create new NGINX instances when the config file gets to a certain size (or based on region)

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
- Express is mostly for NodeJS coding, while javascript is good, it's not what I'm looking for in terms of concurrency
- Kong is built on top of NGINX
- NGINX has an API Management Module, that performs better than Kong
- Kong comes preloaded with a lot of the features I want to use
- I will be limited to Kong's requirements if I use it, with NGINX, I could get better performance if I configure correctly https://www.f5.com/company/blog/nginx/nginx-controller-api-management-module-vs-kong-performance-comparison

Requirement: Requests should be processed quickly

## 23rd December

