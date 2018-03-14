package main

// TODO
// Later - option to not to delta processing in tool, just return counts

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	flag "github.com/ogier/pflag"
)

// constants
const (
	version   = "0.1"
	copyright = "Copyright 2018 by OpsStack"
)

// Global vars
var (
	argProcessName   string
	argProcessorName string
	//argServerServer   string
	//argServerPort     string
	argServerUser     string
	argServerPassword string
	argStatsMetric    string
	argStatusFileName string
	argCredFileName   string
	flagBeginning     bool
	flagVerbose       bool
	flagHelp          bool
	flagJMXTerm       string
	flagJMXinputFile  string
	flagJMXoutputFile string
)

// init is called automatically at start
func init() {

	// Setup arguments, must do before calling Parse()
	flag.StringVarP(&argProcessName, "process", "p", "", "Process Name")
	flag.StringVarP(&argProcessorName, "processor", "r", "", "Request Processor Name")
	//flag.StringVarP(&argServerServer, "server", "S", "127.0.0.1", "Server Host")
	//flag.StringVarP(&argServerPort, "port", "P", "3306", "Server Port")
	//flag.StringVarP(&argServerUser, "user", "u", "", "User")
	//flag.StringVarP(&argServerPassword, "password", "w", "", "Password")
	flag.StringVarP(&argStatsMetric, "metric", "m", "r", "Metric Type")
	flag.StringVarP(&argStatusFileName, "statusfile", "f", "", "Status File")
	flag.StringVarP(&argCredFileName, "credfile", "c", "", "Credential File")
	flag.BoolVarP(&flagBeginning, "beginning", "b", false, "From Beginning")
	flag.BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose Output")
	//flag.BoolVarP(&flagVeryVerbose, "very-verbose", "w", false, "Very Verbose Output")
	flag.BoolVarP(&flagHelp, "help", "h", false, "Help")
	flag.StringVarP(&flagJMXTerm, "jmxterm", "j", "jmxterm-1.0.0-uber.jar", "path to jmxterm")
	flag.StringVar(&flagJMXinputFile, "jmxinputfile", "jmx.input", "temp file to store input to jmxterm")
	flag.StringVar(&flagJMXoutputFile, "jmxoutputfile", "jmx.output", "temp file to store output of jmxterm")

	flag.Parse() // Process argurments
}

