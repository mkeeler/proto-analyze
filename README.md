# proto-analyze
Some protobuf analysis tools

## usage

`proto-field-usage -format json -message envoy.service.discovery.v3.DiscoveryResponse -protoset /Users/mkeeler/envoy.protoset clusters listeners routes secrets endpoints builtin_extension`

* `-message` is the fully qualified protobuf message name of top level files to be processed
* `-protoset` is a file containing a protobuf FileDescriptorSet. The types within the set should include the main message type and all other types nested within it.
* `-format` specifies the format of the files to be processed. This can be either `proto` or `json. `proto` will unmarshal the files assuming they are in binary protobuf wire format whereas `json` will unmarshal assuming they are in the canonical json protobuf representation supported by google.golang.org/protobuf/encoding/protojson.

The positional argument list contains filesystem paths where the data should be loaded from

## example output

Note that this represents the fields that Consul will use for configuring an Envoy L4 network RBAC filter

```
(envoy.config.listener.v3.Filter)
   .name
   .typed_config
      .(envoy.extensions.filters.network.rbac.v3.RBAC)
         .rules
            .action
            .policies
               ["consul-intentions-layer4"]
                  .permissions
                     .any
                  .principals
                     .and_ids
                        .ids
                           .authenticated
                              .principal_name
                                 .safe_regex
                                    .google_re2
                                    .regex
                           .not_id
                              .authenticated
                                 .principal_name
                                    .safe_regex
                                       .google_re2
                                       .regex
                     .authenticated
                        .principal_name
                           .safe_regex
                              .google_re2
                              .regex
         .stat_prefix
```
