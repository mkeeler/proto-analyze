package dynamicregistry

import (
	"io/ioutil"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func CreateRegistryFromFile(file string) (*protoregistry.Types, error) {
	var typeRegistry protoregistry.Types

	err := AddToRegistryFromFile(file, &typeRegistry)
	if err != nil {
		return nil, err
	}
	return &typeRegistry, nil
}

func AddToRegistryFromFile(file string, registry *protoregistry.Types) error {
	files, err := LoadProtoFileDescriptorSet(file)
	if err != nil {
		return err
	}

	return AddProtoFilesToRegistry(files, registry)
}

func LoadProtoFileDescriptorSet(file string) (*protoregistry.Files, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	descriptorSet := descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(raw, &descriptorSet)
	if err != nil {
		return nil, err
	}

	files, err := protodesc.NewFiles(&descriptorSet)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func AddProtoFilesToRegistry(files *protoregistry.Files, registry *protoregistry.Types) error {
	var err error
	files.RangeFiles(func(desc protoreflect.FileDescriptor) bool {
		enums := desc.Enums()
		for i := 0; i < enums.Len(); i++ {
			err = registry.RegisterEnum(dynamicpb.NewEnumType(enums.Get(i)))
			if err != nil {
				return false
			}
		}

		extensions := desc.Extensions()
		for i := 0; i < extensions.Len(); i++ {
			err = registry.RegisterExtension(dynamicpb.NewExtensionType(extensions.Get(i)))
			if err != nil {
				return false
			}

		}

		messages := desc.Messages()
		for i := 0; i < messages.Len(); i++ {
			err = registry.RegisterMessage(dynamicpb.NewMessageType(messages.Get(i)))
			if err != nil {
				return false
			}
		}
		return true
	})

	return err
}
