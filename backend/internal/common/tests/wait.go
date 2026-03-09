package tests

import (
	"net"
	"time"
)

// WaitForPort is a helper function that watis until the given address/port is open.
// If the timeout is reached without a sucessfull connection it return false.
// Source: https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/blob/0e3e9d80eb14639bc42935795f7ca3b73da36304/internal/common/tests/wait.go
// Source license: MIT
// Copyright (c) 2021 Three Dots Labs
func WaitForPort(address string) bool {
	waitChan := make(chan struct{})

	go func() {
		for {
			conn, err := net.DialTimeout("tcp", address, time.Second)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			if conn != nil {
				waitChan <- struct{}{}
				return
			}
		}
	}()

	timeout := time.After(5 * time.Second)
	select {
	case <-waitChan:
		return true
	case <-timeout:
		return false
	}
}
