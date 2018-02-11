# TomcatSignals
Golden Tomcat Metric Tool

This is a Linux command-line tool written in Go for collecting stats from Apache Tomcat via JMX. It requires the jmxterm
command line java tool.

## BETA RELEASE 
Still in testing and features and/or output may change.
In addition, the code is not fully cleaned up.

Usage:

tomcatsignals [-m metric] [-b] [-v] [-h]

  The options are as follows:

       -p      Tomcat Process Name pattern. Default is 'catalina'. Tomcat ver7-8.5 often uses 'tomcat'
                    Note pgrep -f 'name' must return PID of Tomcat 
       -r      Request Processor Name. Default is 'http-nio-8080'. Tomcat ver7 often uses 'http-bio-8080'
                    Must match processor in your Tomcat config files.
       -s      Status File Name.  Default is 'status.file'. Must be writable.       
       -m      Metric to produce: (c)ount, (r)ate, (e)errors, (l)atency
 
       -b      Start from beginning of file. Default off.
       -v      Verbose, for debugging and more info. Default off.
       -h      Help and usage.

## TODO
- Handle Tomcat restarts

## Contributing
We are not ready for contributors until we can get the code cleaned up and standardized for Go best practices.

However, you can contribute by:
- [Report bugs](https://github.com/opsstack/weblog-signals/issues/new)
- [Improve documentation](https://github.com/opsstack/weblog-signals/issues?q=is%3Aopen+label%3Adocumentation)
- [Review code and feature proposals](https://github.com/opsstack/weblog-signals/pulls)

## Installation:

You can download the binaries directly from the [binaries](https://github.com/opsstack/weblog-signals/binaries) section.  We'll have RPMs and DEB packages as soon as things stabilize a bit.

### From Source:

This is a single source file project for now, so you can just compile as you would any Golang project.

There is a single external dependency, [pflag](https://github.com/ogier/pflag)

## How to use it:

See usage with:

```
./goldweblog --help
```
