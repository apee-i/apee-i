package ws

import (
	"fmt"
	"time"

	"github.com/IbraheemHaseeb7/apee-i/cmd"
	"github.com/Jeffail/gabs/v2"
	"github.com/gorilla/websocket"
)

// Strategy is used to make selection for the strategy
type Strategy struct{}

// Hit acts as a socket client and interacts with the HOST
func (s *Strategy) Hit(fileContents *cmd.Structure, structure cmd.PipelineBody) (cmd.APIResponse, error) {

	// Define the WebSocket server URL
	url := structure.BaseURL

	// Establish a connection to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return cmd.APIResponse{}, err
	}
	defer conn.Close() // Ensure the connection is closed when done

	// Send a message to the WebSocket server
	message := []byte(structure.Body.(string))
	err = conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return cmd.APIResponse{}, err
	}
	fmt.Printf("Sent: %s\n", message)

	// Set a read deadline for receiving messages
	conn.SetReadDeadline(time.Now().Add(time.Duration(structure.Timeout) * time.Second))

	// Read messages from the server
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return cmd.APIResponse{}, err
			}
			break
		}
		fmt.Printf("Received: %s\n", msg)
	}

	body, err := gabs.ParseJSON([]byte(`{"message": "successfully closed the connection to websocket"}`))
	if err != nil {
		return cmd.APIResponse{}, err
	}

	return cmd.APIResponse{
		Body:       body,
		StatusCode: 200,
	}, nil
}
