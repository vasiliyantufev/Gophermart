package statuses

type Statuses string

const (
	New        Statuses = "NEW"
	Registered Statuses = "REGISTERED"
	Invalid    Statuses = "INVALID"
	Processing Statuses = "PROCESSING"
	Processed  Statuses = "PROCESSED"
)

/*
-- NEW — заказ загружен в систему, но не попал в обработку;
-- PROCESSING — вознаграждение за заказ рассчитывается;
-- INVALID — система расчёта вознаграждений отказала в расчёте;
-- PROCESSED — данные по заказу проверены и информация о расчёте успешно

-- REGISTERED — заказ зарегистрирован, но не начисление не рассчитано;
-- INVALID — заказ не принят к расчёту, и вознаграждение не будет начислено;
-- PROCESSING — расчёт начисления в процессе;
-- PROCESSED — расчёт начисления окончен;
*/
