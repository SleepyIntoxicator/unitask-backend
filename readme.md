# Unitask-backend ![GO][go-badge]

[go-badge]: https://img.shields.io/github/go-mod/go-version/SleepyIntoxicator/unitask-backend?style=plastic
[db_schema]: ./doc/img/db_schema.jpg
[site_design]: ./doc/img/page_design.png

## Definition

University Task Manager [backend].
Monolithic REST API server for the Unitask service.

Unitask - is the task manager service for studies of universities.

The service allows the administrator (headman or any) of the group to upload tasks received from teachers,
both individual and group, setting deadlines for their delivery. The task manager helps students visualize
the current situation of their individual learning trajectory, as they receive and complete tasks,
which allows students to evenly distribute the workload throughout the semester with increased learning efficiency.
The service is also aimed at organizing centralized, orderly storage of educational materials
throughout the entire period of study, which reduces the time spent searching for them.

At the moment, a service server has been developed without a client application (website).
The potential for the development of the service is great, since it is possible to develop both a website and
mobile applications, with the addition of a variety of functionality to them.

![site_design]

[Site design](https://github.com/SleepyIntoxicator/unitask-backend) in figma.


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
    - services\ - services implementations
  - store\
    - sqlstore\ - repositories
    - teststore\ - test database (hardcoded)
  - migrations
  - pkg
    - auth - auth pkg
    - database\postgres - pkg for working with postgres
    - hooks - pkg with hooks for logrus logger
    - serviceData - pkg for reading the initial data of the services

### Prerequisites
- go 1.16
- Docker & docker-compose

---

The .env file from the dev environment.

---


Database schema:
---
![db_schema]

----
TODO list:
---
N. (1 - 5) - Task priority: from 1 (high) to 5 (low). v - done

1. (2) Put the work with jwt-auth in a separate pkg packages
2. (5) JWT Tokens: Embed FingerPrint for the user. 
   Define the items than define the user's devices. Introduce the concept of a user session.
   1. IP
   2. device (mobile\PC)
   3. date of last use (request)
   4. (5) Add statistics collection of the devices from which requests come.

---
TODO list: done:
---
1. [x] Create a generic config package
2. [x] Create environment config variables
3. [x] (1) Restructure the project (internal part)
4. [x] (2) Throw context through services and repositories

---
Known issues:
---

1. ( ) TokenBlacklistManager: with a large number of users, there is an extremely low propability of a collision
    blocked tokens and new ones that have just been issued.
    Result: the user will immediately receive a error about the expired token, i.e. he will not be able to continue
    working untill he changes the token by re-authorization.
