from fastly_compute.wsgi import WsgiHttpIncoming


def app(environ, start_response):
    status = "200 OK"
    headers = [("Content-type", "text/plain; charset=utf-8")]
    start_response(status, headers)
    return [b"Hello, world!"]


HttpIncoming = WsgiHttpIncoming(app)
