package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
func main(){

}

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, MaxInputDataLen)
	out := make(chan interface{}, MaxInputDataLen)
	for _, job := range jobs {
		wg.Add(1)
		go runJob(wg, job, in, out)
		in = out
		out = make(chan interface{}, MaxInputDataLen)
	}
	wg.Wait()
	close(out)
}

func InterfaceToString(input interface{}) string {
	value, ok := input.(string)
	if !ok {
		value = strconv.Itoa(input.(int))
	}
	return value
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for data := range in {
		value := InterfaceToString(data)
		wg.Add(1)

		go func(data string) {
			defer wg.Done()
			hash1 := make(chan string)
			hash2 := make(chan string)

			go func() {
				hash1 <- DataSignerCrc32(data)
			}()

			go func() {
				mu.Lock()
				md5hash := DataSignerMd5(data)
				mu.Unlock()
				hash2 <- DataSignerCrc32(md5hash)
			}()

			out <- (<-hash1 + "~" + <-hash2)
		}(value)
	}

	wg.Wait()
}


func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for data := range in {
		value := InterfaceToString(data)

		wg.Add(1)
		go func(data string) {
			defer wg.Done()
			innerWg := &sync.WaitGroup{}
			dataHashes := make([]string, 6)
			for th := 0; th < 6; th++ {
				innerWg.Add(1)
				go func(th int) {
					defer innerWg.Done()
					dataHashes[th] = DataSignerCrc32(strconv.Itoa(th) + data)
				}(th)
			}
			innerWg.Wait()
			out <- strings.Join(dataHashes, "")
		}(value)
	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	results := make([]string, 0, 5)
	for data := range in {
		results = append(results, data.(string))
	}
	sort.Strings(results)
	result := strings.Join(results, "_")
	out <- result
}

func runJob(wg *sync.WaitGroup, jobF job, in, out chan interface{}) {
	defer wg.Done()
	defer close(out)
	jobF(in, out)
}

