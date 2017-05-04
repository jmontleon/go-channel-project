package main

import "log"
import "net"

import "github.com/docker/libchan"
import "github.com/docker/libchan/spdy"

func sieveOfEratosthenes(X uint64, Y uint64) (primes []uint64) {
  b := make([]bool, Y+1)
  for i := uint64(2); i <= Y; i++ {
  if b[i] == true { continue }
    if i >= X {
      primes = append(primes, i)
    }
    for k := uint64(i * i); k <= Y; k += i {
      b[k] = true
    }
  }
  return
}

type Request struct {
  Start      uint64
  End        uint64
  StatusChan libchan.Sender
}

type Response struct {
    Status int
    Result []uint64
}

func main() {
  var listener net.Listener
  var err error
  listener, err = net.Listen("tcp", "localhost:9323")
  if err != nil {
    log.Fatal(err)
  }

  for {
    c, err := listener.Accept()
    if err != nil {
      log.Print(err)
      break
    }
    p, err := spdy.NewSpdyStreamProvider(c, true)
    if err != nil {
      log.Print(err)
      break
    }
    t := spdy.NewTransport(p)

    go func() {
      for {
        receiver, err := t.WaitReceiveChannel()
        if err != nil {
          log.Print(err)
          break
        }

        go func() {
          for {
            request := &Request{}
            err := receiver.Receive(request)
            if err != nil {
              log.Print(err)
              break
            }

            returnResult := &Response{}
            returnResult.Result = sieveOfEratosthenes(request.Start, request.End)
            err = request.StatusChan.Send(returnResult)
            if err != nil {
              log.Print(err)
            }
          }
        }()
      }
    }()
  }
}
