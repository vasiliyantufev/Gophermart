CREATE TABLE users
(
    id         serial       primary key,
    login      varchar(255) not null,
    password   text not null,
    created_at timestamp    not null
);
