# Concurrency example
#### Bursty producer and slower but steady consumer

I have a production service that is having to deal with a bursty producer (public facing API endpoint) and a slower but more predictable consumer (writing to database). I wrote this as an exercise in how to implement such a producer and consumer in two separate goroutines.

##### Main goroutines
###### Producer
The producer is simply an infinite loop that pushes numbers onto a buffered channel. It first attempts a non-blocking insert. If that fails, it increments a variable that counts how many times the buffer reached full capacity and then performs a blocking insert. This second scenario will of course affect the response time of the public facing API call and is undesirable.

The idea behind this is that this variable can be exposed as an [expvar](https://pkg.go.dev/expvar) or logged to the service logs so that decisions can be made as to whether and by how much the buffer size (and possibly system resources) should be increased.

The producer is simulated here by sleeping 20 milliseconds between messages and an additional 1 second sleep time with a probability of p=0.1 to simulate burstiness.

###### Consumer

There is nothing fancy about the consumer here. It simply reads messages from the channel on 300 millisecond intervals.

##### Additional Plumbing
**bufferPos -** This variable is incremented each time a message is produced onto the channel and decremented each time a message is consumed from the channel. The [atomic package](https://pkg.go.dev/sync/atomic) is used to increment and decrement this counter.

**prodChan & consChan -** These two unbuffered channels are simply used to synchronise the producer and consumer goroutines with a third goroutine that outputs the current value of the producer, the consumer and also the current buffer position. These are not necessary for the functioning of the producer and consumer but in order to visualise the progress, I had to have some goroutine be notified every time a message is produced or consumed.

**illustrate() -** This function reads the abovementioned *prodChan* and *consChan* and calls the *draw()* function to write the state to screen.

**draw() -** This function draws a number scale from 0 to 100 (buffer size) on the screen and shows the current position of the producer on top of the line, the consumer underneath the line and the number of messages in the buffer below that. The producer and consumer pointers will wrap around when it reaches the end of the line but the buffer position will obviously always be somewhere between 0 and 100.

###### Caveat
The *draw()* function makes use of ```fmt.Print("\033[H\033[2J")``` which is effectively a clearscreen command. This will not work on all terminals and might result in unreadable output, depending on which terminal you're using.