func main() {

	var (
		err            error
		infoFileName   string
		lastRunInfo    [8]int
		jmxResults     [6]int
		jmxUpTime      int
		jmxTimeElapsed int
		latency        int
		pid            int
	)

	startTime := time.Now()

	if flagVerbose {
		fmt.Println()
		fmt.Printf("Tomcat Signals Version %s - %s\n", version, copyright)
		fmt.Printf("Starting at: %s\n", startTime.Format(time.UnixDate))
		if argServerPassword == "" {
			fmt.Printf("Arguments: %s\n\n", os.Args[1:]) // Skip program name
		} else {
			fmt.Printf("Arguments include password, can't show.\n\n") // Now show if PW here
		}
	}

	// Check our command-line arguments
	argsCheck(version, copyright)

	// Get our last run counters from status file
	infoFileName = argStatusFileName // Need more safety checks before open argument filename?
	lastRunInfo, err = getLastRunInfo(infoFileName)
	checkErr(err)
	// lastRunInfo:
	//  0 - Tomcat PID on last run
	//	1 - Request Count last run time (Rate mode)
	//	2 - Request Count last counter value (Rate mode)
	//	3 - Error Count last run time (Error mode)
	//	4 - Error Count last counter value (Error mode)
	//  5 - Processing Time last run time (Latency mode)
	//	6 - Processing Time last time counter value (Latency mode)
	//	7 - Processing Time last request counter value (Latency mode)

	// Get our Tomcat PID
	pid = getTomcatPID(argProcessName)
	if pid != lastRunInfo[0] {
		// We have a new Tomcat instance
		// Reset last run info to zero
		if flagVerbose {
			fmt.Printf("Tomcat has restarted, old PID: %d, new PID: %d\n", lastRunInfo[0], pid)
		}
		lastRunInfo[0] = pid
		// Set all data to zero, except first element which is PID
		for i := 1; i < 8; i++ {
			lastRunInfo[i] = 0
		}
	} else {
		if flagVerbose {
			if !flagBeginning {
				fmt.Printf("Tomcat still running, PID: %d, using last run status info.\n", lastRunInfo[0])
			} else {
				fmt.Printf("Tomcat still running, PID: %d, but have -b beginning flag, so not using last run status info.\n", lastRunInfo[0])
			}
		}
	}

	jmxResults, err = runJMX(pid, flagJMXTerm, flagJMXinputFile, flagJMXoutputFile)
	checkErr(err)
	// jmxReults[]:
	// 0 - Uptime
	// 1 - Request Count
	// 2 - Error Count
	// 3 - Processing Time (ms)
	// 4 - Threads Busy
	// 5 - Threads Max

	jmxUpTime = jmxResults[0] // int(time.Now().Unix()) // Get after JMX run as JMX takes several seconds

	// Output
	if flagVerbose {
		fmt.Printf("Uptime         : %d\n", jmxResults[0])
		fmt.Printf("Request   Count: %d\n", jmxResults[1])
		fmt.Printf("Error     Count: %d\n", jmxResults[2])
		fmt.Printf("Processing Time: %d\n", jmxResults[3])
		fmt.Printf("Current Threads: %d\n", jmxResults[4])
		fmt.Printf("Max     Threads: %d\n\n", jmxResults[5])
	}

	// Make calculations & display results
	switch argStatsMetric {
	case "r":
		if flagBeginning { // Don't use old stats, set to 0
			lastRunInfo[1] = 0
			lastRunInfo[2] = 0
		}
		jmxTimeElapsed = jmxUpTime - lastRunInfo[1]
		newRequests := jmxResults[1] - lastRunInfo[2]
		rate := float64(newRequests / jmxTimeElapsed * 1000)
		// Update status info
		lastRunInfo[1] = jmxUpTime
		lastRunInfo[2] = jmxResults[1]

		if flagVerbose {
			fmt.Printf("newRequests: %d\n", newRequests)
			fmt.Printf("Elapsed Time: %d ms\n", jmxTimeElapsed)
			fmt.Printf("Rate: %f/sec\n", rate)
			fmt.Printf("\nStatus File values\n")
			fmt.Printf("Uptime         : %d\n", lastRunInfo[1])
			fmt.Printf("Request   Count: %d\n", lastRunInfo[2])
		} else {
			fmt.Printf("%f", rate)
		}

	case "e":
		if flagBeginning { // Don't use old stats, set to 0
			lastRunInfo[3] = 0
			lastRunInfo[4] = 0
		}
		jmxTimeElapsed = jmxUpTime - lastRunInfo[3]
		newErrors := jmxResults[2] - lastRunInfo[4]
		rate := float64(newErrors / jmxTimeElapsed * 1000)
		// Update status info
		lastRunInfo[3] = jmxUpTime
		lastRunInfo[4] = jmxResults[2]

		if flagVerbose {
			fmt.Printf("Elapsed Time: %d ms\n", jmxTimeElapsed)
			fmt.Printf("Error Rate: %f/sec\n", rate)
		} else {
			fmt.Printf("%f", rate)
		}

	case "l": // Note this includes errors (as it's a Request count)
		if flagBeginning { // Don't use old stats, set to 0
			lastRunInfo[7] = 0
			lastRunInfo[6] = 0
		}
		newRequests := jmxResults[1] - lastRunInfo[7]
		newProcTime := jmxResults[3] - lastRunInfo[6]
		if newRequests > 0 {
			latency = int(newProcTime / newRequests)
		} else {
			latency = 0.0
		}
		// Update status info
		lastRunInfo[7] = jmxResults[1]
		lastRunInfo[6] = jmxResults[3]

		if flagVerbose {
			fmt.Printf("New Requests: %d\n", newRequests)
			fmt.Printf("New Procsssing Time: %d ms\n", newProcTime)
			fmt.Printf("Latency: %d ms\n", latency)
		} else {
			fmt.Printf("%d", latency)
		}

	case "u": // Does not use saved status info
		utilization := float64(jmxResults[4] / jmxResults[5] * 100) // Get %

		if flagVerbose {
			fmt.Printf("Utilization: %4.1f%%\n", utilization)
		} else {
			fmt.Printf("%f", utilization)
		}

	default:
		panic("Invdalid Stats Metric in Calc/Ouput Section")
	} // Switch on argStatsMetric

	// Save data back to status file
	err = saveLastRunInfo(infoFileName, lastRunInfo)
	checkErr(err)

	endTime := time.Now()

	if flagVerbose {
		fmt.Println()
		fmt.Printf("Ending at: %s\n", endTime.Format(time.UnixDate))
		fmt.Println()
	}

	// Exit normally
	os.Exit(0)
} // Main

