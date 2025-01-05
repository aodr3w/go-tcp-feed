package server

func Start(writePort int, readPort int) {
	go func() {
		s := NewService()
		WriteMessages(writePort, s)
	}()

	go func() {
		s := NewService()
		ReadMessages(readPort, s)
	}()
}
