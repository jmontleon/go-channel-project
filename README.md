INFO:
=====
Just a silly golang prime finding application. Not terribly optimized, but it finds primes between a range of numbers by splitting up the work between multiple go routines.

Borrowed heavily from Rob Pike's balancer for Go. The main challenge for me was that the program was intended to run with a never ending supply of 'work'. Once that was changed to a discrete set of tasks, my god it's full of deadlocks. Now that that's solved I'll work on networking.

TODO:
=====
Right now 'workers' are just goroutines on the local host.

I'd like to get to sending work to workers running on this and a couple other hosts.
