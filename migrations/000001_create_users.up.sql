CREATE TABLE users (
                       id serial primary key,
                       login varchar not null unique,
                       full_name varchar not null,
                       email varchar not null unique,
                       encrypted_password varchar not null
);