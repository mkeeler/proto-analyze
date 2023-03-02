package fieldusage

import (
	"io/ioutil"
	"sort"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type CollectorConfig struct {
	RemoveListIndexing bool
}

type Collector struct {
	config CollectorConfig
	tree   map[string]*pathNode
}

// WalkFunc will be invoked with a single path element
type WalkFunc func(pathElement string)

type pathNode struct {
	name     string
	children map[string]*pathNode
}

func NewCollector(config CollectorConfig) *Collector {
	return &Collector{
		config: config,
		tree:   make(map[string]*pathNode),
	}
}

func (c *Collector) Walk(push WalkFunc, pop WalkFunc) {
	walkRecurse(c.tree, push, pop)
}

func walkRecurse(values map[string]*pathNode, push WalkFunc, pop WalkFunc) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}

	sort.Sort(sort.StringSlice(keys))

	for _, k := range keys {
		push(k)
		walkRecurse(values[k].children, push, pop)
		pop(k)
	}
}

func (c *Collector) CollectUsageFromJSONFile(topLevelType protoreflect.FullName, file string, types *protoregistry.Types) error {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return c.CollectUsageFromJSON(topLevelType, string(raw), types)
}

func (c *Collector) CollectUsageFromProtoFile(topLevelType protoreflect.FullName, file string, types *protoregistry.Types) error {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return c.CollectUsageFromProtoBytes(topLevelType, raw, types)
}

func (c *Collector) CollectUsageFromProtoBytes(topLevelType protoreflect.FullName, data []byte, types *protoregistry.Types) error {
	mtype, err := types.FindMessageByName(topLevelType)
	if err != nil {
		return err
	}

	msg := mtype.New()

	err = proto.UnmarshalOptions{Resolver: types}.Unmarshal(data, msg.Interface())
	if err != nil {
		return err
	}

	return c.CollectUsageFromProtoMessage(msg, types)
}

func (c *Collector) CollectUsageFromJSON(topLevelType protoreflect.FullName, data string, types *protoregistry.Types) error {
	mtype, err := types.FindMessageByName(topLevelType)
	if err != nil {
		return err
	}

	msg := mtype.New()

	err = protojson.UnmarshalOptions{Resolver: types}.Unmarshal([]byte(data), msg.Interface())
	if err != nil {
		return err
	}

	return c.CollectUsageFromProtoMessage(msg, types)
}

func (c *Collector) CollectUsageFromProtoMessage(msg protoreflect.Message, types *protoregistry.Types) error {
	stack := pathStack{collector: c}

	return protorange.Options{Resolver: types}.Range(msg, stack.push, stack.pop)
}

type pathStack struct {
	stack     []string
	collector *Collector
}

func (s *pathStack) push(p protopath.Values) error {
	current := p.Path[len(p.Path)-1]
	// elide intermediate list index elements if desired
	if s.collector.config.RemoveListIndexing && current.Kind() == protopath.ListIndexStep {
		return nil
	}

	// add the latest node to the stack
	s.stack = append(s.stack, current.String())

	// ensure there is a path down to the stack
	s.ensurePath(s.stack)

	return nil
}

func (s *pathStack) ensurePath(path []string) {
	fields := s.collector.tree

	for _, pathElem := range path {
		sub, ok := fields[pathElem]
		if !ok {
			sub = &pathNode{name: pathElem, children: make(map[string]*pathNode)}
			fields[pathElem] = sub
		}

		fields = sub.children
	}
}

func (s *pathStack) pop(p protopath.Values) error {
	// elide intermediate list index elements if desired
	if s.collector.config.RemoveListIndexing && p.Path[len(p.Path)-1].Kind() == protopath.ListIndexStep {
		return nil
	}

	// pop the last item off the stack
	s.stack = s.stack[:len(s.stack)-1]
	return nil
}

func pathWithoutListIndexes(path protopath.Path) protopath.Path {
	var newPath protopath.Path
	for _, step := range path {
		if step.Kind() != protopath.ListIndexStep {
			newPath = append(newPath, step)
		}
	}
	return newPath
}
