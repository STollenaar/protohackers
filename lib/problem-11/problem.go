package problem11

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"protohackers/util"
	"sync"
	"syscall"
)

//Policies are site -> species -> action

type Policy struct {
	Action   ActionType `json:"action"`
	PolicyId uint32     `json:"policyId"`
}

type Site struct {
	PopulationTarget PopulationTarget  `json:"populationTargets"`
	Policies         map[string]Policy `json:"policies"`

	svChan    chan SiteVisitMessage
	authority *Client
}

type SiteMap struct {
	mu   sync.Mutex
	Site map[uint32]Site `json:"sites"`
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
	sites = SiteMap{Site: make(map[uint32]Site)}
}

func dump() {

	data, err := json.MarshalIndent(&sites, "", " ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("debug.json", data, 0644)

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

	for {
		m, err := server.ReadMessage(client)
		if err != nil {
			m.Marshal(client.writer.w)
			if err == io.EOF {
				break
			}
			continue
		}
		// fmt.Println(m, err)

		switch m.Type() {
		case HelloMessageType:
			m.Marshal(client.writer.w)
		case SiteVisitType:
			// Handle
			sites.mu.Lock()
			c, ok := sites.Site[m.(SiteVisitMessage).site]
			if !ok {
				dialAuthority(m.(SiteVisitMessage).site)
				site := sites.Site[m.(SiteVisitMessage).site]
				c = site

				go func() {
					for {
						site.handleSiteVisit(<-site.svChan)
					}
				}()
			}
			c.svChan <- m.(SiteVisitMessage)
			sites.mu.Unlock()
		default:
			e := ErrorMessage{Message: "Unknown Message Type for client"}
			e.Marshal(client.writer.w)
		}
	}
}

func (site *Site) handleSiteVisit(s SiteVisitMessage) {
	// fmt.Println(s)
	for _, sv := range s.populations {
		target, err := findTarget(sv, site.PopulationTarget)
		if err != nil {
			continue
		}
		pv, ok := site.Policies[sv.species]
		if sv.count < target.Min {
			// Conserve
			if ok && pv.Action == CullAction {
				// Delete
				delete(site.Policies, sv.species)
				site.sendAuthorityMessage(DeletePolicyMessage{policy: pv.PolicyId})
				site.sendAuthorityMessage(CreatePolicyMessage{species: sv.species, action: ConserveAction})
			} else if !ok {
				// Create Policy
				site.sendAuthorityMessage(CreatePolicyMessage{species: sv.species, action: ConserveAction})
			}
		} else if sv.count > target.Max {
			// Cull
			if ok && pv.Action == ConserveAction {
				// Delete
				delete(site.Policies, sv.species)
				site.sendAuthorityMessage(DeletePolicyMessage{policy: pv.PolicyId})
				site.sendAuthorityMessage(CreatePolicyMessage{species: sv.species, action: CullAction})
			} else if !ok {
				// Create Policy
				site.sendAuthorityMessage(CreatePolicyMessage{species: sv.species, action: CullAction})
			}
		} else if ok {
			// Remove Policy
			delete(site.Policies, sv.species)
			site.sendAuthorityMessage(DeletePolicyMessage{policy: pv.PolicyId})
		}
	}
}

func findTarget(visit PopulationVisit, populationTargets PopulationTarget) (MinMax, error) {
	s, ok := populationTargets[visit.species]
	if !ok {
		return MinMax{}, errors.New("none found")
	}
	return s, nil
}

func dialAuthority(site uint32) (Message, *Client) {
	conn, _ := net.Dial("tcp", autorityServer+":"+autorityPort)
	client := &Client{
		conn: conn,
		reader: serverReader{
			r: bufio.NewReader(conn),
		},
		writer: serverWriter{
			w: bufio.NewWriter(conn),
		},
	}

	HelloMessage{Protocol: "pestcontrol", Version: 1}.Marshal(client.writer.w)
	m, err := server.ReadMessage(client)
	if err != nil {
		fmt.Printf("Error reading helloMessage, %s\n", err)
		if err == io.EOF {
			return nil, nil
		}
		e := ErrorMessage{Message: err.Error()}
		e.Marshal(client.writer.w)
		return nil, nil
	} else if m.Type() == ErrorMessageType {
		fmt.Printf("Error returned helloMessage, %s\n", err)
		m.Marshal(client.writer.w)
		return nil, nil
	}

	DialAuthorityMessage{site: site}.Marshal(client.writer.w)
	m, err = server.ReadMessage(client)
	if err != nil {
		fmt.Printf("Error reading dial message, %s, send with site %d\n", err, site)
		if err == io.EOF {
			return nil, nil
		}
		e := ErrorMessage{Message: err.Error()}
		e.Marshal(client.writer.w)
		return nil, nil
	} else if m.Type() == ErrorMessageType {
		fmt.Printf("Error return dial message, %s, send with site %d\n", err, site)
		m.Marshal(client.writer.w)
		return nil, nil
	}
	mTyped := m.(TargetPopulationMessage)
	sites.Site[site] = Site{
		PopulationTarget: mTyped.Populations,
		Policies:         make(map[string]Policy),
		svChan:           make(chan SiteVisitMessage),
		authority:        client,
	}
	return m, client
}

func (site *Site) sendAuthorityMessage(message Message) Message {
	client := site.authority

	if client == nil {
		return nil
	}
	message.Marshal(client.writer.w)
	m, err := server.ReadMessage(client)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		e := ErrorMessage{Message: err.Error()}
		e.Marshal(client.writer.w)
		return nil
	} else if m.Type() == ErrorMessageType {
		m.Marshal(client.writer.w)
		return nil
	}

	if m.Type() == PolicyResultType {
		createPolicy := message.(CreatePolicyMessage)
		mTyped := m.(PolicyResultMessage)
		site.Policies[createPolicy.species] = Policy{Action: createPolicy.action, PolicyId: mTyped.policy}
	} else if m.Type() == OkMessageType {
		return nil
	} else {
		fmt.Println(m)
		return m
	}

	return nil
}
