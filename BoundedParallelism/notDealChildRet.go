package BoundedParallelism

import (
	"fmt"
	"sync"
	"time"
)

// 描述: 启动n个child goroutine从genChan中读取数据,计算出每个元素的平方值后,继续打印结果
// 优势: 相比buffer chan的两种实现方式, 该方式始终有n 个 child goroutine在运行中,且不会阻塞main goroutine继续执行
func NotDealChildRet() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	wg := sync.WaitGroup{}
	genChan := make(chan int)
	//控制并发度为4,某一时刻只有4个child goroutine在运行中
	limit := 4
	for worker := 0; worker < limit; worker++ {
		// 消费者
		wg.Add(1)
		go func() {
			defer wg.Done()
			for numItem := range genChan {
				if numItem%10 == 0 {
					continue // 如果部分源数据不想处理,用continue 跳过,而不是执行return
				}
				time.Sleep(1 * time.Second)
				fmt.Printf("num:%d squreRet=>%d\n", numItem, numItem*numItem)
			}
		}()
	}
	go func() {
		//关闭genChan,以便让所有goroutine退出
		defer close(genChan)
		// 生产者
		for _, srcItem := range srcNums {
			genChan <- srcItem
		}
	}()

	//main等待child goroutine结束
	wg.Wait()
	fmt.Println("main goroutine exit...")
}

// 缺点: main goroutine会在for循环中卡住很长一段时间
func NotDealChildRetBuffChan01() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	semaphore := make(chan struct{}, 4)
	wg := sync.WaitGroup{}
	for _, srcItem := range srcNums {
		// 消费者
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			if num%10 == 0 {
				return // 如果部分源数据不想处理,用return直接退出child goroutine
			}

			// 通过bufferChan控制并发,某一时刻有len(srcNums)个child goroutine运行中,但是很多goroutine会卡在下面这一步
			// 如果len(srcNums)很大,个人感觉这种方式也不是很好
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()
			//耗时部分
			time.Sleep(1 * time.Second)
			fmt.Printf("num:%d squreRet=>%d\n", num, num*num)
		}(srcItem)
	}
	//main等待child goroutine结束
	wg.Wait()
	fmt.Println("main goroutine exit...")
}

// 缺点: 一次启动len(srcNums)个child goroutine, 而某一时刻只有len(semaphore)个child goroutine在执行耗时操作
// len(srcNums) 很大时,该方式并不合适
func NotDealChildRetBuffChan02() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	semaphore := make(chan struct{}, 4)
	wg := sync.WaitGroup{}
	for _, srcItem := range srcNums {
		// 消费者
		wg.Add(1)
		// 通过bufferChan控制并发,某一时刻运行中的child goroutine数等于len(semaphore),这里是4个
		// main goroutine会在for循环中卡住较长时间,直到只有最后4个元素才能跳出for循环
		semaphore <- struct{}{}
		go func(num int) {
			defer func() {
				wg.Done()
				<-semaphore
			}()

			if num%10 == 0 {
				return // 如果部分源数据不想处理,用return直接退出child goroutine
			}
			//耗时部分
			time.Sleep(1 * time.Second)
			fmt.Printf("num:%d squreRet=>%d\n", num, num*num)
		}(srcItem)
	}
	//main等待child goroutine结束
	wg.Wait()
	fmt.Println("main goroutine exit...")
}
