CREATE TABLE token
(
    id        serial primary key,
    user_id   int not null,
    token      varchar(255) not null,
    create_at timestamp    not null
);