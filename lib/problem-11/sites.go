package problem11

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type Site struct {
	PopulationTarget PopulationTarget  `json:"populationTargets"`
	ID               uint32            `json:"id"`
	
	siteVisits       []PopulationVisit
	policies  map[string]Policy
	svChan    chan SiteVisitMessage
	authority *Client
}

type SiteMap struct {
	mu   sync.Mutex
	Site map[uint32]*Site `json:"sites"`
}

func init() {
	if _, err := os.Stat("sites.json"); err == nil {
		file, _ := os.ReadFile("sites.json")

		err := json.Unmarshal(file, &sites)
		if err != nil {
			log.Fatal(err)
		}
		sites.mu.Lock()
		for _, s := range sites.Site {
			s.svChan = make(chan SiteVisitMessage)
			go func(site *Site) {
				for {
					site.handleSiteVisit(<-site.svChan)
				}
			}(s)
		}

		sites.mu.Unlock()
	}
}

func createSite(site uint32) *Site {
	s := &Site{
		svChan:   make(chan SiteVisitMessage),
		policies: make(map[string]Policy),
		ID:       site,
	}
	s.dialAuthority()

	sites.Site[site] = s

	go func(ss *Site) {
		for {
			ss.handleSiteVisit(<-ss.svChan)
		}
	}(s)

	return s
}

func (site *Site) dialAuthority() *Client {
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
	site.authority = client

	HelloMessage{Protocol: "pestcontrol", Version: 1}.Marshal(client.writer.w)
	m, err := server.ReadMessage(client)
	if err != nil {
		fmt.Printf("Error reading helloMessage, %s\n", err)
		if err == io.EOF {
			return nil
		}
		e := ErrorMessage{Message: err.Error()}
		e.Marshal(client.writer.w)
		return nil
	} else if m.Type() == ErrorMessageType {
		fmt.Printf("Error returned helloMessage, %s\n", err)
		m.Marshal(client.writer.w)
		return nil
	} else if m.Type() != HelloMessageType {
		e := ErrorMessage{Message: err.Error()}
		e.Marshal(client.writer.w)
		return nil
	}

	DialAuthorityMessage{site: site.ID}.Marshal(client.writer.w)
	m, err = server.ReadMessage(client)
	if err != nil {
		fmt.Printf("Error reading dial message, %s, send with site %d\n", err, site.ID)
		if err == io.EOF {
			return nil
		}
		e := ErrorMessage{Message: err.Error()}
		e.Marshal(client.writer.w)
		return nil
	} else if m.Type() == ErrorMessageType {
		fmt.Printf("Error return dial message, %s, send with site %d\n", err, site.ID)
		m.Marshal(client.writer.w)
		return nil
	}
	mTyped := m.(TargetPopulationMessage)
	site.PopulationTarget = mTyped.Populations
	return client
}

func (site *Site) sendAuthorityMessage(message Message) Message {
	client := site.authority

	if client == nil {
		client = site.dialAuthority()
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
		site.policies[createPolicy.species] = Policy{Action: createPolicy.action, PolicyId: mTyped.policy}
	} else if m.Type() == OkMessageType {
		return nil
	} else {
		fmt.Println(m)
		return m
	}

	return nil
}


func (site *Site) handleSiteVisit(s SiteVisitMessage) {
	// fmt.Println(s)
	for species, count := range site.enhanceVisits(s.populations) {
		target, ok := site.PopulationTarget[species]
		if !ok {
			continue
		}
		pv, ok := site.policies[species]
		if count < target.Min || (count == 0 && target.Min == 0) {
			// Conserve
			if ok && pv.Action == CullAction {
				// Delete
				delete(site.policies, species)
				site.sendAuthorityMessage(DeletePolicyMessage{policy: pv.PolicyId})
				site.sendAuthorityMessage(CreatePolicyMessage{species: species, action: ConserveAction})
			} else if !ok {
				// Create Policy
				site.sendAuthorityMessage(CreatePolicyMessage{species: species, action: ConserveAction})
			}
		} else if count > target.Max {
			// Cull
			if ok && pv.Action == ConserveAction {
				// Delete
				delete(site.policies, species)
				site.sendAuthorityMessage(DeletePolicyMessage{policy: pv.PolicyId})
				site.sendAuthorityMessage(CreatePolicyMessage{species: species, action: CullAction})
			} else if !ok {
				// Create Policy
				site.sendAuthorityMessage(CreatePolicyMessage{species: species, action: CullAction})
			}
		} else if ok {
			// Remove Policy
			delete(site.policies, species)
			site.sendAuthorityMessage(DeletePolicyMessage{policy: pv.PolicyId})
		}
	}
}

func (site *Site)enhanceVisits(sv PopulationVisit) (enhanced PopulationVisit) {
	enhanced = sv
	for species := range site.PopulationTarget {
		if _, ok := enhanced[species]; !ok{
			enhanced[species] = 0
		}
	}
	return enhanced
}
