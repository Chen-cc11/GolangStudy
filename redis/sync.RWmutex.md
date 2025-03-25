# 2 Sync.Mutex

## 2.1 核心机制
* 从逻辑上，可以把RWMutex理解成一把读锁加一把写锁；  
* 写锁具有严格的排他性，当其被占用，其他试图写锁或者读锁的goroutine均堵塞；  
* 读锁具有有限的共享性，当其被占用，试图取写锁的goroutine会被阻塞，试图取读锁的goroutine可与当前goroutine共享读锁；  
* 综合来说，RWMutex适用于读多写少的场景，最理想化的情况，当所有操作均使用读锁，则可实现无锁化；最悲观的情况，倘若所有操作均使用写锁，则RWMutex退化为普通的Mutex。  

## 2.2 数据结构  
![alt text](image.png)
```go
const rwmutexMaxReader = 1 << 30

type RWMutex struct {
    w              Mutex  // held if there are pending writers
    writerSem      uint32 // semaphore for writers to wait for completing reader
    readerSem      uint32 // semaphore for readers to wait for completing writer
    readerCount    int32  //number of pending readers
    readerWait     int32  // number of departing readers
} 
```
* rwmutexMaxReaders : 共享读锁的goroutine数量上限，值为2^29;  
* w: RWMutex内置的一把普通互斥锁sync.Mutex;  
* writerSem: 关联写锁阻塞队列的信号量；  
* readerSem：关联读锁阻塞队列的信号量；  
* readerCount：正常情况下等于介入读锁流程的goroutine数量； 当goroutine接入写锁流程时，该值为实际介入读锁流程的goroutine数量减rwmutexMaxReaders.  
* readerWait: 记录在当前goroutine获取写锁前，还需要等待多少个goroutine释放读锁。

## 2.3 读锁流程

### 2.3.1 RLock
```go
func (rw *RWMutex) RLock() {
    if atomic.AddInt32(&rw.readerCount, 1) < 0 {
        runtime_SemacqiureMutex(&rw.readerSem, false, 0)
    }
}
```
* 基于原子操作，将RWMutex的readCount变量+1，表示占用或等待goroutine数+1；  
* 倘若RWMutex.readCount的新值仍小于0，说明有goroutine未释放写锁，因此将当前goroutine添加到读锁的阻塞队列中并阻塞挂起。

### 2.3.2 RUnlock

**(1)RUnlock方法主干**
```go
func (rw *RWMutex) RUnlock() {
    if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
        rw.rUnlockSlow(r)
    }
}
```
* 基于原子操作，将RWMutex的readCount变量+1，表示占用或等待读锁的goroutine数-1；  
* 倘若RWMutex.readCount的新值小于0，说明有goroutine在等待获取写锁，则走入RWMutex.rUnlockSlow流程中。  

**（2）rUnlockSlow**
```go
func (rw *RWMutex) rUnlockSlow(r int32) {
    if r+1 == 0 || r+1 == -rwmutexMaxReaders {
        fatal("sync: RUnlock of unlocked RWMutex")
    } 
    if atomic.AddInt32(&rw.readerWait, -1) == 0 {
        runtime_Semrelease(&rw.writerSem, false, 1)
    }
}
```
* 对RWMutex.readerCount进行校验，倘若发现当前协程此前从未抢占过读锁，或者介入读锁流程的goroutine数量达到上限，则抛出fatal;  
(倘若r+1==-rwmutexMaxReaders，说明此时有goroutine介入写锁流程，但此前没加过读锁；倘若r+==0，则此前没加过读锁)。  


## 2.4 写锁流程

### 2.4.1 Lock
```go
func (rw *RWMutex) Lock() {
    rw.w.lock()
    r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
    if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
        runtime_SemacuqireMutex(&rw.writerSem, false, 0)
    } 
}
```
* 对RWMutex内置的互斥锁进行加锁操作；  
* 基于原子操作，对RWMutex.rreaderCount进行减少-rwmutexMaxReaders操作；  
* 倘若此时存在未释放读锁的goroutine，则基于原子操作在RWMutex.readerWait的基础上加上介入读锁流程的goroutine数量，并将当前goroutine添加到写锁的阻塞队列中挂起。  

### 2.4.2 Unlock
```go
func (rw *RWMutex) Unlock() {
    r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
    if r >= rwmutexMaxReaders {
        fatal("sync: Unlock of unlocked RWMutex")
    }
    for i := 0; i < int(r); i++ {
        runtime_Semrelease(&rw.readerSem, false, 0)
    }
    rw.w.Unlock()
}
```
* 基于原子操作，将RWMutex.readerCount的值加上rwmutexMaxReaders;  
* 倘若发现RWMutex.readerCount的新值大于rwmutexMaxReaders，则说明要么当前RWMutex从未上过写锁，要么介入读锁流程的goroutine数量已经超限，因此直接抛出fatal；  
* 因此唤醒读锁阻塞队列中的所有goroutine；（乐见，竞争读锁的goroutine更具有优势）；  
* 解开RWMutex内置的互斥锁。