package server

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/zzong12/hprotoxy/log"
)

type ProtoDescriptorLoader struct {
	importPath     string
	loadFolder     string
	lock           *sync.RWMutex
	lastLoadTIme   time.Time
	parser         *protoparse.Parser
	fileDesc       []*desc.FileDescriptor
	enumDescMap    map[string]*desc.EnumDescriptor
	messageDescMap map[string]*desc.MessageDescriptor
}

func (p *ProtoDescriptorLoader) GetMessageDescriptor(name string) (*desc.MessageDescriptor, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if desc, ok := p.messageDescMap[name]; ok {
		return desc, nil
	}
	return nil, fmt.Errorf("message descriptor not found: %s", name)
}

func (p *ProtoDescriptorLoader) GetEnumDescriptor(name string) (*desc.EnumDescriptor, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if desc, ok := p.enumDescMap[name]; ok {
		return desc, nil
	}
	return nil, fmt.Errorf("enum descriptor not found: %s", name)
}

func (p *ProtoDescriptorLoader) Start() error {
	err := p.Load()
	if err != nil {
		return err
	}
	go func() {
		t := time.NewTimer(300 * time.Second) // 5 min
		for {
			select {
			case <-t.C:
				err := p.Load()
				if err != nil {
					log.Log.WithError(err).Error("error loading proto files")
				}
			}
		}
	}()
	return nil
}

func (p *ProtoDescriptorLoader) Load() error {
	var pfs []string
	fullLoadFolder := path.Join(p.importPath, p.loadFolder)
	preImportPathLen := len(p.importPath) + 1
	err := WalkDepth(fullLoadFolder, 10, func(dir, name string, isDir bool) error {
		if !isDir && strings.HasSuffix(name, ".proto") {
			relativePath := path.Join(dir, name)[preImportPathLen:]
			pfs = append(pfs, relativePath)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(pfs) == 0 {
		return fmt.Errorf("no proto files found")
	}

	fileDesc, err := p.parser.ParseFiles(pfs...)
	if err != nil {
		return err
	}

	p.fileDesc = fileDesc

	p.lock.Lock()
	defer p.lock.Unlock()

	var keys []string
	for _, fd := range fileDesc {
		for _, v := range fd.GetMessageTypes() {
			p.messageDescMap[v.GetFullyQualifiedName()] = v
		}
		for _, v := range fd.GetEnumTypes() {
			p.enumDescMap[v.GetFullyQualifiedName()] = v
		}
	}
	p.lastLoadTIme = time.Now()

	log.Log.Info("loaded proto", " ,files=", pfs, " ,msgs=", keys)
	return nil
}