CREATE TABLE token
(
    id         serial primary key,
    user_id    int not null references users (id) on delete cascade,
    token      varchar(255) not null,
    created_at timestamp    not null,
    deleted_at timestamp    not null
);
