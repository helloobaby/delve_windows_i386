**Go调试器delve支持32位进程调试。**



![image-20260110142536790](images/README/image-20260110142536790.png)

原生的delve是不支持调试32位应用程序的，这个项目给delve扩充了下功能，支持调试windows 32位应用。





**构建**

在delve\cmd\dlv下执行命令powershell命令,

```
$env:GOARCH="386";go build -o dlv_32.exe .
```



![image-20260122233343512](images/README/image-20260122233343512.png)
