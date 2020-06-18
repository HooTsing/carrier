package src

func Recover() {
	err := recover()
	if err != nil {
		LogStack("goroutine failed, err: %v", err)
	}
}
