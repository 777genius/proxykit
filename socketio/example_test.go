package socketio

import "fmt"

func Example() {
	namespace, event, argsJSON, ok := ParseEvent(`42/chat,17["message",{"body":"hello"}]`)

	fmt.Println(namespace)
	fmt.Println(event)
	fmt.Println(argsJSON)
	fmt.Println(ok)
	// Output:
	// /chat
	// message
	// ["message",{"body":"hello"}]
	// true
}
