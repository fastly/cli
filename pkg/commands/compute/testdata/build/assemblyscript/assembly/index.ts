//! Default Compute@Edge template program.
import { Request, Response, Headers, URL, Fastly } from "@fastly/as-compute";

// The entry point for your application.
//
// Use this function to define your main request handling logic. It could be
// used to route based on the request properties (such as method or path), send
// the request to a backend, make completely new requests, and/or generate
// synthetic responses.

function main(req: Request): Response {
  // Filter requests that have unexpected methods.
  if (!["HEAD", "GET"].includes(req.method)) {
    return new Response(String.UTF8.encode("This method is not allowed"), {
      status: 405,
      headers: null,
      url: null
    });
  }

  let url = new URL(req.url);

  // If request is to the `/` path...
  if (url.pathname == "/") {
    // Below are some common patterns for Compute@Edge services using AssemblyScript.
    // Head to https://developer.fastly.com/learning/compute/assemblyscript/ to discover more.

    // Create a new request.
    // let bereq = new Request("http://example.com", {
    //     method: "GET",
    //     headers: null,
    //     body: null
    // });

    // Add request headers.
    // req.headers.set("X-Custom-Header", "Welcome to Compute@Edge!");
    // req.headers.set(
    //   "X-Another-Custom-Header",
    //   "Recommended reading: https://developer.fastly.com/learning/compute"
    // );

    // Create a cache override.
    // let cacheOverride = new Fastly.CacheOverride();
    // cacheOverride.setTTL(60);

    // Forward the request to a backend.
    // let beresp = Fastly.fetch(req, {
    //     backend: "backend_name",
    //     cacheOverride,
    // }).wait();

    // Remove response headers.
    // beresp.headers.delete("X-Another-Custom-Header");

    // Log to a Fastly endpoint.
    // const logger = Fastly.getLogEndpoint("my_endpoint");
    // logger.log("Hello from the edge!");

    // Send a default synthetic response.
    let headers = new Headers();
    headers.set('Content-Type', 'text/html; charset=utf-8');

    return new Response(String.UTF8.encode("<iframe src='https://developer.fastly.com/compute-welcome' style='border:0; position: absolute; top: 0; left: 0; width: 100%; height: 100%'></iframe>\n"), {
      status: 200,
      headers,
      url: null
    });
  }

  // Catch all other requests and return a 404.
  return new Response(String.UTF8.encode("The page you requested could not be found"), {
    status: 404,
    headers: null,
    url: null
  });
}

// Get the request from the client.
let req = Fastly.getClientRequest();

// Pass the request to the main request handler function.
let resp = main(req);

// Send the response back to the client.
Fastly.respondWith(resp);
