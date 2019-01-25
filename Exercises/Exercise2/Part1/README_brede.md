# Mutex and Channel basics

### What is an atomic operation?
> Program operations that run completely independently of any other processes can simultaneously read a loactions andr wrtie it in the same bus operation.

### What is a semaphore?
> "Bouncer at a nightclub" - limit the number of consumers for a specific resource. The is a que and as one exit the next can enter (use the resource).

### What is a mutex?
> **Mul**tual **ex**clusion object (mutex) is a for shareing the same recsources, such as fil access, but not simultaneously. The thread that needs the resource must lock the mutex from other threads while it is using the resource. The mutex is set to unlock when the data is no longer needed or the routine is finished.

### What is the difference between a mutex and a binary semaphore?
> The shared recsource for a mutex can only be released by the thread that acquired it while semaphore can be released by any thread.

### What is a critical section?
> In concurrent programming, accessing a shared resource can be problamtic. Therefor are parts of the program protected, called the critical section. it cannot be accessed by more than one process at a time. The critical section can forexample be a data structure or network connection, this will not work properly if it accessed by mulitple concurrent users.

### What is the difference between race conditions and data races?
 > When two threads at the same time: read the same variable, and then write to the same variable with different values, there may occur an error when assigning the values -> Data race.

> Race condition: A data race that causes an error.

### List some advantages of using message passing over lock-based synchronization primitives.
> Message passing does not use shared memory, which means they don't need locks. They communicate through sen & receive messages. Therefore there are lesser concurrency bugs with message passing than locking. Messaging --> safer alternative. Locking uses shared resource.

### List some advantages of using lock-based synchronization primitives over message passing.
> You don't need to take into account an other user, as there is no communication between two users. Locking means locking.
hei
