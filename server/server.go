package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"

	"github.com/zzong12/hprotoxy/log"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
)

const (
	ContextReqMsg = "reqmsg"
	ContextResMsg = "resmsg"
)

type (
	Config struct {
		ImportPath  string
		LoadFolder  string
		ProxyPort   uint16
		ManagerPort uint16
	}

	Server struct {
		ProxyPort   uint16
		ManagerPort uint16
		Loader      *ProtoDescriptorLoader
	}

	protoTypes struct {
		reqMsg  string
		respMsg string
	}

	MetaItem struct {
		FileName string `json:"fileName"`
		MsgName  string `json:"msgName"`
		MsgType  string `json:"msgType"`
		Example  string `json:"example"`
	}
)

func NewServer(cfg Config) *Server {
	loader := &ProtoDescriptorLoader{
		importPath: cfg.ImportPath,
		loadFolder: cfg.LoadFolder,
		parser: &protoparse.Parser{
			ImportPaths: []string{cfg.ImportPath},
		},
		lock:           &sync.RWMutex{},
		fileDesc:       make([]*desc.FileDescriptor, 0),
		enumDescMap:    make(map[string]*desc.EnumDescriptor),
		messageDescMap: make(map[string]*desc.MessageDescriptor),
	}
	return &Server{
		ProxyPort:   cfg.ProxyPort,
		ManagerPort: cfg.ManagerPort,
		Loader:      loader,
	}
}

func parseMessageTypes(r *http.Request) (ptypes protoTypes, err error) {
	ctype := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "application/json") {
		err = fmt.Errorf("Content-Type is not application/json")
		return
	}
	_, params, err := mime.ParseMediaType(ctype)
	if err != nil {
		return ptypes, err
	}
	return protoTypes{
		reqMsg:  params[ContextReqMsg],
		respMsg: params[ContextResMsg],
	}, nil
}

func writeErrorResponse(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("HProtoxy was unable to successfully proxy the request. See logs for details."))
}

func replaceReqBody(r *http.Request, msgDescriptor *desc.MessageDescriptor) error {
	msg := dynamic.NewMessage(msgDescriptor)
	err := jsonpb.Unmarshal(r.Body, msg)
	if err != nil {
		log.Log.WithError(err).Error("unable to unmarshal into json")
		return fmt.Errorf("Unable to unmarshal into json: %v", err)
	}

	reqBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Log.WithError(err).Error("unable to marshal message")
		return fmt.Errorf("Unable to marshal message: %v", err)
	}

	// replace request body with protobuf bytes
	buffer := bytes.NewBuffer(reqBytes)
	r.Body = ioutil.NopCloser(buffer)
	r.ContentLength = int64(buffer.Len())

	return nil
}

func (s *Server) findMessageDescriptors(reqMsg string, respMsg string) (reqMsgDesc *desc.MessageDescriptor, respMsgDescs *desc.MessageDescriptor, err error) {
	reqMsgDesc, err = s.Loader.GetMessageDescriptor(reqMsg)
	if err != nil {
		return
	}
	respMsgDescs, err = s.Loader.GetMessageDescriptor(respMsg)
	if err != nil {
		return
	}
	return
}

func (s *Server) proxyRequest(w http.ResponseWriter, r *http.Request) {
	msgTypes, err := parseMessageTypes(r)
	if err != nil {
		log.Log.WithError(err).Error("error parsing message types")
		writeErrorResponse(w, http.StatusBadRequest)
		return
	}

	reqMsgDesc, respMsgDesc, err := s.findMessageDescriptors(msgTypes.reqMsg, msgTypes.respMsg)
	if err != nil {
		log.Log.WithError(err).Error("error finding message descriptors")
		writeErrorResponse(w, http.StatusBadRequest)
		return
	}

	if reqMsgDesc != nil {
		if err = replaceReqBody(r, reqMsgDesc); err != nil {
			log.Log.WithError(err).Error("error converting JSON body to proto")
			writeErrorResponse(w, http.StatusBadRequest)
			return
		}
	}

	// Override content-type to remove params
	r.Header.Set("Content-Type", "application/x-protobuf")

	modifyResp := func(r *http.Response) error {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("Failed to read response body: %v", err)
		}
		err = r.Body.Close()
		if err != nil {
			return fmt.Errorf("Error closing body: %v", err)
		}

		// Try all possible responses until something works
		msg := dynamic.NewMessage(respMsgDesc)
		err = proto.Unmarshal(body, msg)
		if err != nil {
			log.Log.WithError(err).Error("error unmarshaling response")
			return err
		}

		marshaler := jsonpb.Marshaler{
			EmitDefaults: true,
		}
		buf := bytes.NewBuffer(nil)
		err = marshaler.Marshal(buf, msg)
		if err != nil {
			return fmt.Errorf("Failed to marshal response: %v", err)
		}
		r.Body = ioutil.NopCloser(buf)
		r.ContentLength = int64(buf.Len())
		r.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
		r.Header.Set("Content-Type", "application/json")
		return nil
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Log.WithError(err).Error("unable to proxy response from server")
		writeErrorResponse(w, http.StatusBadRequest)
	}

	proxy := &httputil.ReverseProxy{
		Director:       func(*http.Request) {},
		ModifyResponse: modifyResp,
		ErrorHandler:   errorHandler,
	}

	proxy.ServeHTTP(w, r)
}

func (s *Server) apiMeta(w http.ResponseWriter, r *http.Request) {
	var res []*MetaItem
	for _, fd := range s.Loader.fileDesc {
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
	if err := s.Loader.Load(); err != nil {
		res["status"] = "error"
		res["error"] = err.Error()
	} else {
		res["status"] = "ok"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Run starts the proxy server.
func (s *Server) Run() {
	s.Loader.Start()

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
		// managerSvrMux.HandleFunc("/do/upload", s.apiUpload)
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
