# Reasons for concurrency and parallelism


To complete this exercise you will have to use git. Create one or several commits that adds answers to the following questions and push it to your groups repository to complete the task.

When answering the questions, remember to use all the resources at your disposal. Asking the internet isn't a form of "cheating", it's a way of learning.

 ### What is concurrency? What is parallelism? What's the difference?
 > Concurrency - When two or more tasks share resource and the computations are executed during overlapping time periods
 > Parallelism - the computations are happening in parallell
 > ![Concurrency vs Parallelism](images/2019/01/concurrency-vs-parallelism.png)

 ### Why have machines become increasingly multicore in the past decade?
 > In the begining of the 21st century the advancement in single core processor technology hit a roadblock due to the ILP-wall, frequency-wall and the Power-wall. The solution to these challenges was to increase the number of cores in the processors, giving a better performance then a singel core processor in the same are. Each core in the multicore processor can run at a lower frequency, giving it a smaller power consumption than the single cored with the same surface area.

 ### What kinds of problems motivates the need for concurrent execution?
 (Or phrased differently: What problems do concurrency help in solving?)
 > Problems where there is a need to run processes simultaneously with shared resources. With task that are independent they can be run in such a manner that it apers as if the run in simultaneously utilizing the resource of the system in the best possible way.

 ### Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
 (Come back to this after you have worked on part 4 of this exercise)
 > There are both positive and negative sides to using concurrent programs.
 > Concurrent programs are great for gives a higher level of modularity and abstraction - great for a programs manageability.
 > A more segnificant negative side of concurrent programming is the complexity of debugging.
> Deadlock – two or more processes are unable to proceed because each is waiting for one of the others to do something
> Race condition – multiple threads or processes read and write a shared data item and the final result depends on the relative timing of their execution
> Starvation – a runnable process is overlooked indefinitely by the scheduler

 ### What are the differences between processes, threads, green threads, and coroutines?
 > Processes: OS-manged. Has its one adress space
 > Threads: OS-manged. Same adress space as the parent
 > Green threads: Same concept as threads, but not OS-managed
 > Coroutines: Not managed by OS. Function that has mulitple entry/exit points, so that it can run more smothly.

 ### Which one of these do `pthread_create()` (C/POSIX), `threading.Thread()` (Python), `go` (Go) create?
 > pthread_create() - Creates a thread within a process
 > threading.Thread() - This module constructs higher-level threading interfaces on top of the lower level _thread module*
 > go - Goroutines are functions or methods that run concurrently with other functions or methods. Goroutines can be thought of as light weight threads managed by the Go runtime (et bibliotek)

 ### How does pythons Global Interpreter Lock (GIL) influence the way a python Thread behaves?
 > Interpreter is a computer program that directly executes instructions written in a programming/scripting language. GIL ensures that interpreter is held by a single thread at any particular instant of time. A python program with multiple threads works in a single interpreter. Only the thread which is holding the interpreter is running at any instant of time.
 > The issue. The machine could be having multiple cores/processors but since the interpreter is held by a single thread, other threads are not doing anything even though they have access to a core. The GIL makes sure that only one of your 'threads' can execute at any one time. A thread acquires the GIL, does a little work, then passes the GIL onto the next thread. As a result, multiple threads can’t effectively make use of  multiple cores.*
 > The difference between a compiler and an interpreter: the compiler translates the source code to machine language. The interpreter interprets the machine language to executions.

 ### With this in mind: What is the workaround for the GIL (Hint: it's another module)?
 > GIL is only a problem when dealing with problems with CPU-bounds.
 > You can use the multiprocess module to workaround the GIL. This is compatible with the threading module, but data is not shared among processes.

 ### What does `func GOMAXPROCS(n int) int` change?
 >  Sets the maximum number of CPUs that can be executing simultaneously and returns the previous setting, by setting n.
