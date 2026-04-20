package application

import "net/http"

func HTTPServerWithAddr(addr string) HTTPServerOption {
	return func(a *Application) {
		a.httpConfig.Server.Addr = addr
	}
}

func HTTPServerWithName(name string) HTTPServerOption {
	return func(a *Application) {
		a.httpConfig.Server.Name = name
	}
}

func HTTPServerWithHandler(handler http.Handler) HTTPServerOption {
	return func(a *Application) {
		a.HTTP = handler
	}
}
