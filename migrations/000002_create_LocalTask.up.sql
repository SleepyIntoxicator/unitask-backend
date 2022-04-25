create table Task
(
    id             serial primary key,
    type_id        int,
    is_task_group  boolean,
    is_task_local  boolean,

    name           varchar,
    content        varchar,

    subject_id     int REFERENCES subject (id),
    group_id       int REFERENCES "group" (id),
    user_id        int REFERENCES users (id),
    parent_task_id int REFERENCES Task (id),
    next_task_id   int REFERENCES Task (id),
    added_by_id    int REFERENCES users (id),

    created_at     timestamp,
    updated_at     timestamp,
    updates_count  int,
    views          int
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


create table UserTask
(
    id             serial primary key,
    user_id        int unique REFERENCES "user" (id),
    task_status_id int REFERENCES TaskStatus (id),
    parent_task    int REFERENCES Task (id),
    notes          varchar
);