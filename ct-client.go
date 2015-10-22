// Package to create an API compatible with CT-Core
// Contains some utils to launch the program and communicate with the Kernel
package ctclient

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
    CtClient struct {
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

var Server CtClient

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

func (client *CtClient) connect() {
    var err error
    client.Conn, err = net.Dial("tcp", "0.0.0.0:" + client.KernelPort)
    CheckError(err)

    client.Status.StartedAt = time.Now()
    client.writeToKernel("connect")
    if response := client.readFromKernel(); response.Status != 200 {
        CheckError(errors.New(response.Message))
    }
}

func (client *CtClient) Shutdown() {
    Server.Conn.Close()
    os.Exit(2)
}

func (client *CtClient) refreshStatus() {
     runtime.ReadMemStats(&client.Status.MemoryStats)
     client.Status.UpdatedAt = time.Now()
}

/*
 *  This method accepts encoded JSON and send it to the Kernel
 */
func (client *CtClient) writeToKernel(command string) {
    client.refreshStatus()

    status := Status {
        StartedAt: client.Status.StartedAt,
        UpdatedAt: client.Status.UpdatedAt,
        HeapAlloc: client.Status.MemoryStats.HeapAlloc,
        HeapSys: client.Status.MemoryStats.HeapSys,
        EnableGC: client.Status.MemoryStats.EnableGC,
        LastGC: client.Status.MemoryStats.LastGC,
        NextGC: client.Status.MemoryStats.NextGC,
        NumGC: client.Status.MemoryStats.NumGC,
    }
    message := Message {
        Command: command,
        Emmitter: client.Token,
        Status: status,
    }

    data := make([]byte, unsafe.Sizeof(message))
    var err error
    data, err = json.Marshal(message)
    CheckError(err)

    _, err = client.Conn.Write(data)
    CheckError(err)
}

func (client *CtClient) readFromKernel() Response {

    var response Response
    buffer := make([]byte, unsafe.Sizeof(response))

    _, err := client.Conn.Read(buffer)
    CheckError(err)

    json.Unmarshal(buffer, &response)

    return response
}