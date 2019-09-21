# CLI命令行程序selpg
本程序实现了[开发Linux命令行实用程序](https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html)中的selpg，selpg的思想十分简单，就是从不同的输入文件中，根据输入的参数，读取第s页到第e页的内容，并输出到标准输出中。下面主要讲一下实现的细节：

### 参数解析
本程序使用`pflag`库来对参数进行解析，参数的定义如下：
```go
var start = flag.IntP("start", "s", -1, "Start page")
var end = flag.IntP("end",  "e", -1, "End page")
var lNumber = flag.IntP("lNumber", "l", 72, "line number of each page")
var fd = flag.BoolP("f", "f",false,"whether use \\f as end of page")
var dst = flag.StringP("dDestination", "d", "", "select a destination")

var file string = ""
flag.Parse()
if (len(flag.Args()) == 1) {
    file = flag.Args()[0]
}
```

selpg规定`-s`和`-e`参数必须是前两个参数，并且`-l`和`-f`是互斥的，所以要对参数增加约束：
- 判断参数中是否同时存在`-l`和`-f`
    ```go
    is_f, is_l := false, false
    for i:=1; i < len(os.Args); i+=1 {
        if os.Args[i][1] == 'l' {
            is_l = true
        }
        if os.Args[i][1] == 'f' {
            is_f = true
        }
    }
    if is_f && is_l {
        io.WriteString(os.Stderr, "Please choose only one type of text, -l or -f\n")
        os.Exit(0)
    }
    ```
    参数`-l`和`-f`表示读取何种类型的文件，这两个参数互斥。
- 判断`-s`和`-e`的位置是否正确
    ```go
    if os.Args[1][:2] != "-s" && os.Args[2][:2] != "-e" {
		io.WriteString(os.Stderr, "Please input -snumber and -enumber first\n")
		os.Exit(0);
	}
    ```
    `-s`和`-e`必须分别是输入的第一个和第二个参数。
- 判断输入的参数是否有意义
    ```go
    if *start <= 0 || *end <= 0 || *start > *end {
        io.WriteString(os.Stderr, "Please input valid numbers of start and end page\n")
        os.Exit(0)
    }
    if *lNumber <= 0 {
        io.WriteString(os.Stderr, "Please input valid line number\n")
        os.Exit(0)
    }
    ```
    起始页数和结束页数不能为0，起始页不能大于结束页。每一页的行数不能小于0。

### 文件读取
使用`bufio`包进行文件的读取，对于打开的文件。`start-1`\*`lNumber`为文件开始读取的前一行的行数，假如当前文件行数大于该行数，即可开始读取。`end`\*`number`是文件读取的最后一行的行数，假如当前行数大于此行数，则停止读取。

### 打印机调用
假如输入的参数中包含`-d`，要通过管道将输出流输送到打印命令的输入流。
```go
var output io.WriteCloser = os.Stdout
var printer *exec.Cmd
if dst != "" {
    printer = exec.Command("lp", "-d", dst)
    printoutput, err := printer.StdinPipe()
    if err != nil {
        fmt.Fprint(os.Stderr, "Failed to get printer!\n")
        os.Exit(0)
    }
    output = printoutput
}
```
通过在终端输入`$man lp`，查看`lp`命令的用法，发现`lp`命令使用参数`-d`指定打印机。通过`exec.Command()`函数指定要运行的系统命令，并且通过`StdinPipe()`函数将程序的输出作为`lp`命令的输入。
```go
if dst != "" {
    printer.Stdout = os.Stdout
    printer.Stderr = os.Stderr
    if err := printer.Start(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to start printer!\n")
    }
}
```
将`lp`命令的输出设置标准输出，然后通过`Start()`函数执行该命令。

### 测试案例

作为每页固定行数的文件，这里选择`littleprince.txt`作为输入样例，该文件为小说《小王子》；作为每页以`\f`结尾的文件，这里选择`testf.txt`作为输入样例，该文件是我自己编写的测试文件，只有六页，每页一句话。
1. `$ go run selpg.go -s1 -e1 littleprince.txt`
> $...
“啊？这真滑稽”

输出从1到72行，测试通过。

2. `$ go run selpg.go -s1 -e1 < littleprince.txt`
> $...
“啊？这真滑稽”

该命令将输入重定向为`littleprince.txt`，输出同`1.`，测试通过。

3. `$ cat littleprince.txt | go run selpg.go -s1 -e1`
> $...
“啊？这真滑稽”

该命令将`cat littleprince.txt `的输出通过管道，作为`selpg`的输入，输出和`1.2.`相同，测试通过。

4. `$ go run selpg.go -s6 -e10 littleprince.txt > output_file`

执行完命令后，可见`output_file`中共有360行数据，测试通过。

5. `$ go run selpg.go -s1 -e30 littleprince.txt >output_file 2>error_file`

这里读取的行数超过了文本的行数，所以会报错。打开`error_file`，可见
> Failed to read line!

测试通过。

6. `$ go run selpg.go -s1 -e1 littleprince.txt | cat`

这里将`selpg`的输出作为`cat`的输入，输出结果跟`1.`相同，测试通过。

7. `$ go run selpg.go -s1 -e5 -l1 littleprince.txt`

> $ 小王子　
\
（法）圣埃克苏佩里——关于生命和生活的寓言
\
\
‘这就像花一样。如果你爱上了一朵生长
\
　在一颗星星上的花，那么夜间，

这里将每页的行数改为1，1到5页只输出了5行，测试通过。

8. `$ go run selpg.go -s1 -e3 -f testf.txt`

> page1
     page2
          page3
读取使用`\f`换页的文件`testf.txt`，测试通过
