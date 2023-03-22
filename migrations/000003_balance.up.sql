CREATE TABLE balance
(
    id        serial primary key,
    user_id   int not null,
    order_id  varchar(255),
--     delta     float not null,
    accrue     float not null,
    withdraw   float not null,
    create_at timestamp    not null
);
