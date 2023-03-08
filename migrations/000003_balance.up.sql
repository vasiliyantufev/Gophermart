CREATE TABLE balance
(
    id        serial primary key,
    user_id   int not null,
    debit     float,
    credit    float,
    create_at timestamp    not null
);