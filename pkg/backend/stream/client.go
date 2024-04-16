package stream

import (
	"bufio"
	"context"

	"io"

	"encoding/json"
	"strings"

	"time"

	"github.com/mike1808/ax/pkg/backend/common"
	"github.com/mike1808/ax/pkg/heuristic"
)

type Client struct {
	reader io.Reader
	msgKey string
}

func New(file io.Reader, msgKey string) *Client {
	if msgKey == "" {
		msgKey = "message"
	}
	return &Client{file, msgKey}
}

func (client *Client) parseLine(line string) common.LogMessage {
	decoder := json.NewDecoder(strings.NewReader(line))
	obj := make(map[string]interface{})
	err := decoder.Decode(&obj)
	if err != nil {
		obj[client.msgKey] = strings.TrimSpace(line)
		return common.LogMessage{
			Timestamp:  time.Now(),
			Attributes: obj,
			MessageKey: client.msgKey,
		}
	}
	return common.LogMessage{
		Timestamp:  time.Now(), // TODO: Fix this
		Attributes: obj,
		MessageKey: client.msgKey,
	}
}

func (client *Client) ImplementsAdvancedFilters() bool {
	return true
}

func (client *Client) Query(ctx context.Context, q common.Query) <-chan common.LogMessage {
	resultChan := make(chan common.LogMessage)
	reader := bufio.NewReader(client.reader)
	go func() {
		var ltFunc heuristic.LogTimestampParser
	LFor:
		for {
			select {
			case <-ctx.Done():
				// Context canceled, let's get outta here
				break LFor
			default:
			}
			line, err := reader.ReadString('\n')
			if err != nil {
				// TODO: erro != EOF break?
				if err == io.EOF {
					break
				}
				//fmt.Println("Error: ", err)
				break
			}
			message := client.parseLine(line)
			if ltFunc == nil {
				ltFunc = heuristic.FindTimestampFunc(message)
			}
			if ltFunc != nil {
				ts := ltFunc(message)
				if ts != nil {
					message.Timestamp = *ts
				} else {
					ltFunc = heuristic.FindTimestampFunc(message)
					if ltFunc != nil {
						ts := ltFunc(message)
						message.Timestamp = *ts
					}
				}
			}
			if common.MatchesQuery(message, q) {
				message.Attributes = common.Project(message.Attributes, q.SelectFields)
				resultChan <- message
			}
		}
		close(resultChan)
	}()

	return resultChan
}

var _ common.Client = &Client{}
