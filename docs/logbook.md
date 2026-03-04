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
- Learned how to write an NGINX config
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

## 8th January
- Control plane now links to data plane
- Fixed unit tests

## 9-10th January
### AWS API Gateway
- "At any scale"
- Restful APIs and WebSocket APIs
- Supports containerised and serverless workloads and web applications
- "API Gateway Portals"
- Up to hundred of thousands of concurrent API calls
- Traffic management, CORS support, authorisation, access control, throttling, monitoring, API version management
- supports HTTP APIs, REST APIs and Websocket APIs
- 90c per million requests
- Monitor performance metrics on API calls, data latency, and error rates from a dashboard using Amazon CloudWatch
- Pay-as-you-go
- Can generate SDKs for Android, JavaScript and iOS
- Developer portal
- Rate Limiting/Throttling
- API code must be hooked up to AWS Lambda
- JWT Auth, AWS identity
- https://aws.amazon.com/api-gateway/features/
- https://daniil-sokolov.medium.com/performance-analysis-of-lambda-backed-rest-api-on-the-aws-api-gateway-for-python-go-and-typescript-bc296732e5ae

### Azure API Management
- Routing, security, rate limiting/throttling, caching and observability
- API keys, JWT, certificates
- Packaged as a Linux-based docker container deployed to K8s
- Policies
- Users and user groups
- https://learn.microsoft.com/en-us/azure/api-management/api-management-key-concepts

### Kong Gateway
- Runs into of REST APIs
- Extend with modules and plugins
- Designed to be run on decentralised architectures, such as hybrid-cloud and multi-cloud deployments
- AI Gateway, Authentication, Rate Limiting, Observability, Load balancing
- https://developer.konghq.com/gateway/

### Commonalities
#### Functional requirements
- Request Routing
  - Mostly HTTP, REST
  - AWS and Kong support WebSockets natively
- Authentication
  - All support JWT
  - Amazon and Azure have their own auth systems
  - API key validation
  - Support OAuth2
- Rate Limiting/Throttling
- Observability
  - Logging
  - Monitoring
- Caching - reduces load by storing and serving repeated queries

#### Non-Functional Requirements
- Cost Effective
  - https://aws.amazon.com/api-gateway/pricing/
  - https://azure.microsoft.com/en-us/pricing/details/api-management/
- Low Latency
- Scalability
- Availability
  - https://konghq.com/legal/service-level-agreement
  - https://learn.microsoft.com/en-us/azure/api-management/api-management-key-concepts#developer-portal
  - https://aws.amazon.com/api-gateway/sla/
- Security
- Maintainability
  - https://medium.com/@communication_93652/the-importance-of-separation-of-concerns-in-software-engineering-8d5964ba65d9

### How can my Gateway be different
- Education?
- Be explicit about what I am not building - hyper-scale multi-user AZ global distribution system whatever
- Ideas
  1. Specialise in a narrow domain
  2. Emphasize developer ergonomics
  3. Make non-functional behaviour explicit and inspectable
  4. Focus on one or two NFRs, done well
  5. Make it self-hosted or transparent

## 12th January
- https://codeworks.me/blog/why-javascript-is-used-in-web-development/

## 13th January
- Changed from Dockerfiles to docker-compose
- Decisions: How to implement observability
  - Need to be able to display the logs: Grafana is good for this
  - How to check the logs from the gateway to Grafana?
  - Loki, Grafana, Promtail, Alloy, OTel, Prometheus, InfluxDB, FluentBit, ELK/EFK Stack, Plain logs file

## 14th January
- How is this API gateway different?
  - How can a developer understand what their configuration does before it actually goes live?
  - Make the configuration easy to understand
  - Catch confusing or risky configurations early
  - Show the impact of changes before and after applying them
  - Apply changes safely when they are accepted
  - build an API gateway that explains what your configuration does and what will change before it goes live, instead of just applying it and hoping for the best
  - Solves: understanding what a configuration does
  - correctness and understanding first