// Process arguments
func argsCheck(version string, copyright string) {

	// Tomcat 7
	//n := "tomcat"
	//p := "http-bio-8080"

	//// Tomcat 8.0 & 8.5
	//n := "tomcat"
	//p := "http-nio-8080"

	//// Tomcat 8.5
	//n := "tomcat"
	//p := "http-nio-8080"

	//// Tomcat 9
	//n := "catalina"
	//p := "http-nio-8080"

	if argProcessName == "" {
		argProcessName = "catalina" // For Tomcat
	}

	if argProcessorName == "" {
		argProcessorName = "http-nio-8080" // For Tomcat 9, maybe others
	}

	if flagHelp {
		fmt.Printf("GoldenWebReader Version %s - %s\n\n", version, copyright)
		fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Require metric type
	// Require metric type
	if argStatsMetric == "" {
		log.Fatalln("Stats Metric missing - should be r, e, l, or u")
		os.Exit(1)
	}
	if argStatsMetric != "r" && argStatsMetric != "e" &&
		argStatsMetric != "l" && argStatsMetric != "u" {
		log.Fatalln("Stats Metric not valid - should be r, e, l, or u")
		os.Exit(1)
	}

	// Can have cred file OR user/password
	if argCredFileName != "" && (argServerUser != "" || argServerPassword != "") {
		log.Fatalln("Cannot supply BOTH a credential file and a user or password.")
		os.Exit(1)
	}
}

// Error checking for various things
func checkErr(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}

// Get Tomcat's PID
func getTomcatPID(name string) int {
	var (
		pid   int
		myPid int
	)

	//OS := runtime.GOOS // If needed later
	myPid = os.Getpid() // Needed for pgprep filtering

	// If a process, get the PID
	if argProcessName != "" {

		if flagVerbose {
			fmt.Printf("Getting PID for Process: %s\n", name)
			fmt.Printf("Note my PID is: %d\n", myPid)
		}

		// Realy need to check/safe this argument going to the OS
		// Note -i is okay on Mac, but not on Linux for pgrep, so we are case SENSITIVE
		//v := "pgrep -f " + name

		v := "ps -A -o pid,comm,args | grep java | grep " + name +
			" | grep -v grep | awk '{print $1}'"

		cmd := exec.Command("sh", "-c", v)
		stdoutStderr, err := cmd.CombinedOutput()
		if err == nil {
			pid, err = strconv.Atoi(strings.TrimSpace(string(stdoutStderr))) //Remove trailing newline char (\n)
			if pid == 0 {                                                    // If we got here and have PID = 0, then we got multiple matches
				fmt.Printf("Found 0 or Several PIDs for process: %s\n", name)
				fmt.Println("- This means you have 0 or >1 running. following command must return only one PID.")
				fmt.Printf("   %s:\n%s\n", v, string(stdoutStderr))
				fmt.Println("- Exiting.")
				os.Exit(1)
			}
			if flagVerbose {
				fmt.Printf("PID for Process: %s is: %d\n", name, pid)
			}
		} else {
			fmt.Println("Error in exec call to ps, grep, and awk.")
			log.Fatal(err)
		}
	} // Arg name not blank

	return pid
} // getTomcatPID
