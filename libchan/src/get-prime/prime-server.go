package main
import "fmt"
import "math"
import "container/heap"
import "sync"
import "net"
import "log"

import "github.com/docker/libchan"
//import "github.com/docker/libchan/spdy"

const nWorker = 3
const nRequester = 3

var wg sync.WaitGroup

type RemoteRequest struct {
    Start     uint64
    End       uint64
    StatusChan libchan.Sender
}

type Response struct {
    Status int
    Result []uint64
}

type Request struct {
  start uint64
  end uint64
  c chan []uint64
}

func requester(work chan Request, X uint64, Y uint64) {
  c := make(chan []uint64)
  defer wg.Done()
  work <- Request{X, Y, c}
  result := <-c
  fmt.Println(len(result))
}

type Worker struct {
  i int
  requests chan Request
  pending int
  address string
}

func (w *Worker) work(done chan *Worker) {
  req := <-w.requests
  req.c <- dial(w, req.start, req.end)
  done <- w
}

func dial(w *Worker, start uint64, end uint64) ([]uint64) {
  //var client net.Conn
  var err error
  endpoint := w.address + ":9323"
  _, err = net.Dial("tcp", endpoint)

  if err != nil {
      log.Fatal(err)
  }
  var u []uint64
  return u
}

type Pool []*Worker

func (p Pool) Len() int { return len(p) }

func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending
}

func (p *Pool) Swap(i, j int) {
	a := *p
	a[i], a[j] = a[j], a[i]
	a[i].i = i
	a[j].i = j
}

func (p *Pool) Push(x interface{}) {
	a := *p
	n := len(a)
	a = a[0 : n+1]
	w := x.(*Worker)
	a[n] = w
	w.i = n
	*p = a
}

func (p *Pool) Pop() interface{} {
	a := *p
	*p = a[0 : len(a)-1]
	w := a[len(a)-1]
	w.i = -1 // for safety
	return w
}

type Balancer struct {
	pool Pool
	done chan *Worker
	i    int
}

func NewBalancer() *Balancer {
	done := make(chan *Worker, nWorker)
	b := &Balancer{make(Pool, 0, nWorker), done, 0}
	for i := 0; i < nWorker; i++ {
		w := &Worker{requests: make(chan Request, nRequester)}
		heap.Push(&b.pool, w)
		go w.work(b.done)
	}
	return b
}

func (b *Balancer) balance(work chan Request, jobs uint64) {
     i := uint64(0)
     for i < jobs {
		select {
		case req := <-work:
			b.dispatch(req)
            i++
		case w := <-b.done:
			b.completed(w)
        }
     }
      //  default:
            //fmt.Println("No message received")
		//b.print()
}

func (b *Balancer) print() {
	sum := 0
	sumsq := 0
	for _, w := range b.pool {
		fmt.Printf("%d ", w.pending)
		sum += w.pending
		sumsq += w.pending * w.pending
	}
	avg := float64(sum) / float64(len(b.pool))
	variance := float64(sumsq)/float64(len(b.pool)) - avg*avg
	fmt.Printf(" %.2f %.2f\n", avg, variance)
}

func (b *Balancer) dispatch(req Request) {
	if false {
		w := b.pool[b.i]
		w.requests <- req
		w.pending++
		b.i++
		if b.i >= len(b.pool) {
			b.i = 0
		}
		return
	}

	w := heap.Pop(&b.pool).(*Worker)
	w.requests <- req
	w.pending++
	//	fmt.Printf("started %p; now %d\n", w, w.pending)
	heap.Push(&b.pool, w)
}

func (b *Balancer) completed(w *Worker) {
	if false {
		w.pending--
		return
	}

	w.pending--
	//	fmt.Printf("finished %p; now %d\n", w, w.pending)
	heap.Remove(&b.pool, w.i)
	heap.Push(&b.pool, w)
}

func main() {
    work := make(chan Request)
    x := uint64(1000000000)
    workers := uint64(3)

    index :=  uint64(1)

    if x > workers {
      index = uint64(math.Ceil(float64(x) / float64(workers)))
    }

    start := uint64(0)
    end := uint64(0)
    for w := uint64(0); w < workers; w++ {
      start = (end + 1)
      if start > x {
        break
      }
      end = (end + (index))
      if start == end {
        end++
      }
      if end > x {
        end = x
      }
      fmt.Println("Worker start:  ", start)
      fmt.Println("Worker end:    ", end)
      fmt.Println("--------")
      wg.Add(1)
      go requester(work, start, end)
    }

    NewBalancer().balance(work, workers)

    wg.Wait()
}
