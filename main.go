package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"unsafe"
	"encoding/binary"
	"bytes"
)

const (
	Void = 0
	Writer = 1
	Reader = 2
)

/*
var sem_name string = "/shm_benchmark_sem"
var sem Semaphore

func init_sem() {
	if err := sem.Remove(sem_name); err != nil {
		//fmt.Println("remove error: " + err.Error())
	}
}

func open_sem() {
	if err := sem.Open(sem_name, 0, 0); err != nil {
		fmt.Println("open error: " + err.Error())
	} else {
		fmt.Println("successfully opened semaphore " + sem.name)
	}
}

func unlink_sem() {
	if err := sem.Unlink(); err != nil {
		fmt.Println("unlink error: " + err.Error())
	} else {
		fmt.Println("successfully unlinked semaphore " + sem.name)
	}
}

func close_sem() {
	if err := sem.Close(); err != nil {
		fmt.Println("close error: " + err.Error())
	} else {
		fmt.Println("successfully closed semaphore " + sem.name)
	}
}
*/

func read_ring_buffer(s *shmi) *RingBuffer {
	return (*RingBuffer)(s.v)
}

func init_ring_buffer(s *shmi, index int) {
	rb := read_ring_buffer(s)
	rb.Init(index)
}

func write_handler(n int, size int) {
	rb := new(RingBuffer)

	w, _ := shm_create("shm_benchmark", int32(unsafe.Sizeof(*rb)))
	defer w.shm_close()

	init_ring_buffer(w, 0)
	rb = read_ring_buffer(w)

	for !rb.IsReady() {}

	fmt.Println("Started writing data...")

	tickStart := time.Now().UnixNano()

	for i := 0; i < n; i ++ {
		s1 := IntToByteArray(int64(i + 1))
		s2 := IntToByteArray(time.Now().UnixNano())

		buf := new(bytes.Buffer)

		if err := binary.Write(buf, binary.BigEndian, s1); err != nil {
			panic("writer: " + err.Error())
		}

		if err := binary.Write(buf, binary.BigEndian, s2); err != nil {
			panic("writer: " + err.Error())
		}

		if size < len(s1) + len(s2) {
			panic("writer: too small size of message")
		}

		prb := make([]byte, size - len(s1) - len(s2))

		if err := binary.Write(buf, binary.BigEndian, prb); err != nil {
			panic("writer: " + err.Error())
		}

		rb.Write(buf.Bytes())
	}

	fmt.Println("writer: elapsed", (time.Now().UnixNano() - tickStart), "nanoseconds")
}

func read_handler(n int, size int) {
	rb := new(RingBuffer)

	var r *shmi
	var err error

	for {
		r, err = shm_open("shm_benchmark", int32(unsafe.Sizeof(*rb)))
		if err == nil {
			break
		}
	}

	defer r.shm_close()

	init_ring_buffer(r, 1)
	rb = read_ring_buffer(r)

	for !rb.IsReady() {}

	fmt.Println("Started reading data...")

	tickStart := time.Now().UnixNano()

	var oldSeq int64 = 0

	var sum, max, min int64 = 0, 0, int64((^uint64(0)) >> 1)

	for i := 0; i < n; i ++ {
		for !rb.IsAvailable() {}

		prb, err := rb.Read()
		if err != nil {
			panic("reader: " + err.Error())
		}

		n1 := ByteArrayToInt(prb[:8])
		n2 := ByteArrayToInt(prb[8:16])

		n2 = time.Now().UnixNano() - n2

		if oldSeq + 1 != n1 {
			panic("reader: sequence number skipped " + strconv.FormatInt(oldSeq, 10) + " => " + strconv.FormatInt(n1, 10))
		}

		oldSeq = n1

		sum += n2
		if max < n2 {
			max = n2
		}

		if min > n2 {
			min = n2
		}
	}

	fmt.Println("reader: elapsed", (time.Now().UnixNano() - tickStart), "nanoseconds")
	fmt.Printf("reader: span %d(ns), max %d(ns), min %d(ns), avg %d(ns)\n", sum, max, min, sum / int64(n))
}

func main() {
	fmt.Println("Share memory benchmark started")

	argsWithProg := os.Args
	context := Void
	count := 0
	msg_size := 50

	for i := 1; i < len(argsWithProg); i ++ {
		if argsWithProg[i] == "-w" {
			context = Writer
		} else if argsWithProg[i] == "-r" {
			context = Reader
		} else if argsWithProg[i] == "-n" {
			if i + 1 < len(argsWithProg) {
				count, _ = strconv.Atoi(argsWithProg[i + 1])
				i ++
			}
		} else if argsWithProg[i] == "-m" {
			if i + 1 < len(argsWithProg) {
				msg_size, _ = strconv.Atoi(argsWithProg[i + 1])
				i ++
			}
		}
	}

	switch context {
	case Writer:
		fmt.Printf("running writer(%d messages in %d bytes)...\n", count, msg_size)
		// init_sem()
		// open_sem()
		write_handler(count, msg_size)
		// close_sem()
		// unlink_sem()
	case Reader:
		fmt.Printf("running reader(%d messages in %d bytes)...\n", count, msg_size)
		// open_sem()
		read_handler(count, msg_size)
		// close_sem()
	default:
		fmt.Println("not writer/reader, exiting")
	}

	// rbuf == wbuf
}
