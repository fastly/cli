// The entry point for your application.
//
// Use this fetch event listener to define your main request handling logic. It could be
// used to route based on the request properties (such as method or path), send
// the request to a backend, make completely new requests, and/or generate
// synthetic responses.
addEventListener('fetch', async function handleRequest(event) {

  // NOTE: By default, console messages are sent to stdout (and stderr for `console.error`).
  // To send them to a logging endpoint instead, use `console.setEndpoint:
  // console.setEndpoint("my-logging-endpoint");

  // Get the client request from the event
  let req = event.request;

  // Make any desired changes to the client request.
  req.headers.set("Host", "example.com");

  // We can filter requests that have unexpected methods.
  const VALID_METHODS = ["GET"];
  if (!VALID_METHODS.includes(req.method)) {
    let response = new Response("This method is not allowed", {
      status: 405
    });
    // Send the response back to the client.
    event.respondWith(response);
    return;
  }

  let method = req.method;
  let url = new URL(event.request.url);

  // If request is a `GET` to the `/` path, send a default response.
  if (method == "GET" && url.pathname == "/") {
    let headers = new Headers();
    headers.set('Content-Type', 'text/html; charset=utf-8');
    let response = new Response("<iframe src='https://fastly.com/documentation/help/compute-welcome' style='border:0; position: absolute; top: 0; left: 0; width: 100%; height: 100%'></iframe>\n", {
      status: 200,
      headers
    });
    // Send the response back to the client.
    event.respondWith(response);
    return;
  }

  // Catch all other requests and return a 404.
  let response = new Response("The page you requested could not be found", {
    status: 404
  });
  // Send the response back to the client.
  event.respondWith(response);
});
