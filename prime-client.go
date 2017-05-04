package main
import "fmt"
import "math"
import "container/heap"
import "sync"
import "net"
import "log"
import "os"
import "strconv"

import "github.com/docker/libchan"
import "github.com/docker/libchan/spdy"

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
  fmt.Println("Worker start:  ", X)
  fmt.Println("Worker end:    ", Y)
  fmt.Println("Worker primes: ", len(result))
  fmt.Println("--------")
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
  var client net.Conn
  var err error
  endpoint := w.address + ":65521"
  client, err = net.Dial("tcp", endpoint)
  if err != nil {
      log.Fatal(err)
  }

  p, err := spdy.NewSpdyStreamProvider(client, false)
  if err != nil {
    log.Fatal(err)
  }
  transport := spdy.NewTransport(p)
  sender, err := transport.NewSendChannel()
  if err != nil {
    log.Fatal(err)
  }

  receiver, remoteSender := libchan.Pipe()

  request := &RemoteRequest{
    Start:      start,
    End:        end,
    StatusChan: remoteSender,
  }

  err = sender.Send(request)
  if err != nil {
    log.Fatal(err)
  }

  response := &Response{}

  err = receiver.Receive(response)
  if err != nil {
    log.Fatal(err)
  }

  return response.Result
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
	for i := 0; i < len(os.Args[2:]); i++ {
		w := &Worker{requests: make(chan Request, nRequester), address: os.Args[i+2]}
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
    if len(os.Args) < 2 {
        log.Fatal("usage: <end> <worker> [ worker ... ]")
    }

    work := make(chan Request)

    c, _ := strconv.ParseInt((os.Args[1]), 0, 64)

    x := uint64(c)
    workers := uint64(len(os.Args[2:]))

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
      wg.Add(1)
      go requester(work, start, end)
    }

    NewBalancer().balance(work, workers)

    wg.Wait()
}
