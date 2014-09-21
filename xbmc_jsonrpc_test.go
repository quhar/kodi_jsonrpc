package xbmc_jsonrpc

import "fmt"

func ExampleNew() {
	xbmc, err := New(`localhost:9090`, 15) // timeout after 15 secs
	defer xbmc.Close()                     // always close to free resources

	if err != nil {
		panic(fmt.Sprintf(`Couldn't connect to XBMC: %v`, err))
	}

	request := Request{Method: `JSONRPC.Version`}
	response := xbmc.Send(request, true) // second param says we need a response

	// wait indefinitely for response (timeout 0)
	result, err := response.Read(0)

	if err != nil {
		panic(fmt.Sprintf(`XBMC responded with error: %v`, err))
	}

	fmt.Println(result)

	// Output:
	// map[version:map[major:6 minor:14 patch:3]]
}
