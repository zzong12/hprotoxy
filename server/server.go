package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"

	"github.com/zzong12/hprotoxy/codec"
	"github.com/zzong12/hprotoxy/loader"
	"github.com/zzong12/hprotoxy/log"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
)

const (
	HEADER_REQ_CODEC = "ReqCodec"
	HEADER_RES_CODEC = "ResCodec"
)

type (
	Config struct {
		ImportPath     string
		LoadFolder     string
		ReloadInterval uint16
		ProxyPort      uint16
		ManagerPort    uint16
	}

	Server struct {
		ProxyPort   uint16
		ManagerPort uint16
	}

	MetaItem struct {
		FileName string `json:"fileName"`
		MsgName  string `json:"msgName"`
		MsgType  string `json:"msgType"`
		Example  string `json:"example"`
	}
)

func NewServer(cfg Config) *Server {
	loader.InitLoader(cfg.ImportPath, cfg.LoadFolder, cfg.ReloadInterval)
	return &Server{
		ProxyPort:   cfg.ProxyPort,
		ManagerPort: cfg.ManagerPort,
	}
}

func writeErrorResponse(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("HProtoxy was unable to successfully proxy the request. See logs for details.\nerror: " + err.Error()))
}

func replaceReqBody(r *http.Request, cs codec.Codecs) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body.Close()

	data, err := cs.EncodeAll(body)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(data)
	r.Body = ioutil.NopCloser(buffer)
	r.ContentLength = int64(buffer.Len())
	return nil
}

func (s *Server) paresrCodecs(r *http.Request) (codec.Codecs, codec.Codecs, error) {
	reqCodec := r.Header.Get(HEADER_REQ_CODEC)
	if reqCodec == "" {
		return nil, nil, fmt.Errorf("request code is empty")
	}

	reqCodecs, err := codec.ParserCodes(reqCodec)
	if err != nil {
		return nil, nil, fmt.Errorf("request code is invalid: %v", err)
	}
	var resCodecs codec.Codecs
	resCode := r.Header.Get(HEADER_RES_CODEC)
	if resCode == "" { // default use request code as response code
		resCodecs = reqCodecs.Inverted()
	} else {
		resCodecs, err = codec.ParserCodes(resCode)
		if err != nil {
			return nil, nil, fmt.Errorf("response code is invalid: %v", err)
		}
	}
	return reqCodecs, resCodecs, nil

}

