package discord

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Microsoft/go-winio"
)

type Client struct {
	appID string
	conn  net.Conn
}

type Assets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
	SmallImage string `json:"small_image,omitempty"`
	SmallText  string `json:"small_text,omitempty"`
}

type Timestamps struct {
	Start int64 `json:"start,omitempty"` // Unix ms
	End   int64 `json:"end,omitempty"`   // Unix ms
}

type Activity struct {
	Details    string      `json:"details,omitempty"`
	State      string      `json:"state,omitempty"`
	Type       int         `json:"type"`
	Timestamps *Timestamps `json:"timestamps,omitempty"`
	Assets     *Assets     `json:"assets,omitempty"`
}

func NewClient(appID string) *Client {
	return &Client{appID: appID}
}

func (c *Client) Connect() error {
	var err error

	for i := 0; i < 10; i++ {
		pipePath := fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i)
		c.conn, err = winio.DialPipe(pipePath, nil)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("could not connect to discord: %v", err)
	}

	payload, _ := json.Marshal(map[string]string{"v": "1", "client_id": c.appID})
	return c.Send(0, payload)
}

func (c *Client) SetActivity(activity Activity) error {
	payload := map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid":      os.Getpid(),
			"activity": activity,
		},
		"nonce": fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	data, _ := json.Marshal(payload)
	return c.Send(1, data)
}

func (c *Client) Send(opcode int32, payload []byte) error {
	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], uint32(opcode))
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(payload)))

	_, err := c.conn.Write(append(header, payload...))
	return err
}
