#include <pthread.h>
#include <stdio.h>

int i = 0;
pthread_mutex_t mtx;

// Note the return type: void*
void* incrementingThreadFunction(){
    for (int j = 0; j < 1000000; j++) {
       pthread_mutex_lock(&mtx);
       i++;
       pthread_mutex_unlock(&mtx);
    }
    return NULL;
}

void* decrementingThreadFunction(){
    for (int j = 0; j < 1000000; j++) {
       pthread_mutex_lock(&mtx);
       i--;
       pthread_mutex_unlock(&mtx);
    }
    return NULL;
}


int main(){
    pthread_t incrementingThread, decrementingThread;

    // 2nd arg is a pthread_mutexattr_t
    pthread_mutex_init(&mtx, NULL);

    pthread_create(&incrementingThread, NULL, incrementingThreadFunction, NULL);
    pthread_create(&decrementingThread, NULL, decrementingThreadFunction, NULL);

    pthread_join(incrementingThread, NULL);
    pthread_join(decrementingThread, NULL);

    pthread_mutex_destroy(&mtx);

    printf("The magic number is: %d\n", i);
    return 0;
}