## 17th January
- Finished logging, doesn't log NGINX access/error as far as I'm aware
- https://blog.nginx.org/blog/rate-limiting-nginx
- Started rate limiting
- Rate limiting works by keeping track of each IP address that requests a service in binary format
- Allows bursty requests to cope with expected HTTP traffic and buffer the requests
- Modified the NGINX template to accept custom rate limit parameters (need to put an upper bound on these)

## 19th January
- Started implementing basic auth in NGINX

## 27th January
- Completed Basic Auth
  - Used apache2-utils to create a password manager
  - GitHub Secrets
  - Curl to test the routes were working

## 4th February
- Semantics Analysis
  - When user updates their config, show them how this what it will change
  - diff?
  - Warn of anything that may be poorly configured
  - Apply changes safely, rollback?
  - Correctness and understanding first
- Step 1: User needs to be able to edit the config
  - Management plane
- Created a simple management plane that shows the user their config file and allows them to edit it

## 7th February
- Started on the semantics analysis
  - Management plane now contains no logic
    - Only responsibility is accepting the config and passing it to the control plane for analysis
    - Control plane reads the config and validates it with a simple rule so far

## 9th February
- Current Goal: Semantics analysis - think of appropriate rules
- Removed api.go: hosting routes now done in main.go
- Created the skeleton code for doing the semantics analysis
  - Two types: Error checking & Update explanation

## 10th February
- Created rules for semantics analysis

## 11th February
- Sent findings from control plane to management plane

## 15th/16th February
- Finished with the semantics analysis
  - Added checks for new routes
- Frontend allows the user to edit their config
- When the user submits their changes, semantics analysis happens
  - TODO: does not yet allow you to remove routes
- If the user has no errors and is happy with the explanation of their changes they hit "accept"
- The gateway config is then updated and the watcher notices and begins the update NGINX process
- Uses a lot of cross container communicates which could be a security problem
- Semantics analysis done on the control plane
- Management Plane used for displaying the config to the user

## 16th February
- Started making the app scalable
- Ideas
  - Need to add a login to the management plane
  - Need to store information per user instead of a single config
  - Firebase?
  - Store information somewhere else

## 18th February
Checklist for allowing multiple users

### Phase 1 - Introduce Users & Persistence
-[x] Add a Database - 21/02
-[x] Create Tables - 21/02
-[x] Implement Authentication - 23/02

### Phase 2 - Scope Users by Config
-[x] Modify Config Endpoints - 28/02

### Phase 3 - Per-User NGINX Files
-[x] Create User Config Directory 02/03
-[x] Update Config Generator 02/03
-[ ] Atomic Writes File
-[ ] Serialize Reloads

### Phase 4 - Concurrency Protection
-[ ] Add optimistic Locking

## 21st February
Issues
-[x] Config.html is not showing the gateway.yaml 24/02
-[x] 401 unauthorised when trying to validate user 22/02
-[x] STATE 23/02

## 24th February
- Instead of serving the config from a static location, the database should be contacted
- See if any of the routes relating directly to the config are in need of changing?
- What works:
  - Management plane now displays the config by loading it from the postgres database
- What is there still to do?
  -[x] Edit the core functions e.g. updating the config file 28/02
  -[x] Load actual configs from the database, not just "test" 28/02

## 25th February
-[x] In HandleNewConfig, content does not exist 28/02

## 27th February
- https://gateway.example.com/<user-id>/auth → Auth Service
- https://gateway.example.com/<user-id>/content → Content Service
- https://gateway.example.com/<user-id>/analytics → Analytics Service

## 28th February
- Merged phase 2 of scaling

## 2nd March
- Created signup on the management, logic handled in control plane
- Updated the way NGINX configs are loaded, now done as soon as a user signs up
  - Initially uses a default config
- Began to modify the watcher to detect config changes, if change is detected it should notify the data plane

## 3rd March
- Watcher looks at correct directory
- Per-User Config should be working
- Problem: watched directory is not exist