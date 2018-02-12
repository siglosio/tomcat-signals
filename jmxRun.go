package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//jmxterm <OPTIONS> - see end of this file

// https://tomcat.apache.org/tomcat-7.0-doc/funcspecs/mbean-names.html

// Catalina:type=GlobalRequestProcessor,name=http-8080][requestCount]
// Catalina:type=GlobalRequestProcessor,name=http-8080][errorCount]
// Catalina:type=GlobalRequestProcessor,name=http-8080][processingTime]  valid?
// Catalina:type=ThreadPool,name=http-8080 currentThreadCount
// Catalina:type=ThreadPool,name=http-8080 currentThreadsBusy
// Catalina:type=ThreadPool,name=http-8080 maxThreads
// java.lang:type=Memory][HeapMemoryUsage,used
// java.lang:type=Memory][HeapMemoryUsage,max

// How to find PID, maybe get unique string as argument; error if count() is 0 or >1?

// Run JMX

func runJMX(pid int) ([6]int, error) {
	var (
		vals [6]int
	)

	// JVM info
	jmxBeanRuntime := "bean java.lang:type=Runtime\n"
	jmxCommandRunTime := "get Uptime\n"

	// Request Count
	jmxBeanReq := "bean Catalina:name=\"" + argProcessorName + "\",type=GlobalRequestProcessor\n"
	jmxCommandReq := "get requestCount\n"
	jmxCommandErr := "get errorCount\n"
	jmxCommandTime := "get processingTime\n"

	jmxBeanPool := "bean Catalina:type=ThreadPool,name=\"" + argProcessorName + "\"\n"
	jmxCommandThdCur := "get currentThreadsBusy\n"
	jmxCommandThdMax := "get maxThreads\n"

	// Now write input file with PID in it
	inputFile := "jmx.input"
	fw, err := os.Create(inputFile)
	checkErr(err)
	// JMX tool commands, one per line
	s := "close\n" +
		"open " + fmt.Sprintf("%d\n", pid) +
		jmxBeanRuntime + jmxCommandRunTime +
		jmxBeanReq + jmxCommandReq + jmxCommandErr + jmxCommandTime +
		jmxBeanPool + jmxCommandThdCur + jmxCommandThdMax +
		"close\n" + "exit\n"

	if flagVerbose {
		fmt.Printf("Tomcat Request Processor Argument: %s\n\n", argProcessorName)
		fmt.Println("JMXterm commands --------------")
		fmt.Printf("%s", s)
		fmt.Println("-------------------------------")
	}

	_, err = io.WriteString(fw, s)
	checkErr(err)
	fw.Close()

	// Delete output file so no old stuff lying around
	outputFile := "jmx.output"
	os.Remove(outputFile) // Not checking output, don't care

	// Build up command; have to use 'sh' to run due to arg issues
	v := "java -jar jmxterm-1.0.0-uber.jar -n -e -i jmx.input -o jmx.output"
	cmd := exec.Command("sh", "-c", v)
	if flagVerbose {
		fmt.Printf("Starting JMX run ... with: %s %s\n\n", cmd.Path, cmd.Args)
	}
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s\n", string(stdoutStderr))
		log.Fatal(err)
	}

	// Read from the JMX output file
	var scanner *bufio.Scanner
	fr, err := os.Open(outputFile)
	checkErr(err)
	defer fr.Close()
	scanner = bufio.NewScanner(fr)
	//scanner.Split(bufio.ScanWords) // Split words

	c := 0
	for scanner.Scan() {
		err = scanner.Err()
		checkErr(err)
		s := scanner.Text()
		if len(s) > 0 { // Skip blank lines
			fmt.Printf("Scan text: %s\n", s)
			s = strings.TrimRight(s, ";") // Remove trailing semicolon
			// Get the last value
			fields := strings.Fields(s) // Split on spaces
			lastField := fields[len(fields)-1]
			vals[c], err = strconv.Atoi(lastField) //ParseFloat(lastField, 64)
			checkErr(err)
			c++
		}
	}

	e := error(nil) // No errors for now
	return vals, e

} // runJMX()

// Check if Java is here, jar is here, etc.
//func checkLsExists() {
//	path, err := exec.LookPath("ls")
//	if err != nil {
//		fmt.Printf("didn't find 'ls' executable\n")

type jmxConnectInfo struct {
	Mode     string // P or T for PID or TCP
	PID      int    // PID
	Host     string // IP or DNS name
	Port     int
	User     string
	Password string
}

type beanInfo struct {
	Domain    string
	MbeanName string
	Attribute string
	Value     string
}

//-a --appendto            With this flag, the outputfile is preserved and content is appended to it
//-e --exitonfa            With this flag, terminal exits for any Exception
//-h --help                Show usage of this command line
//-i --input    <value>    Input script file. There can only be one input file. "stdin" is the default value which means console input
//-n --noninter            Non interactive mode. Use this mode if input doesn't come from human or jmxterm is embedded
//-o --output   <value>    Output file, stdout or stderr. Default value is stdout
//-p --password <value>    Password for user/password authentication
//-l --url      <value>    Location of MBean service. It can be <host>:<port> or full service URL.
//-u --user     <value>    User name for user/password authentication
//-v --verbose  <value>    Verbose level, could be silent|brief|verbose. Default value is brief

//about    - Display about page
//bean     - Display or set current selected MBean.
//beans    - List available beans under a domain or all domains
//bye      - Terminate console and exit
//close    - Close current JMX connection
//domain   - Display or set current selected domain.
//domains  - List all available domain names
//exit     - Terminate console and exit
//get      - Get value of MBean attribute(s)
//help     - Display available commands or usage of a command
//info     - Display detail information about an MBean
//jvms     - List all running local JVM processes
//open     - Open JMX session or display current connection
//option   - Set options for command session
//quit     - Terminate console and exit
//run      - Invoke an MBean operation
//set      - Set value of an MBean attribute

// java -jar jmxterm-1.0.0-uber.jar -n -e -i jmx.input -o jmx.output

//	cmd := exec.Command("java", "-jar", "jmxterm-1.0.0-uber.jar", "-n", "-e", "-i jmx.input", "-o jmx.output")
// works with far for SH or singel 3rd arg to command for sh
//cmd := exec.Command("sh", "-c", "java -jar jmxterm-1.0.0-uber.jar -n -e -i jmx.input -o jmx.output")
