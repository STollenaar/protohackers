package problem11

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"protohackers/util"
	"syscall"
	"time"
)

//Policies are site -> species -> action

type Policy struct {
	Action   ActionType `json:"action"`
	PolicyId uint32     `json:"policyId"`
}

var (
	server ServerWithReader
	sites  SiteMap //uint32, TargetPopulation
)

const (
	autorityServer = "pestcontrol.protohackers.com"
	autorityPort   = "20547"
)

func init() {
	server = ServerWithReader{
		util.ServerTCP{
			ConnectionHandler: handle,
		},
	}
	sites = SiteMap{Site: make(map[uint32]*Site)}
}

func dump() {

	data, err := json.MarshalIndent(&sites, "", " ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("sites.json", data, 0644)
	siteVisits := make(map[uint32][]PopulationVisit)
	for k, site := range sites.Site {
		siteVisits[k] = site.siteVisits
	}
	data, _ = json.MarshalIndent(&siteVisits, "", " ")
	os.WriteFile("siteVisits.json", data, 0644)
}

func Problem() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		dump()
		os.Exit(1)
	}()

	server.Start()
}

func handle(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	defer conn.Close()
	client := &Client{
		conn: conn,
		reader: serverReader{
			r: bufio.NewReader(conn),
		},
		writer: serverWriter{
			w: bufio.NewWriter(conn),
		},
	}
	initial := true
	message := HelloMessage{
		Protocol: "pestcontrol",
		Version:  1,
	}
	message.Marshal(client.writer.w)
	for {
		m, err := server.ReadMessage(client)
		fmt.Println(m, err)
		if m == nil {
			m = ErrorMessage{
				Message: err.Error(),
			}
		}
		if initial && m.Type() != HelloMessageType && m.Type() != ErrorMessageType {
			m = ErrorMessage{Message: "Initial message must be hello"}
		}
		if err != nil || m.Type() == ErrorMessageType {
			fmt.Println("Sending error", m)
			m.Marshal(client.writer.w)
			if err == io.EOF {
				break
			}
			continue
		}

		switch m.Type() {
		case HelloMessageType:
			initial = false
		case SiteVisitType:
			// Handle
			sites.mu.Lock()
			c, ok := sites.Site[m.(SiteVisitMessage).site]
			if !ok {

				site := createSite(m.(SiteVisitMessage).site)
				c = site
			}
			c.svChan <- m.(SiteVisitMessage)
			sites.mu.Unlock()
		default:
			e := ErrorMessage{Message: "Unknown Message Type for client"}
			fmt.Println(m)
			e.Marshal(client.writer.w)
			return
		}
	}
}
