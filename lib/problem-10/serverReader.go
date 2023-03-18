package problem10

import (
	"fmt"
	"protohackers/util"
	"strconv"
	"strings"
)

type ServerWithReader struct {
	util.ServerTCP
}

type Response struct {
	fmt.Stringer
	success bool
	message string
}

func (r *Response) String() string {
	success := "ERR"
	if r.success {
		success = "OK"
	}
	return fmt.Sprintf("%s %s", success, r.message)
}

type Request struct {
	cmd  string
	args []string
}

func (s *ServerWithReader) handleLine(request Request, client *Client) {
	if request.cmd == "" {
		return
	}
	fmt.Println(request)
	switch request.cmd {
	case "HELP":
		client.send("OK usage: HELP|GET|PUT|LIST")
	case "GET":
		if len(request.args) == 0 || len(request.args) > 2 {
			client.send("ERR usage: GET file [revision]")
			return
		} else if !strings.HasPrefix(request.args[0], "/") || !files.isLeggalName(request.args[0]) {
			client.send("ERR illegal file name")
			return
		}
		file := files.getFile(request.args[0])
		if file == nil {
			client.send("ERR no such file")
			return
		}

		revision := len(file.data)
		if revision == 0 {
			client.send("OK 0")
			return
		}
		revision--
		var err error
		if len(request.args) == 2 {
			request.args[1] = strings.TrimPrefix(request.args[1], "r")
			revision, err = strconv.Atoi(request.args[1])
			// Preventing off by 1 issues
			revision--
			if err != nil || revision < 0 || revision > len(file.data) {
				client.send("ERR no such revision")
				return
			}
		}
		client.send("OK %d", len(file.data[revision]))
		client.sendRaw(file.data[revision])
	case "LIST":
		if len(request.args) == 0 {
			client.send("ERR usage: LIST dir")
			return
		} else if !strings.HasPrefix(request.args[0], "/") || !files.isLeggalName(request.args[0]) {
			client.send("ERR invalid dir name")
			return
		}

		dir := files.listDir(request.args[0])

		client.send("OK %d", len(dir))
		for _, e := range dir {
			if len(e.data) > 0 {
				client.send("%s %s%d", e.name, "r", len(e.data))
			} else {
				client.send("%s/ DIR", e.name)
			}
		}
	case "PUT":
		if len(request.args) != 2 {
			client.send("ERR usage: PUT file length newline data")
			return
		} else if !strings.HasPrefix(request.args[0], "/") || !files.isLeggalName(request.args[0]) {
			client.send("ERR illegal file name")
			return
		}
		n, err := strconv.Atoi(request.args[1])

		var data []byte

		if err == nil {
			data = client.readLength(n)
			if !files.IsPrintable(data) {
				client.send("ERR illegal payload")
				return
			}
		}
		revision := files.putFile(request.args[0], data)
		client.send("OK r%d", revision)
	default:
		client.send("ERR illegal method: %s", request.cmd)
	}
}
