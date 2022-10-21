package problem5

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"protohackers/util"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const (
	mobServer         = "chat.protohackers.com:16963"
	mobBitcoinAddress = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

var (
	serverChat          util.ServerTCP
	users               []*User
	bitcoinAddressRegex *regexp.Regexp
)

type User struct {
	userId             string
	userConnection     *net.Conn
	upstreamConnection *net.Conn
	userName           string
}

func init() {
	serverChat = util.ServerTCP{
		ConnectionHandler: chatHandler,
	}

	bitcoinAddressRegex, _ = regexp.Compile(`(^|\s{0,1})(\w{26,36})($|\s)`)

}

func Problem() {
	serverChat.Start()
}

func upstreamHandler(conn net.Conn, userId string) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		line = fmt.Sprintf("%s\n", line)
		sendToUsers(line, userId, true)
	}
}

func chatHandler(conn net.Conn) {
	upstreamCon, _ := net.Dial("tcp", mobServer)
	userId := uuid.New().String()

	defer closeConnection(userId, conn, upstreamCon)

	reader := bufio.NewReader(conn)

	currentUser := new(User)
	currentUser.userConnection = &conn
	currentUser.userId = userId

	joined := false
	currentUser.upstreamConnection = &upstreamCon
	users = append(users, currentUser)

	go upstreamHandler(upstreamCon, userId)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if !joined {
			currentUser.userName = line
			joined = true
		}
		sendToUpstream(line, currentUser)
	}
}

func swapBitcoinAddress(line string) string {
	matches := bitcoinAddressRegex.FindAllString(line, -1)

	if len(matches) > 0 {
		for _, match := range matches {
			match = strings.TrimSpace(match)
			if string(match[0]) != "7" {
				continue
			}
			line = strings.ReplaceAll(line, match, mobBitcoinAddress)
		}
	}
	return line
}

func sendToUpstream(line string, currentUser *User) {
	line = swapBitcoinAddress(line)
	_, err := (*currentUser.upstreamConnection).Write([]byte(line))
	if err != nil {
		log.Fatal(err)
	}
}

func sendToUsers(line, userId string, self bool) {
	line = swapBitcoinAddress(line)

	for _, u := range users {
		if u.userId != userId && !self {
			(*u.userConnection).Write([]byte(line))
		} else if u.userId == userId && self {
			(*u.userConnection).Write([]byte(line))
		}
	}
}

func closeConnection(userId string, userConn, upstreamCon net.Conn) {
	user := getUser(userId)
	if user.userConnection == nil {
		upstreamCon.Close()
		userConn.Close()
		return
	}

	defer (*user.userConnection).Close()
	defer (*user.upstreamConnection).Close()
	removeFromSlice(userId)
}

func removeFromSlice(userId string) {
	index := getUserIndex(userId)
	if index == -1 {
		return
	}
	remove(index)
}

func remove(s int) {
	users = append(users[:s], users[s+1:]...)
}

func getUserIndex(userId string) int {
	for i, u := range users {
		if u.userId == userId {
			return i
		}
	}
	return -1
}

func getUser(userId string) *User {
	for _, u := range users {
		if u.userId == userId {
			return u
		}
	}
	return &User{}
}
