// Package to create an API compatible with Gleipnir
// Contains some utils to launch the program and communicate with the Kernel
package GleipnirServer

import(
    "os"
    "flag"
    "net"
    "runtime"
    "time"
    "encoding/json"
    "encoding/gob"
    "bytes"
    "unsafe"
)

type(

    GleipnirServer struct {

        Conn net.Conn
        Token string
        KernelPort string
        DedicatedPort string
        Status ServerStatus
    }
    Message struct {
        Command string `json:"command"`
        Emmitter string `json:"emmitter"`
        Status Status `json:"status"`
    }
    Response struct {
        Command string `json:"command"`
    }
    ServerStatus struct {
        StartedAt time.Time
        UpdatedAt time.Time `json:"updated_at"`
        MemoryStats runtime.MemStats
    }
    Status struct {
        StartedAt time.Time `json:"started_at"`
        UpdatedAt time.Time `json:"updated_at"`
        ConsumedMemory uint64 `json:"consumed_memory"`
        AllocatedMemory uint64 `json:"allocated_memory"`
    }
)

var Server GleipnirServer

func Initialize() {

    defer func(){
        if r := recover(); r != nil {
            var buf bytes.Buffer
            enc := gob.NewEncoder(&buf)
            enc.Encode(r)

            f, _ := os.Create("errors.txt")
            defer f.Close()
            f.Write(buf.Bytes())
            os.Exit(2)
        }
    }()

    flag.StringVar(&Server.Token, "token", "0", "The Service token")
    flag.StringVar(&Server.DedicatedPort, "service-port", "0", "The Service port")
    flag.StringVar(&Server.KernelPort, "kernel-port", "0", "The Kernel port")
    flag.Parse()

    if Server.Token == "0" {
        panic("The service token flag must be given")
    }
    if Server.DedicatedPort == "0" {
        panic("The API port flag must be given")
    }
    if Server.KernelPort == "0" {
        panic("The Kernel port flag must be given")
    }
    var err error
    if Server.Conn, err = net.Dial("tcp", "127.0.0.1:" + Server.KernelPort); err != nil {
        panic(err)
    }
    Server.Status.StartedAt = time.Now()
    Server.writeToKernel("connect")

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

    status := Status {
        StartedAt: gs.Status.StartedAt,
        UpdatedAt: gs.Status.UpdatedAt,
        ConsumedMemory: gs.Status.MemoryStats.HeapAlloc,
        AllocatedMemory: gs.Status.MemoryStats.HeapSys,
    }
    message := Message {
        Command: command,
        Emmitter: gs.Token,
        Status: status,
    }

    data := make([]byte, unsafe.Sizeof(message))
    var err error
    if data, err = json.Marshal(message); err != nil {
        panic(err)
    }

    if _, err := gs.Conn.Write(data); err != nil {
        panic(err)
    }
}

func (gs *GleipnirServer) readFromKernel() []byte {

    var response Response
    buffer := make([]byte, unsafe.Sizeof(response))

    if _, err := gs.Conn.Read(buffer); err != nil {
        panic(err)
    }

    json.Unmarshal(buffer, &response)

    return buffer
}