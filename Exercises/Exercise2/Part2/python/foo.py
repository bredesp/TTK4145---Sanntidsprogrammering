from threading import Thread
from threading import Lock

i = 0
mtx = Lock()

def incrementingFunction():
    global i
    for j in range (0, 1000000):
        mtx.acquire()
        i += 1
        mtx.release()

def decrementingFunction():
    global i
    for k in range (0, 1000000):
        mtx.acquire()
        i -= 1
        mtx.release()

def main():
    global i

    incrementing = Thread(target = incrementingFunction, args = (),)
    decrementing = Thread(target = decrementingFunction, args = (),)
    
    incrementing.start()
    decrementing.start()
    
    incrementing.join()
    decrementing.join()
    
    print("The magic number is %d" % (i))

main()