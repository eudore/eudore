# signal

signal是进程的命令信号，程序可以通过响应对应进程信号作出相应，例如stop、reload、restart。

go signal定义在os.Signal的一个接口，os/signal实现系统响应接口。


## kill

linux的kill、systemctl命令都可以发出信号。

kill 默认是15退出程序，常用 kill -9 就是强制退出，kill -l可以查询系统全部信号。

可以使用kill -0 测试pid是否存在，信号0不会对程序有影响。

`kill -0 11148 && echo "succes" || echo "faild"`

systemctl通常来start、stop、restart、reload、statue使用，stop、reload、restart分别对应的信号是15、10、12，只要自己程序响应这几个信号也可以使用systemctl管理。

## golang signal

### func

go中涉及signal的标准库有三个，os、os/signal、syscall。

os定义Signal接口，os/signal定义对系统Signal的监听，syscall定义系统Signal对象。

```golang
// package os
// Signal define
type Signal interface {
        String() string
        Signal() // to distinguish from other Stringers
}
// Process sned signal
func (p *Process) Signal(sig Signal) error

// package os/signal
// define signal method
func Ignore(sig ...os.Signal)
func Ignored(sig os.Signal) bool
func Notify(c chan<- os.Signal, sig ...os.Signal)
func Reset(sig ...os.Signal)
func Stop(c chan<- os.Signal)

// package syscall
func Kill(pid int, sig Signal) (err error)
...
```
### example

一个[官网例子][1]

```golang
package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)
}
```
signal.Notify函数第一个参数是chan<- os.Signal，用来接收系统收到的信号，第二个参数是...os.Signal，指定监听的信号，如果为空就监听全部。

通常goroutine一个协程for处理一个chan<- os.Signal，对收到的系统信号处理。

```golang
import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
)
var (
	graceSignalChan		chan os.Signal
)

func init() {
	graceSignalChan = make(chan os.Signal)
	go func() {
		for {
			SignalHandle(<-graceSignalChan)
		}
	}()
}

func SignalHandle(sig os.Signal) error {
	fmt.Println("Get signal: ", sig)
	return nil
}

// Relisten Signal is sigs
func SignalListen(sigs ...os.Signal) {
	signal.Stop(graceSignalChan)
	signal.Notify(
		graceSignalChan,		
		sigs...,
	)
}

func main() {
	fmt.Println("pid is ", os.Getpid())
	SignalListen(syscall.SIGUSR1, syscall.SIGUSR2)
	time.Sleep(time.Duration(100)*time.Second)
}
```

init启动了一个独立协程，然后获取signal并处理，SignalListen就先关闭原先监听，然后监听新信号。

## signal list

在POSIX.1-1990标准中定义的信号列表

|  信号  |  值  |  动作  |  说明 | 
| ------------ | ------------ | ------------ | ------------ | 
|  SIGHUP  |  1  |  Term  |  终端控制进程结束(终端连接断开) | 
|  SIGINT  |  2  |  Term  |  用户发送INTR字符(Ctrl+C)触发 | 
|  SIGQUIT  |  3  |  Core  |  用户发送QUIT字符(Ctrl+/)触发 | 
|  SIGILL  |  4  |  Core  |  非法指令(程序错误、试图执行数据段、栈溢出等) | 
|  SIGABRT  |  6  |  Core  |  调用abort函数触发 | 
|  SIGFPE  |  8  |  Core  |  算术运行错误(浮点运算错误、除数为零等) | 
|  SIGKILL  |  9  |  Term  |  无条件结束程序(不能被捕获、阻塞或忽略) | 
|  SIGSEGV  |  11  |  Core  |  无效内存引用(试图访问不属于自己的内存空间、对只读内存空间进行写操作) | 
|  SIGPIPE  |  13  |  Term  |  消息管道损坏(FIFO/Socket通信时，管道未打开而进行写操作) | 
|  SIGALRM  |  14  |  Term  |  时钟定时信号 | 
|  SIGTERM  |  15  |  Term  |  结束程序(可以被捕获、阻塞或忽略) | 
|  SIGUSR1  |  30,10,16  |  Term  |  用户保留 | 
|  SIGUSR2  |  31,12,17  |  Term  |  用户保留 | 
|  SIGCHLD  |  20,17,18  |  Ign  |  子进程结束(由父进程接收) | 
|  SIGCONT  |  19,18,25  |  Cont  |  继续执行已经停止的进程(不能被阻塞) | 
|  SIGSTOP  |  17,19,23  |  Stop  |  停止进程(不能被捕获、阻塞或忽略) | 
|  SIGTSTP  |  18,20,24  |  Stop  |  停止进程(可以被捕获、阻塞或忽略) | 
|  SIGTTIN  |  21,21,26  |  Stop  |  后台程序从终端中读取数据时触发 | 
|  SIGTTOU  |  22,22,27  |  Stop  |  后台程序向终端中写数据时触发 |

在SUSv2和POSIX.1-2001标准中的信号列表:

| 信号  |  值  |  动作  |  说明 |
| ------------ | ------------ | ------------ | ------------ |
|  SIGTRAP  |  5  |  Core  |  Trap指令触发(如断点，在调试器中使用)|
|  SIGBUS  |  0,7,10  |  Core  |  非法地址(内存地址对齐错误) |
|  SIGPOLL  |    |  Term  |  Pollable event (Sys V). Synonym for SIGIO | 
|  SIGPROF  |  27,27,29  |  Term  |  性能时钟信号(包含系统调用时间和进程占用CPU的时间) | 
|  SIGSYS  |  12,31,12  |  Core  |  无效的系统调用(SVr4) | 
|  SIGURG  |  16,23,21  |  Ign  |  有紧急数据到达Socket(4.2BSD) | 
|  SIGVTALRM  |  26,26,28  |  Term  |  虚拟时钟信号(进程占用CPU的时间)(4.2BSD) | 
|  SIGXCPU  |  24,24,30  |  Core  |  超过CPU时间资源限制(4.2BSD) | 
|  SIGXFSZ  |  25,25,31  |  Core  |  超过文件大小资源限制(4.2BSD) |


[1]: https://golang.org/pkg/os/signal/#example_Notify
