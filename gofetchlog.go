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
    // fmt.Println(*bookmarkfile)
    // fmt.Println(*logfile)
    // fmt.Println(*regexpr)
    var lastpos int = 0
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
            for bmScan.Scan() {
                innerlastpos, bmscanerr := strconv.Atoi(bmScan.Text())
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
func processFileFromStartPosition(lastpos int) {
    fileScanner := bufio.NewScanner(logf)
    for fileScanner.Scan() {
        // pass each line to checkRegex
        checkRegEx(fileScanner.Text())
        // add 1 for trailing whitespace, need a better solution
        processedLen = processedLen + (len(fileScanner.Bytes())+1)
    }
    updateBookmarkFile(processedLen)
}
func processFileFromLastPosition(lastpos int) {
    var offset int64 = int64(lastpos)
    var whence int = 0
    //seek to new position in log file
    newPosition, poserr := logf.Seek(offset, whence)
    if poserr != nil {
        fmt.Println("Attempted to seek to: ", newPosition)
        fmt.Println("Error: ", poserr)
    }
    fileScanner := bufio.NewScanner(logf)
    for fileScanner.Scan() {
        //pass each line to checkRegex
        checkRegEx(fileScanner.Text())
        // add 1 for trailing whitespace, need a better solution
        processedLen = processedLen + (len(fileScanner.Bytes())+1)
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
    //create new bm file every time, wipe out old
    bookmarkFileHandle, werr := os.Create(*bookmarkfile)
    if werr != nil {
        fmt.Println("Cannot write file", werr)
    }
    defer bookmarkFileHandle.Close()        
    bmString := strconv.Itoa(processedLen)
    //write new processedLen to bookmark file
    fmt.Fprintf(bookmarkFileHandle, bmString)
}
