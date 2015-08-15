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
    "unsafe"
    "errors"
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
        Status int `json:"status"`
        Message string `json:"message"`
    }
    ServerStatus struct {
        StartedAt time.Time
        UpdatedAt time.Time `json:"updated_at"`
        MemoryStats runtime.MemStats
    }
    Status struct {
        StartedAt time.Time `json:"started_at"`
        UpdatedAt time.Time `json:"updated_at"`
        HeapAlloc uint64 `json:"heap_alloc"`
        HeapSys uint64 `json:"heap_sys"`
        EnableGC bool `json:"enable_gc"`
        LastGC uint64 `json:"last_gc"`
        NextGC uint64 `json:"next_gc"`
        NumGC uint32 `json:"num_gc"`
    }
)

var Server GleipnirServer

func Initialize() {
    flag.StringVar(&Server.Token, "token", "0", "The Service token")
    flag.StringVar(&Server.DedicatedPort, "service-port", "0", "The Service port")
    flag.StringVar(&Server.KernelPort, "kernel-port", "0", "The Kernel port")
    flag.Parse()

    if Server.Token == "0" {
        CheckError(errors.New("The service token flag must be given"))
    }
    if Server.DedicatedPort == "0" {
        CheckError(errors.New("The API port flag must be given"))
    }
    if Server.KernelPort == "0" {
        CheckError(errors.New("The Kernel port flag must be given"))
    }
    Server.connect()
}

func (s *GleipnirServer) connect() {
    var err error
    s.Conn, err = net.Dial("tcp", "0.0.0.0:" + s.KernelPort)
    CheckError(err)

    s.Status.StartedAt = time.Now()
    s.writeToKernel("connect")
    if response := s.readFromKernel(); response.Status != 200 {
        CheckError(errors.New(response.Message))
    }
}

func (s *GleipnirServer) Shutdown() {
    Server.Conn.Close()
    os.Exit(2)
}

func (gs *GleipnirServer) refreshStatus() {
     runtime.ReadMemStats(&gs.Status.MemoryStats)
     gs.Status.UpdatedAt = time.Now()
}

/*
 *  This method accepts encoded JSON and send it to the Kernel
 */
func (gs *GleipnirServer) writeToKernel(command string) {
    gs.refreshStatus()

    status := Status {
        StartedAt: gs.Status.StartedAt,
        UpdatedAt: gs.Status.UpdatedAt,
        HeapAlloc: gs.Status.MemoryStats.HeapAlloc,
        HeapSys: gs.Status.MemoryStats.HeapSys,
        EnableGC: gs.Status.MemoryStats.EnableGC,
        LastGC: gs.Status.MemoryStats.LastGC,
        NextGC: gs.Status.MemoryStats.NextGC,
        NumGC: gs.Status.MemoryStats.NumGC,
    }
    message := Message {
        Command: command,
        Emmitter: gs.Token,
        Status: status,
    }

    data := make([]byte, unsafe.Sizeof(message))
    var err error
    data, err = json.Marshal(message)
    CheckError(err)

    _, err = gs.Conn.Write(data)
    CheckError(err)
}

func (gs *GleipnirServer) readFromKernel() Response {

    var response Response
    buffer := make([]byte, unsafe.Sizeof(response))

    _, err := gs.Conn.Read(buffer)
    CheckError(err)

    json.Unmarshal(buffer, &response)

    return response
}