func (s *Server) proxyRequest(w http.ResponseWriter, r *http.Request) {
	reqCodes, resCodes, err := s.paresrCodecs(r)
	if err != nil {
		log.Log.WithError(err).Error("unable to parse codes")
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = replaceReqBody(r, reqCodes)
	if err != nil {
		log.Log.WithError(err).Error("error converting JSON body to proto")
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Override content-type to remove params
	// r.Header.Set("Content-Type", "application/x-protobuf")

	modifyResponse := func(r *http.Response) error {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("Failed to read response body: %v", err)
		}
		err = r.Body.Close()
		if err != nil {
			return fmt.Errorf("Error closing body: %v", err)
		}

		data, err := resCodes.DecodeAll(body)
		if err != nil {
			return fmt.Errorf("Failed to decode response body: %v", err)
		}

		buf := bytes.NewBuffer(data)
		r.Body = ioutil.NopCloser(buf)
		r.ContentLength = int64(buf.Len())
		r.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
		r.Header.Set("Content-Type", "application/json")
		return nil
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Log.WithError(err).Error("unable to proxy response from server")
		writeErrorResponse(w, http.StatusBadRequest, err)
	}

	proxy := &httputil.ReverseProxy{
		Director:       func(*http.Request) {},
		ModifyResponse: modifyResponse,
		ErrorHandler:   errorHandler,
	}

	proxy.ServeHTTP(w, r)
}

func (s *Server) apiMeta(w http.ResponseWriter, r *http.Request) {
	var res []*MetaItem
	for _, fd := range loader.GetLocalLoader().ListFileDescriptor() {
		for _, v := range fd.GetMessageTypes() {
			zeroV, _ := dynamic.NewMessage(v).MarshalJSONPB(&jsonpb.Marshaler{
				OrigName:     true,
				EnumsAsInts:  true,
				EmitDefaults: true,
			})
			res = append(res, &MetaItem{
				FileName: fd.GetName(),
				MsgName:  v.GetFullyQualifiedName(),
				MsgType:  "message",
				Example:  string(zeroV),
			})
		}
		for _, v := range fd.GetEnumTypes() {
			res = append(res, &MetaItem{
				FileName: fd.GetName(),
				MsgName:  v.GetFullyQualifiedName(),
				MsgType:  "enum",
				Example:  v.String(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) apiReload(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]string)
	if err := loader.GetLocalLoader().Load(); err != nil {
		res["status"] = "error"
		res["error"] = err.Error()
	} else {
		res["status"] = "ok"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) apiUpload(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]string)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		res["status"] = "error"
		res["error"] = err.Error()
	} else {
		for _, fheaders := range r.MultipartForm.File {
			for _, hdr := range fheaders {
				if !strings.HasSuffix(hdr.Filename, ".proto") {
					res["status"] = "error"
					if res["error"] == "" {
						res["error"] = "only .proto files are allowed, bug got "
					}
					res["error"] += fmt.Sprintf(" %s", hdr.Filename)
					continue
				}
				file, err := hdr.Open()
				if err != nil {
					res["status"] = "error"
					res["error"] = err.Error()
					break
				}
				if err := loader.GetLocalLoader().AddFile(hdr.Filename, file); err != nil {
					res["status"] = "error"
					res["error"] = err.Error()
					break
				}
			}
		}
		if _, ok := res["status"]; !ok {
			res["status"] = "ok"
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) apiDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res := make(map[string]string)
	if r.Method != "DELETE" {
		res["status"] = "error"
		res["error"] = "only DELETE method is allowed"
	} else {
		fileName := r.URL.Query().Get("file")
		err := loader.GetLocalLoader().DelFile(fileName)
		if err != nil {
			res["status"] = "error"
			res["error"] = err.Error()
		} else {
			res["status"] = "ok"
		}
	}
	json.NewEncoder(w).Encode(res)
}

func (s *Server) apiRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res := make(map[string]string)
	if r.Method != "GET" {
		res["status"] = "error"
		res["error"] = "only GET method is allowed"
	} else {
		fileName := r.URL.Query().Get("file")
		file, err := loader.GetLocalLoader().ReadFile(fileName)
		if err != nil {
			res["status"] = "error"
			res["error"] = err.Error()
		} else {
			res["status"] = "ok"
			res["file"] = file
		}
	}
	json.NewEncoder(w).Encode(res)
}

func (s *Server) webPages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// for test html page
	// if html, err := ioutil.ReadFile(r.URL.Path[1:]); err != nil {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// } else {
	// 	w.Write(html)
	// }

	if html, ok := WebPages[r.URL.Path]; ok {
		w.Write([]byte(html))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// Run starts the proxy server.
func (s *Server) Run() {
	loader.GetLocalLoader().Start()

	var wg sync.WaitGroup
	wg.Add(2)

	// proxy server
	go func() {
		proxySvrMux := http.NewServeMux()
		proxySvrMux.HandleFunc("/", s.proxyRequest)
		proxySvr := http.Server{
			Addr:    fmt.Sprintf(":%d", s.ProxyPort),
			Handler: proxySvrMux,
		}
		log.Log.Infof("Proxy server started on port %d", s.ProxyPort)
		wg.Done()
		proxySvr.ListenAndServe()
	}()

	// manager server
	go func() {
		managerSvrMux := http.NewServeMux()
		managerSvrMux.HandleFunc("/st/meta", s.apiMeta)
		managerSvrMux.HandleFunc("/do/reload", s.apiReload)
		managerSvrMux.HandleFunc("/do/upload", s.apiUpload)
		managerSvrMux.HandleFunc("/do/delete", s.apiDelete)
		managerSvrMux.HandleFunc("/do/read", s.apiRead)
		managerSvrMux.HandleFunc("/", s.webPages)
		managerSvr := http.Server{
			Addr:    fmt.Sprintf(":%d", s.ManagerPort),
			Handler: managerSvrMux,
		}
		log.Log.Infof("Manager server started on port %d", s.ManagerPort)
		wg.Done()
		managerSvr.ListenAndServe()
	}()

	wg.Wait()
	// wait for interrupt signal to gracefully shutdown the server with
	log.Log.Info("Press Ctrl+C to stop the server")
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
	log.Log.Info("Shutting down server...")
}
