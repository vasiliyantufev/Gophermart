CREATE TABLE balance
(
    id        serial primary key,
    user_id   int not null,
    order_id  varchar(255),
    accrue     float not null,
    withdraw   float not null,
    created_at timestamp    not null
);
