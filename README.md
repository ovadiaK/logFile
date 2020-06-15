# logFile

logFile is a lightweight struct wrapping *log.Logger with some more functionality:
 
* log file gets replaced if lost or full.
* supplements debug information about the caller and caller's caller

### New()

Creates new file if necessary and returns new LogFile.
This function should be called to initiate a new LogFile.
Several Log files with different specifications can be created.
The file will be created in the target path with name and current time.
If maxLines is 0, defaults to 65535, should be less than 10mb for one or two line logs entries.

All methods should be called on the log file.


#### Log()

logs message like log.Println()

#### Error()

logs error like log.Println() with debug info about caller and caller's caller

#### Panic()
initiates and logs panic like log.Panic() with debug info about caller and caller's caller
#### Fatal()
calls and logs os.Exit like log.Fatal() with debug info about caller and caller's caller

## plans for the future

implement automated sending of logs per email deleting them after they are full