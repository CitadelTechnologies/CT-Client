// Package to create an API compatible with Gleipnir
// Contains some utils to launch the program and communicate with the Kernel
package GleipnirServer

import(
    "flag"
    "net"
    "runtime"
    "time"
    "encoding/json"
)

type(

    GleipnirServer struct {

        Conn net.TCPConn
        TokenId string
        KernelPort string
        DedicatedPort string
        Status Status

    }
    Message struct {

        Command string `json:"command"`
        Emmitter string `json:"emmitter"`
        Status Status `json:"status"`

    }
    Status struct {

        StartedAt time.Time `json:"started_at"`
        UpdatedAt time.Time `json:"updated_at"`
        MemoryStats runtime.MemStats `json:"memory_stats"`

    }
)

var Server GleipnirServer

func init() {

    flag.StringVar(&Server.DedicatedPort, "port", "0", "The Server port")
    flag.StringVar(&Server.KernelPort, "port", "0", "The Server port")
    flag.Parse()

    if(Server.DedicatedPort == "0") {

        panic("The API port flag must be given")

    }

    if(Server.KernelPort == "0") {

        panic("The Kernel port flag must be given")

    }

    var error err
    if Server.Conn, err = net.Dial("tcp", ":" + gs.KernelPort); err != nil {
        panic(err)
    }

    Server.Status.StartedAt = time.Now()
    Server.writeToKernel()

}

func Shutdown() {

    Server.Conn.Close()

}

func (gs *GleipnirServer) refreshStatus() {

     runtime.ReadMemStats(&gs.Status.MemoryStats)
     gs.Status.UpdatedAt = time.Now()

}

// This method accepts encoded JSON and send it to the Kernel
func (gs *GleipnirServer) writeToKernel(command string) {
    gs.refreshStatus()

    message := Message{Command: command, Emmitter: gs.TokenId, Status: gs.Status}

    if json, err := json.Marshal(message); err != nil {
        panic(err)
    }

    if _, err = gs.Conn.Write(json); err != nil {
        panic(err)
    }
}

func (gs *GleipnirServer) readFromKernel() []byte {

    buffer := make([]byte, 2048)

    if _, err := gs.Conn.Read(&buffer); err != nil {
        panic(err)
    }

    return buffer
}