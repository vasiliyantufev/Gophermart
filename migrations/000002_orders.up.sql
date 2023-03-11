CREATE TYPE statuses AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');

CREATE TABLE orders
(
    id             serial primary key,
    user_id        int       not null,
    order_id       int       not null,
    current_status statuses    not null,
    created_at     timestamp not null,
    updated_at     timestamp not null
);