CREATE TYPE status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');

CREATE TABLE orders
(
    id           serial    primary key,
    user_id      int       not null,
    order_id     int       not null,
    current_status status,
    created_at   timestamp not null,
    uploaded_at  timestamp not null
);

-- NEW — заказ загружен в систему, но не попал в обработку;
-- PROCESSING — вознаграждение за заказ рассчитывается;
-- INVALID — система расчёта вознаграждений отказала в расчёте;
-- PROCESSED — данные по заказу проверены и информация о расчёте успешно

-- REGISTERED — заказ зарегистрирован, но не начисление не рассчитано;
-- INVALID — заказ не принят к расчёту, и вознаграждение не будет начислено;
-- PROCESSING — расчёт начисления в процессе;
-- PROCESSED — расчёт начисления окончен;