package main

import (
	"fmt"
	"os"
	"os/exec"
	"bufio"
	"io"
	flag "github.com/spf13/pflag"
)

func getParam() (int, int, int, bool, string, string){
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
		fmt.Fprint(os.Stderr, "Please choose only one type of text, -l or -f\n")
		os.Exit(0)
	}
	if os.Args[1][:2] != "-s" && os.Args[2][:2] != "-e" {
		fmt.Fprint(os.Stderr, "Please input -snumber and -enumber first\n")
		os.Exit(0);
	}

	if *start <= 0 || *end <= 0 || *start > *end {
		fmt.Fprint(os.Stderr, "Please input valid numbers of start and end page\n")
		os.Exit(0)
	}
	if *lNumber <= 0 {
		fmt.Fprint(os.Stderr, "Please input valid line number\n")
		os.Exit(0)
	}
	

	// fmt.Printf("start: %d\n", *start)
	// fmt.Printf("end:%d\n", *end)
	// fmt.Printf("line number of each page: %d\n", *lNumber)
	// fmt.Printf("whether use \\f as end of page: %t\n", *fd)
	// fmt.Printf("select destination: %s\n", *dst)
	// fmt.Printf("target file: %s\n", file)
	return *start, *end, *lNumber, *fd, *dst, file
}

func selpg(start, end, lNumber int, fd bool, dst, file string) {
	bfRd := bufio.NewReader(os.Stdin)
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprint(os.Stderr, "Failed to open file!\n")
			os.Exit(0)
		}
		defer f.Close()
		bfRd = bufio.NewReader(f);
	}

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

	if fd == false {
		count := 0
		start = (start-1) * lNumber
		end = end * lNumber
		for {
			line, err := bfRd.ReadString('\n')
			if err != nil {
				fmt.Fprint(os.Stderr, "Failed to read line!\n")
				os.Exit(0)
			}
			count += 1

			if count > start {
				fmt.Fprint(output, line)
			}
			if count == end {
				break
			}
		}
	} else {
		count := 0
		start = start - 1
		for {
			page, err := bfRd.ReadString('\f')
			if err != nil {
				fmt.Fprint(os.Stderr, "Failed to read page!\n")
				os.Exit(0)
			}
			count += 1
			if count > start {
				fmt.Fprint(output, page)
			}
			if (count == end) {
				break
			}
		}
	}
	output.Close()

	if dst != "" {
		printer.Stdout = os.Stdout
		printer.Stderr = os.Stderr
		if err := printer.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start printer!\n")
		}
	}
}

func main() {
	start, end, lNum, fd, dst, file := getParam()
	selpg(start, end, lNum, fd, dst, file)

}
