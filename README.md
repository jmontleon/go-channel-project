INFO:
=====
Just a silly golang prime finding application I cobbled together to help me learn channels and docker/libchan.

Each worker is running the sieve of Eratosthenes. I don't believe it's optimized for splitting up the task like this and realistically each worker is finding primes from 2 to it's ending adddress if I'm not mistaken. There's ways to imrpove this a bit, but I don't care about primes. I just wanted to send data to multiple places, get data back from multiple places, and display it. This was sufficient for me to learn about channels and libchan.

I started by borrowing heavily from Rob Pike's balancer for Go at https://gist.github.com/angeldm/2421216. The main challenge for me was that the program was intended to run with a never ending supply of 'work'. Once that was changed to a discrete set of tasks, my god it's full of deadlocks. So part of the learning was sorting all of that out.

To pass stuff around on the network I used the docker/libchan example as a starting point:
https://github.com/docker/libchan/tree/master/examples

USAGE:
======
Run prime-worker on each host you want to use as a worker and make sure port 65521 is open.  

Run prime-client with an end number and one or more worker hostnames or ip addresses:    
```time go run prime-client.go 10000000000 host1 host2 ...```

10,000,000,000 takes about 4 minutes for me to get the results. I have not tried larger. Smaller is much faster. 1,000,000,000 is under 30 seconds for me.

TODO:
=====
I'm sure a lot of garbage could be cleaned up.
