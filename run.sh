go run bytes.go main.go ring_buffer.go shm_linux.go -w -n 1024 -m 200 &
go run bytes.go main.go ring_buffer.go shm_linux.go -r -n 1024 -m 200 &