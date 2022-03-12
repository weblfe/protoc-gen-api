package app

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
		"io"
		"io/fs"
	"io/ioutil"
	"os"
	"sync"
)

type ProtocPlugin struct {
	input      fs.File
	output     io.Writer
	locker     sync.Locker
	generators map[string]Generator
}

type Generator interface {
	Name() string
	Generate(*protogen.Plugin, *protogen.File) (*protogen.GeneratedFile, error)
}

type Option func(*ProtocPlugin)

func ApplyInput(fs fs.File) Option {
	return func(plugin *ProtocPlugin) {
		plugin.input = fs
	}
}

func ApplyOutput(fs io.Writer) Option {
		return func(plugin *ProtocPlugin) {
				plugin.output = fs
		}
}

func ApplyGenerators(g ...Generator) Option {
	return func(plugin *ProtocPlugin) {
		plugin.Register(g...)
	}
}

func NewProtocPluginApp(options ...Option) *ProtocPlugin {
	var plugin = new(ProtocPlugin)
	plugin.init()
	plugin.Apply(options...)
	return plugin
}

func (p *ProtocPlugin) Apply(options ...Option) {
	for _, o := range options {
		o(p)
	}
}

func (p *ProtocPlugin) init() {
	p.locker = &sync.RWMutex{}
	p.generators = make(map[string]Generator, 0)
}

func (p *ProtocPlugin) Register(generators ...Generator) {
	p.locker.Lock()
	defer p.locker.Unlock()
	for _, g := range generators {
		if _, ok := p.generators[g.Name()]; ok {
			continue
		}
		p.generators[g.Name()] = g
	}
}

func (p *ProtocPlugin) Run() error {
	// 解析请求
	request, err := p.GetProtocRequest()
	if err != nil {
		return err
	}
	// 生成plugin
	plugin, err := p.GetProtoGenPlugin(request)
	if err != nil {
		return err
	}
	// 构建生成文件
	out, err := p.MakeFiles(plugin)
	if err != nil {
		return err
	}
	// 相应输出到stdout, 它将被 protoc 接收
	if _, err = fmt.Fprintf(p.GetOutput(), string(out)); err != nil {
		return err
	}
	return nil
}

func (p *ProtocPlugin)GetOutput() io.Writer {
		if p.output == nil {
				return os.Stdout
		}
		return p.output
}

func (p *ProtocPlugin) MakeFiles(plugin *protogen.Plugin) ([]byte, error) {

	for _, fd := range plugin.Files {
		if !fd.Generate {
			continue
		}
		// @TODO 多协程
		if err := p.generate(plugin, fd); err != nil {
			return nil, err
		}
	}
	out, err := proto.Marshal(plugin.Response())
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (p *ProtocPlugin) generate(plugin *protogen.Plugin, fs *protogen.File) error {
	for _, generator := range p.generators {
		if _, err := generator.Generate(plugin, fs); err != nil {

			return err
		}
	}
	return nil
}

func (p *ProtocPlugin) GetProtoGenPlugin(req *pluginpb.CodeGeneratorRequest) (*protogen.Plugin, error) {
	var opts = new(protogen.Options)
	return opts.New(req)
}

func (p *ProtocPlugin) GetProtocRequest() (req *pluginpb.CodeGeneratorRequest, err error) {
	if p.input == nil {
		p.input = os.Stdin
	}
	data, err := ioutil.ReadAll(p.input)
	if err != nil {
		return nil, err
	}
	req = new(pluginpb.CodeGeneratorRequest)
	if err = proto.Unmarshal(data, req); err != nil {
		return nil, err
	}
	return req, nil
}
