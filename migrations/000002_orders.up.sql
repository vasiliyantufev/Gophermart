CREATE TABLE orders
(
    id        serial primary key,
    user_id         int not null,
    order_number    int not null,
    status  string  not null,
    accrual int,
    uploaded_at timestamp  not null,
    created_at timestamp    not null,
    updated_at timestamp    not null
);