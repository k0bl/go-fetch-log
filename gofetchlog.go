package main
import (
        "bufio"
        "fmt"
        "os"
        "flag"
        "strconv"
        "strings"
        // "regexp"
)
var bookmarkfilemsg = "bookmark file where we store log file position"
var regexmsg = "regexp to search in log file since last run"
var lastmsg = "start at the end of the log file if no bookmark file"
var countmsg = "return a count of matching lines instead of line output"

var bookmarkfile = flag.String("bookmarkfile", "", "")
var logfile      = flag.String("logfile", "", bookmarkfilemsg)
var regexpr       = flag.String("regexp", "", regexmsg)
var count       = flag.Bool("count", false, countmsg)
var processedLen int64 = 0

//result counter if we are using count
var resultCount int64 = 0

var logf *os.File

func main() {
    flag.Parse()
    // fmt.Println(*bookmarkfile)
    // fmt.Println(*logfile)
    // fmt.Println(*regexpr)
    var lastpos int64 = 0
    //check if we have a logfile
    if _, err := os.Stat(*logfile); err == nil {
        //we have a logfile, open and defer close
        fileHandle, _ := os.Open(*logfile)
        logf = fileHandle
        defer fileHandle.Close()
        //check if we have a bookmark file
        if _, bmerr := os.Stat(*bookmarkfile); bmerr == nil {
            //open bookmark file
            bookmarkFileHandle, _ := os.Open(*bookmarkfile)
            defer bookmarkFileHandle.Close()
            //find last position from bookmark in log file
            bmScan := bufio.NewScanner(bookmarkFileHandle)
            const maxCapacity = 512*1024 
            buf := make([]byte, maxCapacity)
            bmScan.Buffer(buf, maxCapacity)

            for bmScan.Scan() {
                innerlastpos, bmscanerr := strconv.ParseInt(bmScan.Text(), 10, 64)
                if bmscanerr != nil {
                    fmt.Println("Error: ", bmscanerr)
                }
                lastpos = lastpos + innerlastpos
            }

            if lastpos > 0 {
                //check if length of file is greater than lastpos
                checkFile, ckfierr := fileHandle.Stat()
                if ckfierr != nil {
                    // Could not obtain stat, handle error
                }
                //check if file size is greater than bookmark
                if checkFile.Size() >= int64(lastpos) {
                    //file size equal or greater, log has not rotated
                    //start from last position
                    processedLen = lastpos
                    processFileFromLastPosition(lastpos)
                } else {
                    //file size is smaller, log rotated. start from beginning
                    processFileFromStartPosition(0)
                }
            } else {
                //lastpos is 0, process file from start position
                processFileFromStartPosition(0)
            }
        } else {
            //bookmark file does not exist, process file from start position
            processFileFromStartPosition(0)
        }
        os.Exit(0)
    } else {
        fmt.Printf("log file %s does not exist.\n", *logfile)
        flag.PrintDefaults()
        os.Exit(2)
    }   
}
func processFileFromStartPosition(lastpos int64) {
    fileScanner := bufio.NewScanner(logf)
    //custom buffer for large logs
    const maxCapacity = 512*1024 
    buf := make([]byte, maxCapacity)
    fileScanner.Buffer(buf, maxCapacity)
    for fileScanner.Scan() {
        // pass each line to checkRegex
        checkRegEx(fileScanner.Text())
        // add 1 for trailing whitespace, need a better solution
        processedLen = processedLen + int64(len(fileScanner.Bytes())+1)
        // fmt.Println("processedLen", processedLen)
    }
    updateBookmarkFile(processedLen)
    if (*count) {
        fmt.Println(resultCount);
    }
}
func processFileFromLastPosition(lastpos int64) {
    var offset int64 = int64(lastpos)
    var whence int = 0
    //seek to new position in log file
    newPosition, poserr := logf.Seek(offset, whence)

    if poserr != nil {
        fmt.Println("Attempted to seek to: ", newPosition)
        fmt.Println("Error: ", poserr)
    }
    fileScanner := bufio.NewScanner(logf)
    const maxCapacity = 512*1024 
    buf := make([]byte, maxCapacity)
    fileScanner.Buffer(buf, maxCapacity)
    for fileScanner.Scan() {
        //pass each line to checkRegex
        checkRegEx(fileScanner.Text())
        // add 1 for trailing whitespace, need a better solution
        processedLen = processedLen + int64(len(fileScanner.Bytes())+1)
        // fmt.Println("processedLen", processedLen)
    }
    updateBookmarkFile(processedLen)
    if (*count) {
        fmt.Println(resultCount);
    }
}
func checkRegEx(text string) {
    // had to disable regex. Too slow
    // match, _ := regexp.MatchString(*regexpr, text)
    // if match == true {
    //     fmt.Println(text)
    // }

    //strings.Contains much faster for this purpose
    stringSlice := strings.Split(*regexpr, "|")
    for _, v := range stringSlice {       
        if strings.Contains(text, v) {
            if (*count) {
                resultCount +=1;
            } else {
                fmt.Println(text)
            }
        }
    }

}
func updateBookmarkFile(processedLen int64) {
    //create new bm file every time, wipe out old
    bookmarkFileHandle, werr := os.Create(*bookmarkfile)
    if werr != nil {
        fmt.Println("Cannot write file", werr)
    }
    defer bookmarkFileHandle.Close()        
    bmString := strconv.FormatInt(processedLen, 10)
    //write new processedLen to bookmark file
    fmt.Fprintf(bookmarkFileHandle, bmString)
}
