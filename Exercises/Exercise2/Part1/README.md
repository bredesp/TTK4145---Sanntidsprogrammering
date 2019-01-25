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
> Is section of the code where the

### What is the difference between race conditions and data races?
 > *Your answer here*

### List some advantages of using message passing over lock-based synchronization primitives.
> *Your answer here*

### List some advantages of using lock-based synchronization primitives over message passing.
> *Your answer here*
