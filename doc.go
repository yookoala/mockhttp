// Copyright 2018 Yeung Shu Hung (Koala Yeung).
// This software is licensed with the MIT license.
// You may obtain a copy of the license in this
// repository.

/*

Package mockhttp helps you to mock network behaviour for testing against
http stack.

Basic Usage

To mock the transport for whatever test you want to run with "http" library,
you may override the response with some of the utils like
mockhttp.StaticResponseRT:

	mocktransport := mockhttp.StaticResponseRT("hello world", "text/plain")
	mockhttp.UseTransport(mocktransport)

	...

	// resp.Body will always be "hello world", no matter the URL or method
	resp, err := http.Get("http://whatever.com")
	...

	mochhttp.RestoreTransport()

	// resp.Body will be normal
	resp, err := http.Get("http://whatever.com")

	...

Override Multiple Sites Behaviour

You may mux different response for different URL host name.

	mux := mockhttp.MuxRoundTripper{}
	mux.AddFunc("www.google.com",
		mockhttp.StaticResponseRT("fake google", "text/plain"))
	mux.AddFunc("www.facebook.com",
		mockhttp.StaticResponseRT("fake facebook", "text/plain"))

	// "*" for setting fallback http.RoundTripper
	mux.AddFunc("*", mockhttp.TransportErrorRT(fmt.Errorf("no connection to host")))

	mockhttp.UseTransport(mocktransport)

	...

	// resp.Body will be "fake google"
	resp, err := http.Get("http://www.google.com/helloAPI")

	// resp.Body will be "fake facebook"
	resp, err := http.Get("http://www.facebook.com/helloAPI")

	// will always return error: "no connection to host"
	resp, err := http.Get("http://www.archive.org/helloAPI")

	...

	mochhttp.RestoreTransport()

Use Without Overriding DefaultTransport

MuxRoundTripper also can be used directly to create client.

	mux := mockhttp.MuxRoundTripper{}
	...
	mux.GetClient().Post("http://foobar.com", strings.NewReader("some+data"))

Partial Override

You can partially override the round trip like this:

	mux := mockhttp.MuxRoundTripper{}
	mux.AddFunc("www.google.com",
		mockhttp.StaticResponseRT("fake google", "text/plain"))
	mux.AddFunc("www.facebook.com",
		mockhttp.StaticResponseRT("fake facebook", "text/plain"))

	// "*" for setting fallback http.RoundTripper
	mux.Add("*", http.DefaultTransport)

	client := mux.GetClient()
	...

*/
package mockhttp
