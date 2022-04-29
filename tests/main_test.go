package tests

import (
	"backend/internal/apiserver"
	"fmt"
	vegeta "github.com/tsenart/vegeta/lib"
	"syscall"
	"testing"
	"time"
)

const configPath = "configs/apiserver.toml"

type User struct {
}

func TestBenchmarkService(t *testing.T) {
	go func() {
		apiserver.Run(configPath)
	}()
	time.Sleep(1 * time.Second)
	quit := make(chan int, 2)

	timeStart := time.Now()

	go func() {
		rate := vegeta.Rate{Freq: 600, Per: time.Second}
		duration := 20 * time.Second
		targeter := vegeta.NewStaticTargeter( /*vegeta.Target{
				Method: "GET",
				URL:    "http://localhost:8080/api/v1/test/users?limit=10",
			},*/
			vegeta.Target{
				Method: "GET",
				URL:    "http://localhost:8080/api/v1/test/tasks?limit=10",
				//Header: http.Header{,				},
			})
		attaker := vegeta.NewAttacker()

		var metrics vegeta.Metrics
		var errorsCount uint64
		latencies := make([]time.Duration, 100)
		for res := range attaker.Attack(targeter, rate, duration, "Hello") {
			metrics.Add(res)
			latencies = append(latencies, res.Latency)

			if errorsCount == 0 && res.Code != 200 {
				fmt.Printf("query -> \n\tcode: %d\n\tlatency: %v\n\terror: %s\n\tBytesIn: %d\n\tBody:\n%s\n\n==== End body ====\n", res.Code, res.Latency, res.Error, res.BytesIn, res.Body)
				fmt.Printf("time from start: %.2fs, count requests: %d\n\n", time.Now().Sub(timeStart).Seconds(), len(latencies))
				t.Fail()
			}
			if res.Code != 200 {
				errorsCount++
			}
			//fmt.Printf("request code: %d\n", res.Code)
		}
		metrics.Close()

		var sum float64
		var avg float64
		var count int
		var countZero int
		for _, m := range latencies {
			//fmt.Printf("request latency: %v\n", m)
			if m.Milliseconds() != 0 {
				avg += m.Seconds()
				sum += m.Seconds()
				count++
			} else {
				countZero++
			}

		}
		avg /= float64(count)

		fmt.Println("total", metrics.Latencies.Total)
		fmt.Println("mean", metrics.Latencies.Mean)
		for i := range metrics.Errors {
			fmt.Printf("\terrors: %s\n", metrics.Errors[i])
		}

		fmt.Printf("avg: %v ms with %d notnull users (sum t %.0f) zero: %d total: %d errors: %d (%.2f%%)\n\n99th percentile: %s\n\n", avg, count, sum, countZero, metrics.Requests, errorsCount, float64(errorsCount)/float64(metrics.Requests)*100, metrics.Latencies.P99)

		quit <- 0
	}()
	/*
		go func() {
			rate := vegeta.Rate{Freq: 1000, Per: time.Second}
			duration := 2 * time.Second
			targeter := vegeta.NewStaticTargeter(vegeta.Target{
				Method: "GET",
				URL: "http://localhost:8080/api/v1/test/users?limit=10",
			},
				vegeta.Target{
					Method: "GET",
					URL: "http://localhost:8080/api/v1/test/tasks?limit=10",
				})
			attaker := vegeta.NewAttacker()

			var metrics vegeta.Metrics
			latencies := make([]time.Duration, 100)
			for res := range attaker.Attack(targeter, rate, duration, "Hello") {
				metrics.Add(res)
				latencies = append(latencies, res.Latency)
				//fmt.Printf("request code: %d\n", res.Code)
			}
			metrics.Close()

			var sum float64
			var avg float64
			var count int
			for _, m := range latencies {
				//fmt.Printf("request latency: %v\n", m)
				if m.Milliseconds() != 0 {
					avg += m.Seconds()
					sum += m.Seconds()
					count++
				}

			}
			avg /= float64(count)

			fmt.Println("total2", metrics.Latencies.Total)
			fmt.Println("mean2", metrics.Latencies.Mean)
			fmt.Println("errors2", metrics.Errors)

			fmt.Printf("avg2: %v ms with %d users (sum %.0f)\n\n99th percentile: %s\n\n", avg, count, sum, metrics.Latencies.P99)

			quit <- 0
		}()
	*/

	<-quit
	fmt.Printf("\nfirst run stoped\n\n")
	syscall.Signal.Signal(syscall.SIGTERM)
}
