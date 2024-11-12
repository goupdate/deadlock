package deadlock

type GoroutineID int64

var lastId int64 = 1

func GetGoroutineId() GoroutineID {
	return GoroutineID(atomic.AddInt64(&lastId, 1))
}
