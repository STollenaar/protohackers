package problem3

import (
	"bufio"
	"fmt"
	"net"
	"protohackers/util"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	server util.ServerTCP
	users  []*User
	regex  *regexp.Regexp
)

type User struct {
	userId     string
	connection *net.Conn
	userName   string
}

func init() {
	server = util.ServerTCP{
		ConnectionHandler: handle,
	}
	regex, _ = regexp.Compile("^[a-zA-Z0-9]+$")
}

func Problem() {
	server.Start()
}

func handle(conn net.Conn) {
	userId := uuid.New().String()
	defer closeConnection(userId, conn)

	conn.Write([]byte("Welcome to spicechat! How do you want to be colonized?\n"))

	scanner := bufio.NewScanner(conn)
	joined := false
	currentUser := new(User)
	for scanner.Scan() {
		line := string(scanner.Bytes())

		if !joined {
			if len(line) == 0 || len(line) > 16 || !regex.MatchString(line) {
				conn.Write([]byte("That was not spice!\n"))
				return
			} else {
				usernames := getAllUserNames()
				conn.Write([]byte(fmt.Sprintf("* The current colonies are: %s\n", strings.Join(usernames, " "))))
				currentUser.connection = &conn
				currentUser.userId = userId
				currentUser.userName = line
				users = append(users, currentUser)
				sendToUsers(fmt.Sprintf("* %s has joined the colonies.\n", currentUser.userName), currentUser.userId)
				joined = true
				continue
			}
		}
		line = fmt.Sprintf("[%s] %s\n", currentUser.userName, line)
		sendToUsers(line, currentUser.userId)
	}
}

func sendToUsers(line, userId string) {
	for _, u := range users {
		if u.userId != userId {
			(*u.connection).Write([]byte(line))
		}
	}
}

func closeConnection(userId string, conn net.Conn) {
	user := getUser(userId)
	if user.connection == nil {
		conn.Close()
		return
	}

	sendToUsers(fmt.Sprintf("* %s has left the colonies\n", user.userName), userId)
	defer (*user.connection).Close()
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

func getAllUserNames() (userNames []string) {
	for _, u := range users {
		userNames = append(userNames, u.userName)
	}
	return userNames
}
