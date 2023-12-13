package main

func main() {
	s := NewServer(":8089")
	panic(s.Start())
}
