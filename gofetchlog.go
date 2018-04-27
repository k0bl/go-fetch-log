package main
import (
        "bufio"
        "fmt"
        "os"
        "flag"
        "strconv"
        "regexp"
)
var bookmarkfilemsg = "bookmark file where we store log file position"
var regexmsg = "regexp to search in log file since last run"
var lastmsg = "start at the end of the log file if no bookmark file"
var bookmarkfile = flag.String("bookmarkfile", "", "")
var logfile      = flag.String("logfile", "", bookmarkfilemsg)
var regexpr       = flag.String("regexp", "", regexmsg)
var last         = flag.String("L", "", lastmsg)
var processedLen = 0

var logf *os.File

func main() {
    flag.Parse()
    fmt.Println(*bookmarkfile)
    fmt.Println(*logfile)
    fmt.Println(*regexpr)
    var lastpos int = 0
    //check if we have a logfile
    if _, err := os.Stat(*logfile); err == nil {
        //we have a logfile, open and defer close
        fileHandle, _ := os.Open(*logfile)
        logf = fileHandle
        defer fileHandle.Close()
        //check if we have a bookmark file
        if _, bmerr := os.Stat(*bookmarkfile); bmerr == nil {

            fmt.Println("Attempting to open bookmark file")
            //open bookmark file
            bookmarkFileHandle, _ := os.Open(*bookmarkfile)
            defer bookmarkFileHandle.Close()
            //find last position from bookmark in log file
            bmScan := bufio.NewScanner(bookmarkFileHandle)
            for bmScan.Scan() {
                fmt.Println("bmScan", bmScan.Text())
                innerlastpos, bmscanerr := strconv.Atoi(bmScan.Text())
                if bmscanerr != nil {
                    fmt.Println("Error: ", bmscanerr)
                }
                fmt.Println("lastpos", lastpos  )
                lastpos = lastpos + innerlastpos
            }
            fmt.Println("lastpos outside", lastpos)
            fmt.Println("bookmarkfile, starting from marked position")
            if lastpos > 0 {
                /* There was a bookmark file with a value greater than 0
                 store last position and process file from last position */
                processedLen = lastpos
                fmt.Println("Pre-existing offset found", processedLen)
                processFileFromLastPosition(lastpos)
            } else {
                //lastpos is 0, process file from start position
                processFileFromStartPosition(0)
            }
        } else {
            //bookmark file does not exist, process file from start position
            processFileFromStartPosition(0)
        }
    } else {
        fmt.Printf("log file %s does not exist.\n", *logfile)
        flag.PrintDefaults()
        os.Exit(2)
    }   
}
func processFileFromStartPosition(lastpos int) {
    fileScanner := bufio.NewScanner(logf)
    for fileScanner.Scan() {
        // pass each line to checkRegex
        checkRegEx(fileScanner.Text())
        processedLen = processedLen + (len(fileScanner.Text())+2)
    }
    updateBookmarkFile(processedLen)
}
func processFileFromLastPosition(lastpos int) {
    var offset int64 = int64(lastpos)
    fmt.Println("offset", offset)
    var whence int = 0
    newPosition, poserr := logf.Seek(offset, whence)
    if poserr != nil {
        fmt.Println("Error: ", poserr)
    }
    fmt.Println("Just moved to:", newPosition)
    fileScanner := bufio.NewScanner(logf)
    for fileScanner.Scan() {
        //pass each line to checkRegex
        checkRegEx(fileScanner.Text())
        processedLen = (processedLen + len(fileScanner.Text())+2)
    }
    updateBookmarkFile(processedLen)
}
func checkRegEx(text string) {
    match, _ := regexp.MatchString(*regexpr, text)
    if match == true {
        fmt.Println(text)
    }
}
func updateBookmarkFile(processedLen int) {
    fmt.Println("updateBookmarkFile processedLen", processedLen)
    // err := os.Truncate(*bookmarkfile, 0)
    // if err != nil {
    //     fmt.Println(err)
    // }
    bookmarkFileHandle, werr := os.Create(*bookmarkfile)
    if werr != nil {
        fmt.Println("Cannot write file", werr)
    }
    defer bookmarkFileHandle.Close()        
    bmString := strconv.Itoa(processedLen)
    fmt.Fprintf(bookmarkFileHandle, bmString)
}
