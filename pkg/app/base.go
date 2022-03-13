package app

import (
		"google.golang.org/protobuf/compiler/protogen"
		"google.golang.org/protobuf/types/pluginpb"
		"io"
		"io/fs"
		"os"
		"sync"
)

type ProtocPlugin struct {
	input      fs.File
	output     io.Writer
	locker     sync.Locker
	request    *pluginpb.CodeGeneratorRequest
	generators map[string]Generator
}

var (
		name   string
		version string
		nameConstructor =sync.Once{}
		versionConstructor = sync.Once{}
)

type Generator interface {
	Name() string
	Generate(*protogen.Plugin, *protogen.File) (*protogen.GeneratedFile, error)
}

type Option func(*ProtocPlugin)

func SetVersion(v string)  {
		if v == "" {
				return
		}
		nameConstructor.Do(func() {
				version = v
		})
}

func SetName(n string)  {
		if n == "" {
				return
		}
		versionConstructor.Do(func() {
				name = n
		})
}

func GetName() string {
		return name
}

func GetVersion() string  {
		return version
}

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
	var err error
	defer func() {
			if e:=recover();e!=nil {
					switch e.(type) {
					case error:
							err = e.(error)
					}
			}
	}()
	(protogen.Options{}).Run(func(plugin *protogen.Plugin) error {
			return  p.MakeFiles(plugin)
	})
	return err
}

func (p *ProtocPlugin) GetOutput() io.Writer {
	if p.output == nil {
		return os.Stdout
	}
	return p.output
}

func (p *ProtocPlugin) MakeFiles(plugin *protogen.Plugin) error {

	for _, fd := range plugin.Files {
		if !fd.Generate {
			continue
		}
		// @TODO 多协程
		if err := p.generate(plugin, fd); err != nil {
			return err
		}
	}
	return nil
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
