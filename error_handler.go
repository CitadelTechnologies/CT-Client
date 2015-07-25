package GleipnirServer

import(
    "os"
)

func CheckError(err error) {
    if err != nil {
        f, _ := os.OpenFile("errors.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0660)
        defer f.Close()
        f.WriteString(err.Error())
        os.Exit(2)
    }
}