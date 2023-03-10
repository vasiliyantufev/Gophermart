CREATE TABLE orders
(
    id           serial    primary key,
    user_id      int       not null,
    order_number int       not null,
    status       varchar   not null,
    accrual      int,
    created_at   timestamp not null,
    uploaded_at  timestamp not null,
);