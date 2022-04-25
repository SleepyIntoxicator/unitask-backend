CREATE TABLE "user"
(
    id                 serial primary key,
    login              varchar not null unique,
    full_name          varchar not null,
    email              varchar not null unique,
    encrypted_password varchar not null,
    created_at         timestamptz
);


create table subject
(
    id            serial primary key,
    name          varchar,
    university_id int references university (id)
);

create table "group"
(
    id                  serial primary key,
    university_id       int REFERENCES university (id),
    start_year          varchar(4),
    course_number       int                       not null,
    group_number        varchar,
    created_at          timestamptz default now() not null,
    specialization_name varchar,
    full_name           varchar
);

create table GroupMember
(
    id            serial unique,
    user_id       int REFERENCES "user" (id),
    group_id      int REFERENCES "group" (id),
    invited_by_id int REFERENCES "user" (id),
    primary key (user_id, group_id)
);

create table TaskStatusType
(
    id   serial primary key,
    name varchar(16)
);

create table TaskStatus
(
    id                  serial primary key,
    task_status_type_id int REFERENCES TaskStatusType (id),
    name                varchar not null,
    description         varchar
);

create table Task
(
    id            serial primary key,
    type_id       int,
    is_task_group boolean,
    is_task_local boolean,

    name          varchar,
    content       varchar,
    start_at      timestamptz default now() not null,
    end_at        timestamptz default now() not null,

    subject_id    int REFERENCES subject (id),
    group_id      int REFERENCES "group" (id),
    user_id       int REFERENCES "user" (id),

    added_by_id   int REFERENCES "user" (id),

    created_at    timestamp,
    updated_at    timestamp,
    updates_count int,
    views         int
);

-- update task SET start_at = default
-- where start_at is null;
-- update task SET end_at = default
-- where end_at is null;

create table TaskTree
(
    task_id      int REFERENCES Task (id),
    next_task_id int REFERENCES Task (id),
    PRIMARY KEY (task_id, next_task_id)
);

create table Subtask
(
    task_id        int REFERENCES Task (id),
    parent_task_id int REFERENCES Task (id),
    PRIMARY KEY (task_id, parent_task_id)
);

create table TaskOnGroup
(
    task_id  int REFERENCES Task (id),
    group_id int REFERENCES "group" (id),
    PRIMARY KEY (task_id, group_id)
);

create table TaskOnUser
(
    task_id int REFERENCES Task (id),
    user_id int REFERENCES "user" (id),
    PRIMARY KEY (task_id, user_id)
);

create table UserTask
(
    task_id          serial primary key,
    user_id          int unique REFERENCES "user" (id),
    task_status_id   int REFERENCES TaskStatus (id),
    report_status_id int REFERENCES TaskStatus (id),
    parent_task      int REFERENCES Task (id),
    notes            varchar
);

--Not used yet
create table Permission
(
    id   serial PRIMARY KEY,
    name varchar
);
--Not used yet
create table Role
(
    id          serial PRIMARY KEY,
    name        varchar,
    description varchar
);
--Not used yet
create table RolePermissions
(
    permission_id serial,
    role_id       int REFERENCES Role (id),
    state_boolean boolean,
    PRIMARY KEY (permission_id, role_id)
);

--Not used yet
create table GroupMemberRoles
(
    group_member_id int REFERENCES GroupMember (id),
    role_id         int REFERENCES Role (id),
    PRIMARY KEY (group_member_id, role_id)
);

create table UserRoles
(
    user_id int REFERENCES "user" (id),
    role_id int REFERENCES Role (id),
    PRIMARY KEY (user_id, role_id)
);

create table UserToken
(
    id                   serial PRIMARY KEY,
    user_id              int REFERENCES "user" (id),
    refresh_token        varchar(1024),
    issue_timestamp      timestamptz,
    start_timestamp      timestamptz,
    expiration_timestamp timestamptz,
    exit_timestamp       timestamptz
);

create table RegisteredApp
(
    id         uuid PRIMARY KEY,
    app_name   varchar unique,
    app_secret varchar
);


create table AppToken
(
    token                varchar PRIMARY KEY,
    app_id               uuid REFERENCES RegisteredApp (id),

    issue_timestamp      timestamptz,
    start_timestamp      timestamptz,
    expiration_timestamp timestamptz,
    exit_timestamp       timestamptz
);

create table GroupInviteHashes
(
    id         serial PRIMARY KEY,
    group_id   int REFERENCES "group" (id),
    inviter_id int REFERENCES "user" (id),
    hash       varchar,
    expires_at timestamptz
);

-- 07.06.2021   --

create table university
(
    id       serial PRIMARY KEY,
    name     varchar unique NOT NULL,
    location varchar        NOT NULL,
    site     varchar,
    added_at timestamptz    not null default now()
);
insert into university (id, name, location, site)
VALUES (1, 'РУТ(МИИТ)', 'Россия, Москва, 2-й Вышеславцев переулок, 17', 'miit.ru');


create type status as enum();
alter type status add value  'one';
alter type status add value  'two';
alter type status add value  'three';
drop type status;
