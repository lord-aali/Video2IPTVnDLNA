package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"time"
	"unsafe"

	anacrolixlog "github.com/anacrolix/log"
	"github.com/anacrolix/dms/dlna/dms"
)

func startServer(absFolder string, port int, recursive bool, friendlyName string) error {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	dlnaServer := &dms.Server{
		HTTPConn:       ln,
		RootObjectPath: absFolder,
		FriendlyName:   friendlyName,
		NoProbe:        true,
		NoTranscode:    true,
		IgnoreHidden:   true,
		AllowedIpNets:  allowAllIPNets(),
		Logger:         anacrolixlog.Default,
		NotifyInterval: 30 * time.Second,
	}
	if err := dlnaServer.Init(); err != nil {
		return fmt.Errorf("init DLNA server: %w", err)
	}

	mux, err := dmsServeMux(dlnaServer)
	if err != nil {
		return err
	}
	registerIPTVRoutes(mux, absFolder, recursive)

	log.Printf("serving videos from %s on http://localhost%s", absFolder, addr)
	log.Printf("IPTV playlist: http://localhost%s/playlist.m3u", addr)
	log.Printf("DLNA media server: %q (discoverable on local network)", friendlyName)
	return dlnaServer.Run()
}

func registerIPTVRoutes(mux *http.ServeMux, absFolder string, recursive bool) {
	mux.HandleFunc("/playlist.m3u", func(w http.ResponseWriter, r *http.Request) {
		servePlaylist(w, r, absFolder, recursive)
	})
	mux.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir(absFolder))))
}

func allowAllIPNets() []*net.IPNet {
	_, v4, _ := net.ParseCIDR("0.0.0.0/0")
	_, v6, _ := net.ParseCIDR("::/0")
	return []*net.IPNet{v4, v6}
}

// dmsServeMux returns the internal HTTP mux from an initialized dms.Server.
// The field is unexported, so reflection is used to register IPTV routes on the same port.
func dmsServeMux(srv *dms.Server) (*http.ServeMux, error) {
	field := reflect.ValueOf(srv).Elem().FieldByName("httpServeMux")
	if !field.IsValid() {
		return nil, fmt.Errorf("dms server mux not found")
	}
	mux := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(*http.ServeMux)
	if mux == nil {
		return nil, fmt.Errorf("dms server mux is nil")
	}
	return mux, nil
}
