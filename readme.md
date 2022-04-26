# Unitask [Backend] ![GO][go-badge]

[go-badge]: https://img.shields.io/github/go-mod/go-version/p12s/furniture-store?style=plastic

## Definition

University Task Manager [backend].
Monolithic REST API server for the Unitask service.

## Architecture

- cmd\apiserver - main package
- configs -  working configs
- doc
- internal\
  - api - subject models 
  - apiserver - request handlers
    - api.go - API entries
  - config - server config
  - service\
    - services - services implementations
  - store\
    - sqlstore - repositories
    - teststore - test database (hardcoded)
  - migrations
  - pkg
    - auth - auth pkg
    - hooks - pkg with hooks for logrus logger
    - serviceData - pkg for reading the initial data of the services

### Prerequisites
- go 1.17
- Docker & docker-compose

----
TODO list:
---
N. (1 - 5) - Task priority: from 1 (high) to 5 (low). v - done

1. (1) Restructure the project (internal part)
2. (2) Put the work with jwt-auth in a separate pkg packages
3. (2) Throw context through services and repositories
4. (5) JWT Tokens: Embed FingerPrint for the user. 
   Define the items than define the user's devices. Introduce the concept of a user session.
   1. IP
   2. device (mobile\PC)
   3. date of last use (request)
   4. (5) Add statistics collection of the devices from which requests come.

---
TODO list: done:
---
1. (v) Create a generic config package
2. (v) Create environment config variables

---
Known issues:
---

1. ( ) TokenBlacklistManager: with a large number of users, there is an extremely low propability of a collision
    blocked tokens and new ones that have just been issued.
    Result: the user will immediately receive a error about the expired token, i.e. he will not be able to continue
    working untill he changes the token by re-authorization.